package layout

// FindCursor walks the layout tree and returns the first Box with an active cursor
// (CursorCol >= 0). Returns nil if no cursor is found.
func FindCursor(box *Box) *Box {
	if box == nil {
		return nil
	}
	if box.CursorCol >= 0 && box.CursorRow >= 0 {
		return box
	}
	for _, child := range box.Children {
		if found := FindCursor(child); found != nil {
			return found
		}
	}
	return nil
}
