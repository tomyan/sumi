package layout

// Fragment is one line-rectangle of an inline run. Coordinates are
// relative to the owning Box's origin (they shift with it through
// absolutePositions untouched).
type Fragment struct {
	X, Y int
	Text string
}

// isInlineContext reports whether a container's flow children can form
// an inline formatting context: text nodes with normal white-space and
// no sizing or editing state. Anything else falls back to stacking.
func isInlineContext(children []*Input) bool {
	if len(children) == 0 {
		return false
	}
	for _, c := range children {
		if c.Kind != KindText || c.WhiteSpace != "" || c.ContentEditable ||
			c.FixedWidth > 0 || c.FixedHeight > 0 {
			return false
		}
	}
	return true
}

// layoutInlineChildren lays out text runs as one inline formatting
// context: whitespace collapses, lines break across run boundaries, and
// each run's Box receives box-relative Fragments plus a bounding rect.
func layoutInlineChildren(children []*Input, offsetX, offsetY, availW int) []*Box {
	texts := make([]string, len(children))
	for i, child := range children {
		texts[i] = transformText(child.Content, child.TextTransform)
	}
	perRun := breakInline(texts, availW)

	boxes := make([]*Box, len(children))
	for i, child := range children {
		boxes[i] = fragmentBox(child, perRun[i], offsetX, offsetY)
	}
	return boxes
}

// fragmentBox builds a text Box from its line fragments (given in
// container-content coordinates), computing the bounding rect and
// re-basing the fragments onto the box origin.
func fragmentBox(child *Input, frags []Fragment, offsetX, offsetY int) *Box {
	box := &Box{
		Kind:       KindText,
		Key:        child.Key,
		Style:      child.Style,
		HoverStyle: child.HoverStyle,
		Hovered:    child.Hovered,
		FocusStyle: child.FocusStyle,
		Focused:    child.Focused,
		Visibility: child.Visibility,
		CursorCol:  -1,
		CursorRow:  -1,
		X:          offsetX,
		Y:          offsetY,
	}
	if len(frags) == 0 {
		return box
	}
	minX, minY := frags[0].X, frags[0].Y
	maxX, maxY := 0, 0
	for _, f := range frags {
		if f.X < minX {
			minX = f.X
		}
		if f.Y < minY {
			minY = f.Y
		}
		if end := f.X + runeLen(f.Text); end > maxX {
			maxX = end
		}
		if f.Y+1 > maxY {
			maxY = f.Y + 1
		}
	}
	box.X = offsetX + minX
	box.Y = offsetY + minY
	box.Width = maxX - minX
	box.Height = maxY - minY
	box.Fragments = make([]Fragment, len(frags))
	for i, f := range frags {
		box.Fragments[i] = Fragment{X: f.X - minX, Y: f.Y - minY, Text: f.Text}
	}
	return box
}
