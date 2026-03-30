package layout

// FindClickHandler walks the Input/Box tree and returns the OnClick handler
// of the deepest Input whose Box contains (x, y). Returns nil if no handler
// is found at the click position.
//
// Always walks into all children because fixed-position descendants may have
// viewport-relative coordinates outside the parent's bounds.
func FindClickHandler(input *Input, box *Box, x, y int) func() {
	if input == nil {
		return nil
	}

	// Check children depth-first — deepest match wins.
	for i, child := range input.Children {
		if child == nil {
			continue
		}
		var childBox *Box
		if box != nil && i < len(box.Children) && box.Children[i] != nil {
			childBox = box.Children[i]
		}
		if h := FindClickHandler(child, childBox, x, y); h != nil {
			return h
		}
	}

	// This node is a click target only if the click is within its bounds.
	if box == nil {
		return nil
	}
	hit := x >= box.X && x < box.X+box.Width && y >= box.Y && y < box.Y+box.Height
	if hit {
		return input.OnClick
	}
	return nil
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
