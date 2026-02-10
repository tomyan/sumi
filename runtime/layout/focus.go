package layout

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
