package layout

// HitTestPath returns the chain of Inputs from root to the deepest node
// whose Box contains (x, y). Ancestors stay on the path even when the
// point lies outside their own bounds (fixed-position descendants have
// viewport-relative coordinates). Returns nil when nothing is hit.
func HitTestPath(input *Input, box *Box, x, y int) []*Input {
	if input == nil || box == nil {
		return nil
	}

	// Check children depth-first — deepest match wins.
	for i, child := range input.Children {
		if child == nil {
			continue
		}
		var childBox *Box
		if i < len(box.Children) && box.Children[i] != nil {
			childBox = box.Children[i]
		}
		if path := HitTestPath(child, childBox, x, y); path != nil {
			return append([]*Input{input}, path...)
		}
	}

	if x >= box.X && x < box.X+box.Width && y >= box.Y && y < box.Y+box.Height {
		return []*Input{input}
	}
	return nil
}

// PathTo returns the chain of Inputs from root to target, or nil when
// target is not in the tree.
func PathTo(root, target *Input) []*Input {
	if root == nil {
		return nil
	}
	if root == target {
		return []*Input{root}
	}
	for _, child := range root.Children {
		if p := PathTo(child, target); p != nil {
			return append([]*Input{root}, p...)
		}
	}
	return nil
}

// HasClickHandlers returns true if any node in the tree handles "click".
func HasClickHandlers(input *Input) bool {
	if input == nil {
		return false
	}
	if input.On["click"] != nil {
		return true
	}
	for _, child := range input.Children {
		if HasClickHandlers(child) {
			return true
		}
	}
	return false
}
