package layout

// controlTags are elements that receive focus without a focusable
// attribute. Grows as controls are implemented (input, textarea, a,
// select, summary).
var controlTags = map[string]bool{
	"button":   true,
	"input":    true,
	"select":   true,
	"summary":  true,
	"textarea": true,
}

// IsFocusable reports whether a node participates in focus traversal:
// standard controls, anchors with an href, and elements with
// focusable="true" — unless disabled.
func IsFocusable(n *Input) bool {
	if d, ok := n.Attrs["disabled"]; ok && d != "false" {
		return false
	}
	if n.Tag == "a" {
		return n.Attrs["href"] != ""
	}
	return n.Focusable || controlTags[n.Tag]
}

// CollectFocusables returns all focusable inputs in tree order.
// Nil children (display:none placeholders) are skipped.
func CollectFocusables(root *Input) []*Input {
	var focusables []*Input
	var walk func(n *Input)
	walk = func(n *Input) {
		if n == nil || n.HiddenFromLayout() {
			return
		}
		if IsFocusable(n) {
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
		if n == nil || n.HiddenFromLayout() {
			return false
		}
		path = append(path, n)
		if IsFocusable(n) {
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
