package layout

// applyColumnCollapse shifts children so adjacent bordered boxes overlap by 1 row,
// and sets CollapsedEdges flags on the shared edges.
// parentH is used to extend the last child to fill any remaining space.
func applyColumnCollapse(boxes []*Box, inputs []*Input, parentH int) {
	overlaps := 0
	for i := 1; i < len(boxes); i++ {
		if hasBorder(inputs[i-1].Border) && hasBorder(inputs[i].Border) {
			overlaps++
			boxes[i].Y -= overlaps
			boxes[i-1].Collapsed.Bottom = true
			boxes[i].Collapsed.Top = true
		}
	}
	// Extend last child to fill remaining space (rounding from integer division)
	if parentH > 0 && len(boxes) > 0 {
		last := boxes[len(boxes)-1]
		gap := parentH - (last.Y + last.Height)
		if gap > 0 {
			last.Height += gap
		}
	}
}

// applyRowCollapse shifts children so adjacent bordered boxes overlap by 1 column,
// and sets CollapsedEdges flags on the shared edges.
// parentW is used to extend the last child to fill any remaining space.
func applyRowCollapse(boxes []*Box, inputs []*Input, parentW int) {
	overlaps := 0
	for i := 1; i < len(boxes); i++ {
		if hasBorder(inputs[i-1].Border) && hasBorder(inputs[i].Border) {
			overlaps++
			boxes[i].X -= overlaps
			boxes[i-1].Collapsed.Right = true
			boxes[i].Collapsed.Left = true
		}
	}
	// Extend last child to fill remaining space
	if parentW > 0 && len(boxes) > 0 {
		last := boxes[len(boxes)-1]
		gap := parentW - (last.X + last.Width)
		if gap > 0 {
			last.Width += gap
		}
	}
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
