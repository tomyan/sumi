package layout

// CollectFocusables returns all focusable inputs in tree order.
// Nil children (display:none placeholders) are skipped.
func CollectFocusables(root *Input) []*Input {
	var focusables []*Input
	var walk func(n *Input)
	walk = func(n *Input) {
		if n == nil {
			return
		}
		if n.Focusable {
			focusables = append(focusables, n)
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	walk(root)
	return focusables
}

// FocusablePath returns the chain of Inputs from root to the index-th
// focusable (in tree order). Returns nil when the index is out of range.
func FocusablePath(root *Input, index int) []*Input {
	if index < 0 {
		return nil
	}
	seen := 0
	var found []*Input
	var walk func(n *Input, path []*Input) bool
	walk = func(n *Input, path []*Input) bool {
		if n == nil {
			return false
		}
		path = append(path, n)
		if n.Focusable {
			if seen == index {
				found = append([]*Input{}, path...)
				return true
			}
			seen++
		}
		for _, child := range n.Children {
			if walk(child, path) {
				return true
			}
		}
		return false
	}
	walk(root, nil)
	return found
}

// CycleFocus advances the focus index forward, wrapping around.
func CycleFocus(current, count int) int {
	if count <= 0 {
		return 0
	}
	return (current + 1) % count
}

// CycleFocusBackward moves the focus index backward, wrapping around.
func CycleFocusBackward(current, count int) int {
	if count <= 0 {
		return 0
	}
	return (current - 1 + count) % count
}
