package layout

// HasHoverStyles returns true if any node in the tree has a non-zero HoverStyle.
func HasHoverStyles(input *Input) bool {
	if input == nil {
		return false
	}
	if !input.HoverStyle.IsZero() {
		return true
	}
	for _, child := range input.Children {
		if HasHoverStyles(child) {
			return true
		}
	}
	return false
}
