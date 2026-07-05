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
