package layout

import "github.com/tomyan/sumi/runtime/render"

// applyColumnCollapse shifts children so adjacent bordered boxes overlap by 1 row,
// and sets CollapsedEdges flags on the shared edges.
func applyColumnCollapse(boxes []*Box, inputs []*Input) {
	for i := 1; i < len(boxes); i++ {
		if hasBorder(inputs[i-1].Border) && hasBorder(inputs[i].Border) {
			boxes[i].Y -= i // shift by number of overlaps so far
			boxes[i-1].Collapsed.Bottom = true
			boxes[i].Collapsed.Top = true
		}
	}
}

// applyRowCollapse shifts children so adjacent bordered boxes overlap by 1 column,
// and sets CollapsedEdges flags on the shared edges.
func applyRowCollapse(boxes []*Box, inputs []*Input) {
	for i := 1; i < len(boxes); i++ {
		if hasBorder(inputs[i-1].Border) && hasBorder(inputs[i].Border) {
			boxes[i].X -= i
			boxes[i-1].Collapsed.Right = true
			boxes[i].Collapsed.Left = true
		}
	}
}

// collapseInset returns the border inset to use when border-collapse is active.
// When collapsing, the parent border inset is 0 — children's outer edges form the frame.
func collapseInset(collapse bool) int {
	if collapse {
		return 0
	}
	return -1 // sentinel: use normal inset
}

// collapsedParentSize computes the adjusted parent size for collapsed children.
func collapsedParentSize(boxes []*Box, inputs []*Input, isRow bool) int {
	overlaps := countOverlaps(inputs)
	total := 0
	for _, b := range boxes {
		if isRow {
			total += b.Width
		} else {
			total += b.Height
		}
	}
	return total - overlaps
}

// countOverlaps counts how many adjacent bordered pairs exist.
func countOverlaps(inputs []*Input) int {
	count := 0
	for i := 1; i < len(inputs); i++ {
		if hasBorder(inputs[i-1].Border) && hasBorder(inputs[i].Border) {
			count++
		}
	}
	return count
}

// hasCollapsedEdge returns true if any edge in the CollapsedEdges is set.
func hasCollapsedEdge(c render.CollapsedEdges) bool {
	return !c.IsZero()
}
