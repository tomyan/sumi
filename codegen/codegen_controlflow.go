package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// activeSignals holds signal names during code generation for auto-unwrapping.
var activeSignals map[string]bool

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

// writeDynamicChildrenSync writes a sync assignment: varName.Children = func() []*layout.Input{...}().
func writeDynamicChildrenSync(buf *bytes.Buffer, varName string, children []template.Node, stylesheet *style.Stylesheet) {
	fmt.Fprintf(buf, "\t\t%s.Children = func() []*layout.Input {\n", varName)
	fmt.Fprintf(buf, "\t\t\tvar cs []*layout.Input\n")
	for _, child := range children {
		writeDynamicChild(buf, child, stylesheet, 3, "\t\t\t")
	}
	fmt.Fprintf(buf, "\t\t\treturn cs\n")
	fmt.Fprintf(buf, "\t\t}()\n")
}

// writeDynamicChildren emits an IIFE that builds the children slice dynamically.
func writeDynamicChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string) {
	fmt.Fprintf(buf, "%s\tChildren: func() []*layout.Input {\n", tabs)
	fmt.Fprintf(buf, "%s\t\tvar cs []*layout.Input\n", tabs)
	for _, child := range children {
		writeDynamicChild(buf, child, stylesheet, indent+2, tabs+"\t\t")
	}
	fmt.Fprintf(buf, "%s\t\treturn cs\n", tabs)
	fmt.Fprintf(buf, "%s\t}(),\n", tabs)
}

// writeDynamicChild dispatches a single child inside an IIFE.
func writeDynamicChild(buf *bytes.Buffer, child template.Node, stylesheet *style.Stylesheet, indent int, tabs string) {
	switch n := child.(type) {
	case *template.IfNode:
		writeIfBlock(buf, n, stylesheet, indent, tabs)
	case *template.ForNode:
		writeForBlock(buf, n, stylesheet, indent, tabs)
	default:
		writeAppendNode(buf, child, stylesheet, indent, tabs)
	}
}

// writeAppendNode emits cs = append(cs, &layout.Input{...}) for a static child.
func writeAppendNode(buf *bytes.Buffer, child template.Node, stylesheet *style.Stylesheet, indent int, tabs string) {
	fmt.Fprintf(buf, "%scs = append(cs, ", tabs)
	var nodeBuf bytes.Buffer
	writeInputNode(&nodeBuf, child, stylesheet, 0, nil, nil)
	// Trim trailing comma and newline from the node literal.
	nodeStr := nodeBuf.String()
	nodeStr = trimTrailingComma(nodeStr)
	buf.WriteString(nodeStr)
	buf.WriteString(")\n")
}

func trimTrailingComma(s string) string {
	s = strings.TrimRight(s, " \t\n")
	if len(s) > 0 && s[len(s)-1] == ',' {
		s = s[:len(s)-1]
	}
	return s
}

// writeIfBlock emits an if/else block inside an IIFE.
func writeIfBlock(buf *bytes.Buffer, n *template.IfNode, stylesheet *style.Stylesheet, indent int, tabs string) {
	cond := unwrapSignals(n.Condition, activeSignals)
	fmt.Fprintf(buf, "%sif %s {\n", tabs, cond)
	for _, child := range n.Then {
		writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t")
	}
	if n.Else != nil {
		fmt.Fprintf(buf, "%s} else {\n", tabs)
		for _, child := range n.Else {
			writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t")
		}
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeForBlock emits a for loop inside an IIFE.
func writeForBlock(buf *bytes.Buffer, n *template.ForNode, stylesheet *style.Stylesheet, indent int, tabs string) {
	clause := unwrapSignals(n.Clause, activeSignals)
	fmt.Fprintf(buf, "%sfor %s {\n", tabs, clause)
	for _, child := range n.Children {
		writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t")
	}
	if n.Key != "" {
		fmt.Fprintf(buf, "%s\tcs[len(cs)-1].Key = fmt.Sprint(%s)\n", tabs, n.Key)
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}
