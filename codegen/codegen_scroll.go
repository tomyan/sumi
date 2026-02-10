package codegen

import (
	"bytes"
	"fmt"

	"github.com/tomyan/sumi/parser/template"
)

// scrollableBox tracks a scrollable box element and its index in the tree.
type scrollableBox struct {
	Index    int    // position in document (for naming: scroll0, scroll1, ...)
	TreePath string // path expression to reach this box in the layout tree (e.g. "tree.Children[0]")
}

// findScrollableBoxes finds all boxes with overflow="scroll" or overflow="auto" in the document.
func findScrollableBoxes(doc *template.Document) []scrollableBox {
	var boxes []scrollableBox
	for i, child := range doc.Children {
		findScrollableInNode(child, fmt.Sprintf("tree.Children[%d]", i), &boxes)
	}
	return boxes
}

// findScrollableInNode recursively searches for scrollable boxes.
func findScrollableInNode(node template.Node, path string, boxes *[]scrollableBox) {
	box, ok := node.(*template.BoxElement)
	if !ok {
		return
	}
	overflow := box.Attributes["overflow"]
	if overflow == "scroll" || overflow == "auto" {
		*boxes = append(*boxes, scrollableBox{
			Index:    len(*boxes),
			TreePath: path,
		})
	}
	for i, child := range box.Children {
		findScrollableInNode(child, fmt.Sprintf("%s.Children[%d]", path, i), boxes)
	}
}

// scrollVarName returns the variable name for a scroll state at the given index.
func scrollVarName(index int) string {
	return fmt.Sprintf("scroll%d", index)
}

// writeScrollStateDecls writes var declarations for each scrollable box.
func writeScrollStateDecls(buf *bytes.Buffer, scrollBoxes []scrollableBox) {
	for _, sb := range scrollBoxes {
		fmt.Fprintf(buf, "\tvar %s layout.ScrollState\n", scrollVarName(sb.Index))
	}
	if len(scrollBoxes) > 0 {
		buf.WriteString("\n")
	}
}

// writeScrollTreeWiring writes code that copies scroll state into the layout tree.
func writeScrollTreeWiring(buf *bytes.Buffer, scrollBoxes []scrollableBox) {
	for _, sb := range scrollBoxes {
		name := scrollVarName(sb.Index)
		fmt.Fprintf(buf, "\t\t%s.ScrollY = %s.ScrollY\n", sb.TreePath, name)
		fmt.Fprintf(buf, "\t\t%s.ScrollX = %s.ScrollX\n", sb.TreePath, name)
	}
}

// prevTreePath converts a tree path like "tree.Children[0]" to use prevTree.
func prevTreePath(treePath string) string {
	if len(treePath) > 4 && treePath[:4] == "tree" {
		return "prevTree" + treePath[4:]
	}
	return treePath
}

// writeScrollDispatch writes the EventSpecial handler for scroll keys.
func writeScrollDispatch(buf *bytes.Buffer, scrollBoxes []scrollableBox) {
	if len(scrollBoxes) == 0 {
		return
	}
	buf.WriteString("\t\t\tif evt.Kind == input.EventSpecial && prevTree != nil {\n")
	buf.WriteString("\t\t\t\tswitch evt.Special {\n")
	// For now, all scroll events go to the first scrollable box.
	// Future: focus-based dispatch.
	name := scrollVarName(scrollBoxes[0].Index)
	path := prevTreePath(scrollBoxes[0].TreePath)
	buf.WriteString("\t\t\t\tcase input.KeyDown:\n")
	fmt.Fprintf(buf, "\t\t\t\t\t%s.ScrollDown(%s.ContentHeight, %s.Height)\n", name, path, path)
	buf.WriteString("\t\t\t\t\tdirty = true\n")
	buf.WriteString("\t\t\t\tcase input.KeyUp:\n")
	fmt.Fprintf(buf, "\t\t\t\t\t%s.ScrollUp()\n", name)
	buf.WriteString("\t\t\t\t\tdirty = true\n")
	buf.WriteString("\t\t\t\tcase input.KeyPgDn:\n")
	fmt.Fprintf(buf, "\t\t\t\t\t%s.PageDown(%s.ContentHeight, %s.Height)\n", name, path, path)
	buf.WriteString("\t\t\t\t\tdirty = true\n")
	buf.WriteString("\t\t\t\tcase input.KeyPgUp:\n")
	fmt.Fprintf(buf, "\t\t\t\t\t%s.PageUp(%s.Height)\n", name, path)
	buf.WriteString("\t\t\t\t\tdirty = true\n")
	buf.WriteString("\t\t\t\tcase input.KeyRight:\n")
	fmt.Fprintf(buf, "\t\t\t\t\t%s.ScrollRight(%s.ContentWidth, %s.Width)\n", name, path, path)
	buf.WriteString("\t\t\t\t\tdirty = true\n")
	buf.WriteString("\t\t\t\tcase input.KeyLeft:\n")
	fmt.Fprintf(buf, "\t\t\t\t\t%s.ScrollLeft()\n", name)
	buf.WriteString("\t\t\t\t\tdirty = true\n")
	buf.WriteString("\t\t\t\t}\n")
	buf.WriteString("\t\t\t}\n")
}
