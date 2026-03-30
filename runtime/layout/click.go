package layout

// FindClickHandler walks the Input/Box tree and returns the OnClick handler
// of the deepest Input whose Box contains (x, y). Returns nil if no handler
// is found at the click position.
func FindClickHandler(input *Input, box *Box, x, y int) func() {
	if box == nil || input == nil {
		return nil
	}
	hit := x >= box.X && x < box.X+box.Width && y >= box.Y && y < box.Y+box.Height
	if !hit {
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
		if h := FindClickHandler(child, childBox, x, y); h != nil {
			return h
		}
	}

	// No deeper handler — return this node's handler (may be nil).
	return input.OnClick
}

// HasClickHandlers returns true if any node in the tree has an OnClick handler.
func HasClickHandlers(input *Input) bool {
	if input == nil {
		return false
	}
	if input.OnClick != nil {
		return true
	}
	for _, child := range input.Children {
		if HasClickHandlers(child) {
			return true
		}
	}
	return false
}
