package codegen

import (
	"bytes"
	"fmt"

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

// writeDynamicChildren emits an IIFE that builds the children slice dynamically.
func writeDynamicChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	fmt.Fprintf(buf, "%s\tChildren: func() []*layout.Input {\n", tabs)
	fmt.Fprintf(buf, "%s\t\tvar cs []*layout.Input\n", tabs)
	for _, child := range children {
		writeDynamicChild(buf, child, stylesheet, indent+2, tabs+"\t\t", tracker)
	}
	fmt.Fprintf(buf, "%s\t\treturn cs\n", tabs)
	fmt.Fprintf(buf, "%s\t}(),\n", tabs)
}

// writeDynamicChild dispatches a single child inside an IIFE.
func writeDynamicChild(buf *bytes.Buffer, child template.Node, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	switch n := child.(type) {
	case *template.IfNode:
		writeIfBlock(buf, n, stylesheet, indent, tabs, tracker)
	case *template.ForNode:
		writeForBlock(buf, n, stylesheet, indent, tabs, tracker)
	default:
		writeAppendNode(buf, child, stylesheet, indent, tabs, tracker)
	}
}

// writeAppendNode writes cs = append(cs, ...) for a regular node.
func writeAppendNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	switch n := node.(type) {
	case *template.TextElement:
		writeAppendText(buf, n, stylesheet, tabs)
	case *template.BoxElement:
		writeAppendBox(buf, n, stylesheet, indent, tabs, tracker)
	case *template.ComponentElement:
		writeAppendComponent(buf, tabs, tracker)
	}
}

// writeAppendText writes cs = append(cs, &layout.Input{Kind: KindText, ...}).
func writeAppendText(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, tabs string) {
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)
	fmt.Fprintf(buf, "%scs = append(cs, &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, contentExpr(n.Parts))
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s})\n", tabs)
}

// writeAppendBox writes cs = append(cs, &layout.Input{Kind: KindBox, ...}).
func writeAppendBox(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	props := resolveProps(stylesheet, "box", n.Attributes)
	fmt.Fprintf(buf, "%scs = append(cs, &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, props)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	if len(n.Children) > 0 {
		if hasDynamicChildren(n.Children) {
			writeDynamicChildren(buf, n.Children, stylesheet, indent, tabs, tracker)
		} else {
			writeBoxChildren(buf, n.Children, stylesheet, indent, tabs, tracker, nil)
		}
	}
	fmt.Fprintf(buf, "%s})\n", tabs)
}

// writeAppendComponent writes cs = append(cs, comp.Layout()).
func writeAppendComponent(buf *bytes.Buffer, tabs string, tracker *instanceTracker) {
	varName := tracker.next()
	fmt.Fprintf(buf, "%scs = append(cs, %s.Layout())\n", tabs, varName)
}

// writeIfBlock emits an if/else block inside an IIFE.
func writeIfBlock(buf *bytes.Buffer, n *template.IfNode, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	fmt.Fprintf(buf, "%sif %s {\n", tabs, n.Condition)
	for _, child := range n.Then {
		writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", tracker)
	}
	if n.Else != nil {
		fmt.Fprintf(buf, "%s} else {\n", tabs)
		for _, child := range n.Else {
			writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", tracker)
		}
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeForBlock emits a for loop inside an IIFE.
func writeForBlock(buf *bytes.Buffer, n *template.ForNode, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	fmt.Fprintf(buf, "%sfor %s {\n", tabs, n.Clause)
	for _, child := range n.Children {
		writeDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", tracker)
	}
	if n.Key != "" {
		fmt.Fprintf(buf, "%s\tcs[len(cs)-1].Key = fmt.Sprint(%s)\n", tabs, n.Key)
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// --- Component mode (receiver-prefixed) variants ---

// writeComponentDynamicChildren emits an IIFE with receiver-prefixed conditions/clauses.
func writeComponentDynamicChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	fmt.Fprintf(buf, "%s\tChildren: func() []*layout.Input {\n", tabs)
	fmt.Fprintf(buf, "%s\t\tvar cs []*layout.Input\n", tabs)
	for _, child := range children {
		writeComponentDynamicChild(buf, child, stylesheet, indent+2, tabs+"\t\t", varNames)
	}
	fmt.Fprintf(buf, "%s\t\treturn cs\n", tabs)
	fmt.Fprintf(buf, "%s\t}(),\n", tabs)
}

// writeComponentDynamicChild dispatches a single child in component mode.
func writeComponentDynamicChild(buf *bytes.Buffer, child template.Node, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	switch n := child.(type) {
	case *template.IfNode:
		writeComponentIfBlock(buf, n, stylesheet, indent, tabs, varNames)
	case *template.ForNode:
		writeComponentForBlock(buf, n, stylesheet, indent, tabs, varNames)
	default:
		writeComponentAppendNode(buf, child, stylesheet, indent, tabs, varNames)
	}
}

// writeComponentAppendNode writes cs = append(cs, ...) for a regular node in component mode.
func writeComponentAppendNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	switch n := node.(type) {
	case *template.TextElement:
		writeComponentAppendText(buf, n, stylesheet, tabs, varNames)
	case *template.BoxElement:
		writeComponentAppendBox(buf, n, stylesheet, indent, tabs, varNames)
	}
}

// writeComponentAppendText writes cs = append(cs, &layout.Input{Kind: KindText, ...}) with receiver prefix.
func writeComponentAppendText(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, tabs string, varNames map[string]bool) {
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)
	fmt.Fprintf(buf, "%scs = append(cs, &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, componentContentExpr(n.Parts, varNames))
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s})\n", tabs)
}

// writeComponentAppendBox writes cs = append(cs, &layout.Input{Kind: KindBox, ...}) with receiver prefix.
func writeComponentAppendBox(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	props := resolveProps(stylesheet, "box", n.Attributes)
	fmt.Fprintf(buf, "%scs = append(cs, &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, props)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	if len(n.Children) > 0 {
		if hasDynamicChildren(n.Children) {
			writeComponentDynamicChildren(buf, n.Children, stylesheet, indent, tabs, varNames)
		} else {
			writeComponentBoxChildren(buf, n.Children, stylesheet, indent, tabs, varNames)
		}
	}
	fmt.Fprintf(buf, "%s})\n", tabs)
}

// writeComponentIfBlock emits an if/else block with receiver-prefixed condition.
func writeComponentIfBlock(buf *bytes.Buffer, n *template.IfNode, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	fmt.Fprintf(buf, "%sif %s {\n", tabs, prefixConditionExpr(n.Condition, varNames))
	for _, child := range n.Then {
		writeComponentDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", varNames)
	}
	if n.Else != nil {
		fmt.Fprintf(buf, "%s} else {\n", tabs)
		for _, child := range n.Else {
			writeComponentDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", varNames)
		}
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeComponentForBlock emits a for loop with receiver-prefixed clause.
func writeComponentForBlock(buf *bytes.Buffer, n *template.ForNode, stylesheet *style.Stylesheet, indent int, tabs string, varNames map[string]bool) {
	fmt.Fprintf(buf, "%sfor %s {\n", tabs, prefixConditionExpr(n.Clause, varNames))
	for _, child := range n.Children {
		writeComponentDynamicChild(buf, child, stylesheet, indent+1, tabs+"\t", varNames)
	}
	if n.Key != "" {
		fmt.Fprintf(buf, "%s\tcs[len(cs)-1].Key = fmt.Sprint(%s)\n", tabs, prefixConditionExpr(n.Key, varNames))
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}
