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

// writeDynamicChildrenSync writes a sync assignment: varName.Children = func() []*layout.Input{...}().
// Used by the build-once pattern to rebuild dynamic children in the sync function.
func writeDynamicChildrenSync(buf *bytes.Buffer, varName string, children []template.Node, stylesheet *style.Stylesheet, tracker *instanceTracker) {
	fmt.Fprintf(buf, "\t\t%s.Children = func() []*layout.Input {\n", varName)
	fmt.Fprintf(buf, "\t\t\tvar cs []*layout.Input\n")
	for _, child := range children {
		writeDynamicChild(buf, child, stylesheet, 3, "\t\t\t", tracker)
	}
	fmt.Fprintf(buf, "\t\t\treturn cs\n")
	fmt.Fprintf(buf, "\t\t}()\n")
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

