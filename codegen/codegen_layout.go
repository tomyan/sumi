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
func writeLayoutTree(buf *bytes.Buffer, doc *template.Document, stylesheet *style.Stylesheet, inClosure bool, instances []componentInstance) {
	baseIndent := 1
	if inClosure {
		baseIndent = 2
	}
	tabs := indentStr(baseIndent)
	tracker := newInstanceTracker(instances)

	fmt.Fprintf(buf, "%sroot := &layout.Input{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:      layout.KindBox,\n", tabs)
	fmt.Fprintf(buf, "%s\tDirection: \"column\",\n", tabs)
	fmt.Fprintf(buf, "%s\tChildren:  []*layout.Input{\n", tabs)
	for _, child := range doc.Children {
		writeInputNode(buf, child, stylesheet, baseIndent+2, tracker)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
	fmt.Fprintf(buf, "%s}\n", tabs)
}

// writeInputNode writes a layout.Input literal for a template AST node.
func writeInputNode(buf *bytes.Buffer, node template.Node, stylesheet *style.Stylesheet, indent int, tracker *instanceTracker) {
	switch n := node.(type) {
	case *template.TextElement:
		writeTextInput(buf, n, stylesheet, indent)
	case *template.BoxElement:
		writeBoxInput(buf, n, stylesheet, indent, tracker)
	case *template.ComponentElement:
		writeComponentRef(buf, indent, tracker)
	}
}

// writeTextInput writes a layout.Input literal for a text element.
func writeTextInput(buf *bytes.Buffer, n *template.TextElement, stylesheet *style.Stylesheet, indent int) {
	tabs := indentStr(indent)
	attrs := n.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	props := resolveProps(stylesheet, "text", attrs)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind:    layout.KindText,\n", tabs)
	fmt.Fprintf(buf, "%s\tContent: %s,\n", tabs, contentExpr(n.Parts))
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	fmt.Fprintf(buf, "%s},\n", tabs)
}

// writeBoxInput writes a layout.Input literal for a box element.
func writeBoxInput(buf *bytes.Buffer, n *template.BoxElement, stylesheet *style.Stylesheet, indent int, tracker *instanceTracker) {
	tabs := indentStr(indent)
	props := resolveProps(stylesheet, "box", n.Attributes)

	fmt.Fprintf(buf, "%s{\n", tabs)
	fmt.Fprintf(buf, "%s\tKind: layout.KindBox,\n", tabs)
	writeBoxAttributes(buf, tabs, n.Attributes, props)
	if props != nil {
		writeStyleLiteral(buf, tabs, props)
	}
	writeBoxChildren(buf, n.Children, stylesheet, indent, tabs, tracker)
	fmt.Fprintf(buf, "%s},\n", tabs)
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
	if o, ok := mergedAttr(attrs, props, "overflow"); ok {
		fmt.Fprintf(buf, "%s\tOverflow: %q,\n", tabs, o)
	}
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
func writeBoxChildren(buf *bytes.Buffer, children []template.Node, stylesheet *style.Stylesheet, indent int, tabs string, tracker *instanceTracker) {
	if len(children) == 0 {
		return
	}
	fmt.Fprintf(buf, "%s\tChildren: []*layout.Input{\n", tabs)
	for _, child := range children {
		writeInputNode(buf, child, stylesheet, indent+2, tracker)
	}
	fmt.Fprintf(buf, "%s\t},\n", tabs)
}

// writeRenderTreeFunc generates the renderTree helper function.
// Handles clipping, single-line text (Content), and wrapped text (Lines).
func writeRenderTreeFunc(buf *bytes.Buffer) {
	buf.WriteString("func renderTree(buf *render.Buffer, box *layout.Box, clip *render.Clip) {\n")
	buf.WriteString("\tif box.Border != \"\" && box.Border != \"none\" {\n")
	buf.WriteString("\t\tbuf.DrawStyledBorder(box.Y, box.X, box.Width, box.Height, box.Border, box.Style)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif box.Lines != nil {\n")
	buf.WriteString("\t\tfor i, line := range box.Lines {\n")
	buf.WriteString("\t\t\tbuf.WriteStyledTextClipped(box.Y+i, box.X, line, box.Style, clip)\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t} else if box.Content != \"\" {\n")
	buf.WriteString("\t\tbuf.WriteStyledTextClipped(box.Y, box.X, box.Content, box.Style, clip)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tchildClip := clip\n")
	buf.WriteString("\tif box.Clip != nil {\n")
	buf.WriteString("\t\tif clip != nil {\n")
	buf.WriteString("\t\t\tchildClip = clip.Intersect(box.Clip)\n")
	buf.WriteString("\t\t} else {\n")
	buf.WriteString("\t\t\tchildClip = box.Clip\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tfor _, child := range box.Children {\n")
	buf.WriteString("\t\trenderTree(buf, child, childClip)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

// indentStr returns n tab characters.
func indentStr(n int) string {
	return strings.Repeat("\t", n)
}
