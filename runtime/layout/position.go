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
