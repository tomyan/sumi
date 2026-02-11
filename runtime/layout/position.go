package layout

// filterVisible partitions children into visible and returns a mapping.
// Returns the visible children and a slice of original indices for each visible child.
func filterVisible(children []*Input) ([]*Input, []int) {
	var visible []*Input
	var indices []int
	for i, c := range children {
		if c.Display != "none" {
			visible = append(visible, c)
			indices = append(indices, i)
		}
	}
	return visible, indices
}

// spliceChildren creates a full-length children slice with nil placeholders
// for hidden elements. visibleBoxes[i] is placed at indices[i] in the result.
func spliceChildren(total int, visibleBoxes []*Box, indices []int) []*Box {
	result := make([]*Box, total)
	for i, idx := range indices {
		result[idx] = visibleBoxes[i]
	}
	return result
}

// applyRelativeOffsets shifts relatively-positioned boxes by their offsets.
// The parent's size is computed from flow positions before this is called,
// matching CSS behavior where relative offsets are purely visual.
func applyRelativeOffsets(boxes []*Box) {
	for _, b := range boxes {
		if b == nil || b.Position != "relative" {
			continue
		}
		// Top wins over Bottom; Left wins over Right
		if b.Top != 0 {
			b.Y += b.Top
		} else if b.Bottom != 0 {
			b.Y -= b.Bottom
		}
		if b.Left != 0 {
			b.X += b.Left
		} else if b.Right != 0 {
			b.X -= b.Right
		}
	}
}
