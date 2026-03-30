package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// hasDynamicChildren checks if any immediate child is an IfNode or ForNode.
func hasDynamicChildren(children []template.Node) bool {
	for _, child := range children {
		switch child.(type) {
		case *template.IfNode, *template.ForNode:
			return true
		}
	}
	return false
}

// writeDynamicChildrenSync writes a sync assignment: varName.Children = func() []*sumi.Input{...}().
func writeDynamicChildrenSync(buf *bytes.Buffer, varName string, children []template.Node, stylesheet *style.Stylesheet, signals map[string]bool) {
	fmt.Fprintf(buf, "\t\t%s.Children = func() []*sumi.Input {\n", varName)
	fmt.Fprintf(buf, "\t\t\tvar cs []*sumi.Input\n")
	for _, child := range children {
		writeDynamicChild(buf, child, stylesheet, 3, "\t\t\t", signals)
	}
	fmt.Fprintf(buf, "\t\t\treturn cs\n")
	fmt.Fprintf(buf, "\t\t}()\n")
}

// writeDynamicChildren emits an IIFE that builds the children slice dynamically.
func writeDynamicChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, signals map[string]bool) {
	fmt.Fprintf(buf, "%s\tChildren: func() []*sumi.Input {\n", tabs)
	fmt.Fprintf(buf, "%s\t\tvar cs []*sumi.Input\n", tabs)
	for _, child := range children {
		writeDynamicChild(buf, child, stylesheet, indent+2, tabs+"\t\t", signals)
	}
	fmt.Fprintf(buf, "%s\t\treturn cs\n", tabs)
	fmt.Fprintf(buf, "%s\t}(),\n", tabs)
}

// writeDynamicChild dispatches a single child inside an IIFE.
func writeDynamicChild(buf *bytes.Buffer, child template.Node, stylesheet *style.Stylesheet, indent int, tabs string, signals map[string]bool) {
	switch n := child.(type) {
	case *template.IfNode:
		writeIfBlock(buf, n, stylesheet, indent, tabs, signals)
	case *template.ForNode:
		writeForBlock(buf, n, stylesheet, indent, tabs, signals)
	default:
		writeAppendNode(buf, child, stylesheet, indent, tabs, signals)
	}
}

// writeAppendNode emits cs = append(cs, &sumi.Input{...}) for a static child.
func writeAppendNode(buf *bytes.Buffer, child template.Node, stylesheet *style.Stylesheet, indent int, tabs string, signals map[string]bool) {
	var ext *extractionCtx
	if len(signals) > 0 {
		ext = &extractionCtx{signals: signals, inDynamic: true}
	}
	var nodeBuf bytes.Buffer
	writeInputNode(&nodeBuf, child, stylesheet, 0, ext)
	nodeStr := trimTrailingComma(nodeBuf.String())
	// writeInputNode emits "{...}" for slice literal context.
	// For append, we need "&sumi.Input{...}".
	if strings.HasPrefix(strings.TrimSpace(nodeStr), "{") {
		nodeStr = "&sumi.Input" + strings.TrimSpace(nodeStr)
	}
	fmt.Fprintf(buf, "%scs = append(cs, %s)\n", tabs, nodeStr)
}

// writeAppendTextWithSignals emits a text node with signal auto-unwrapping for dynamic children.
func writeAppendTextWithSignals(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, signals map[string]bool) {
	content := contentExprSignals(n.Parts, signals)
	props := resolveProps(stylesheet, "text", n.Attributes)
	buf.WriteString("&sumi.Input{\n")
	buf.WriteString("\t\tKind: sumi.KindText,\n")
	fmt.Fprintf(buf, "\t\tContent: %s,\n", content)
	if props != nil {
		writeStyleLiteral(buf, "\t", props)
	}
	buf.WriteString("\t}")
}

func trimTrailingComma(s string) string {
	s = strings.TrimRight(s, " \t\n")
	if len(s) > 0 && s[len(s)-1] == ',' {
		s = s[:len(s)-1]
	}
	return s
}

// writeIfBlock emits an if/else block inside an IIFE.
func writeIfBlock(buf *bytes.Buffer, n *template.IfNode, stylesheet *style.Stylesheet, indent int, tabs string, signals map[string]bool) {
	// Don't unwrap signals in conditions — the user writes .Get() explicitly in .sumi files.
	fmt.Fprintf(buf, "%sif %s {\n", tabs, n.Condition)
	for _, child := range n.Then {
		writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", signals)
	}
	if n.Else != nil {
		fmt.Fprintf(buf, "%s} else {\n", tabs)
		for _, child := range n.Else {
			writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", signals)
		}
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeForBlock emits a for loop inside an IIFE.
func writeForBlock(buf *bytes.Buffer, n *template.ForNode, stylesheet *style.Stylesheet, indent int, tabs string, signals map[string]bool) {
	// Don't unwrap signals in for clauses — the user writes .Get() explicitly in .sumi files.
	fmt.Fprintf(buf, "%sfor %s {\n", tabs, n.Clause)
	for _, child := range n.Children {
		writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", signals)
	}
	if n.Key != "" {
		fmt.Fprintf(buf, "%s\tcs[len(cs)-1].Key = sumi.Sprint(%s)\n", tabs, n.Key)
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}
