package codegen

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

// writeLayoutTree writes the layout.Input tree construction code.
// When inClosure is true, adds an extra tab of indentation.
// When ext is non-nil, expression text nodes are extracted as named variables.
func writeLayoutTree(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, inClosure bool, instances []componentInstance, ext *extractionCtx) {
	baseIndent := 1
	if inClosure {
		baseIndent = 2
	}
	tabs := indentStr(baseIndent)
	tracker := newInstanceTracker(instances)
	rootProps := resolveProps(stylesheet, "root", nil)

	fmt.Fprintf(buf, "%sroot := &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	rootAttrs := map[string]string{"direction": "column"}
	writeBoxAttributes(buf, tabs, rootAttrs, rootProps)
	if rootProps != nil {
		writeStyleLiteral(buf, tabs, rootProps)
	}
	if hasDynamicChildren(doc.Children) {
		if ext != nil {
			// Build-once: root.Children rebuilt in sync
			fmt.Fprintf(buf, "%s}\n", tabs)
			writeDynamicChildrenSync(&ext.syncBuf, "root", doc.Children, stylesheet, tracker)
			return
		}
		writeDynamicChildren(buf, doc.Children, stylesheet, baseIndent, tabs, tracker)
	} else {
		fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
		for _, child := range doc.Children {
			writeInputNode(buf, child, stylesheet, baseIndent+2, tracker, ext)
		}
		fmt.Fprintf(buf, "%s\t},\n", tabs)
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeInputNode writes a layout.Input literal for a template AST node.
func writeInputNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, tracker *instanceTracker, ext *extractionCtx) {
	switch n := node.(type) {
	case *template.TextElement:
		writeTextInput(buf, n, stylesheet, indent, ext)
	case *template.BoxElement:
		writeBoxInput(buf, n, stylesheet, indent, tracker, ext)
	case *template.ComponentElement:
		writeComponentElement(buf, indent, tracker, ext)
	}
}

// writeComponentElement handles a component reference — either inlines or falls back.
func writeComponentElement(buf *bytes.Buffer, indent int, tracker *instanceTracker, ext *extractionCtx) {
	inst := tracker.nextInstance()
	if inst != nil && inst.Info.Doc != nil {
		writeInlinedComponent(buf, inst, indent, tracker, ext)
	} else if inst != nil {
		writeComponentRefByName(buf, indent, inst.VarName)
	}
}

// writeTextInput writes a layout.Input literal for a text element.
// When ext is non-nil and the text has expressions, the node is extracted
// as a named variable for sync patching.
func writeTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int, ext *extractionCtx) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)

	if ext != nil && hasExprParts(n.Parts) {
		writeExtractedTextNode(buf, &ext.declBuf, n, props, tabs, ext)
		return
	}

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, contentExpr(n.Parts))
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeExtractedTextNode extracts an expression text node as a named variable.
// The declaration goes to declBuf; a variable reference goes to the tree buf.
func writeExtractedTextNode(treeBuf, declBuf *bytes.Buffer, n *template.TextElement, props map[string]string, tabs string, ext *extractionCtx) {
	name := ext.nextNodeName()
	expr := contentExpr(n.Parts)

	// Write declaration to declBuf (at function scope indent)
	fmt.Fprintf(declBuf, "\t%s := &layout.Input{\n", name)
	fmt.Fprintf(declBuf, "\t\tKind:    layout.KindText,\n")
	fmt.Fprintf(declBuf, "\t\tContent: %s,\n", expr)
	if props != nil {
		writeStyleLiteral(declBuf, "\t", props)
	}
	fmt.Fprintf(declBuf, "\t}\n")

	// Record sync entry
	ext.nodes = append(ext.nodes, extractedNode{
		varName:  name,
		syncExpr: expr,
	})

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// hasExprParts returns true if any part is an ExprPart.
func hasExprParts(parts []template.Part) bool {
	for _, p := range parts {
		if _, ok := p.(*template.ExprPart); ok {
			return true
		}
	}
	return false
}

// writeBoxInput writes a layout.Input literal for a box element.
// When ext is non-nil and the box has dynamic children, the box is extracted
// as a named variable and its Children are rebuilt in sync.
func writeBoxInput(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, tracker *instanceTracker, ext *extractionCtx) {
	if ext != nil && hasDynamicChildren(n.Children) {
		writeExtractedDynamicBox(buf, n, stylesheet, indent, tracker, ext)
		return
	}
	tabs := indentStr(indent)
	props := resolveProps(stylesheet, "box", n.Attributes)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, props)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	writeBoxChildren(buf, n.Children, stylesheet, indent, tabs, tracker, ext)
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeExtractedDynamicBox extracts a box with dynamic children as a named variable.
// The box declaration (without Children) goes to declBuf; Children are rebuilt in syncBuf.
func writeExtractedDynamicBox(treeBuf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, tracker *instanceTracker, ext *extractionCtx) {
	tabs := indentStr(indent)
	name := ext.nextBoxName()
	props := resolveProps(stylesheet, "box", n.Attributes)

	// Write declaration to declBuf (at function scope, no Children)
	fmt.Fprintf(&ext.declBuf, "\t%s := &layout.Input{\n", name)
	fmt.Fprintf(&ext.declBuf, "\t\tKind: layout.KindBox,\n")
	writeBoxAttributes(&ext.declBuf, "\t", n.Attributes, props)
	if props != nil {
		writeStyleLiteral(&ext.declBuf, "\t", props)
	}
	fmt.Fprintf(&ext.declBuf, "\t}\n")

	// Write dynamic children rebuild to syncBuf
	writeDynamicChildrenSync(&ext.syncBuf, name, n.Children, stylesheet, tracker)

	// Write reference in tree
	fmt.Fprintf(treeBuf, "%s%s,\n", tabs, name)
}

// writeComponentRef writes a component Layout() call as a layout tree entry.
func writeComponentRef(buf *bytes.Buffer, indent int, tracker *instanceTracker) {
	tabs := indentStr(indent)
	varName := tracker.next()
	fmt.Fprintf(buf, "%s%s.Layout(),\n", tabs, varName)
}

// writeBoxAttributes writes direction, width, height, padding, and border fields.
func writeBoxAttributes(buf *bytes.Buffer, tabs string, attrs map[string]string, props map[string]string) {
	if dir, ok := mergedAttr(attrs, props, "direction"); ok {
		fmt.Fprintf(buf, "%s\tDirection: %q,\n", tabs, dir)
	}
	writeIntAttr(buf, tabs, attrs, props, "width", "FixedWidth")
	writeIntAttr(buf, tabs, attrs, props, "height", "FixedHeight")
	writeIntAttr(buf, tabs, attrs, props, "gap", "Gap")
	writeIntAttr(buf, tabs, attrs, props, "flex-grow", "FlexGrow")
	if j, ok := mergedAttr(attrs, props, "justify"); ok {
		fmt.Fprintf(buf, "%s\tJustify: %q,\n", tabs, j)
	}
	if a, ok := mergedAttr(attrs, props, "align"); ok {
		fmt.Fprintf(buf, "%s\tAlign: %q,\n", tabs, a)
	}
	if p, ok := mergedAttr(attrs, props, "padding"); ok {
		fmt.Fprintf(buf, "%s\tPadding: layout.ParsePadding(%q),\n", tabs, p)
	}
	if b, ok := mergedAttr(attrs, props, "border"); ok {
		fmt.Fprintf(buf, "%s\tBorder: %q,\n", tabs, b)
	}
	if bt, ok := mergedAttr(attrs, props, "border-title"); ok {
		fmt.Fprintf(buf, "%s\tBorderTitle: %q,\n", tabs, bt)
	}
	if bc, ok := mergedAttr(attrs, props, "border-collapse"); ok && bc == "collapse" {
		fmt.Fprintf(buf, "%s\tBorderCollapse: true,\n", tabs)
	}
	if o, ok := mergedAttr(attrs, props, "overflow"); ok {
		fmt.Fprintf(buf, "%s\tOverflow: %q,\n", tabs, o)
	}
	writeIntAttr(buf, tabs, attrs, props, "min-width", "MinWidth")
	if d, ok := mergedAttr(attrs, props, "display"); ok && d == "none" {
		fmt.Fprintf(buf, "%s\tDisplay: %q,\n", tabs, d)
	}
	if p, ok := mergedAttr(attrs, props, "position"); ok {
		fmt.Fprintf(buf, "%s\tPosition: %q,\n", tabs, p)
	}
	writeIntAttr(buf, tabs, attrs, props, "top", "Top")
	writeIntAttr(buf, tabs, attrs, props, "left", "Left")
	writeIntAttr(buf, tabs, attrs, props, "right", "Right")
	writeIntAttr(buf, tabs, attrs, props, "bottom", "Bottom")
	writeIntAttr(buf, tabs, attrs, props, "z-index", "ZIndex")
}

// writeIntAttr writes an integer attribute field if present and parseable.
func writeIntAttr(buf *bytes.Buffer, tabs string, attrs, props map[string]string, attrKey, fieldName string) {
	val, ok := mergedAttr(attrs, props, attrKey)
	if !ok {
		return
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		return
	}
	fmt.Fprintf(buf, "%s\t%s: %d,\n", tabs, fieldName, v)
}

// writeBoxChildren writes the Children field of a box input if there are children.
func writeBoxChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker, ext *extractionCtx) {
	if len(children) == 0 {
		return
	}
	if hasDynamicChildren(children) {
		writeDynamicChildren(buf, children, stylesheet, indent, tabs, tracker)
		return
	}
	fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
	for _, child := range children {
		writeInputNode(buf, child, stylesheet, indent+2, tracker, ext)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// indentStr returns n tab characters.
func indentStr(n int) string {
	return strings.Repeat("\t", n)
}
