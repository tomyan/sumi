package layout

// Fragment is one line-rectangle of an inline run. Coordinates are
// relative to the owning Box's origin (they shift with it through
// absolutePositions untouched).
type Fragment struct {
	X, Y int
	Text string
}

// layoutInlineChildren lays out inline-level children as one inline
// formatting context: text runs are gathered depth-first through
// display:inline elements, whitespace collapses, lines break across run
// boundaries, and each run's Box receives box-relative Fragments plus a
// bounding rect. Inline elements keep their pairwise child structure
// with a union bounding rect.
func layoutInlineChildren(children []*Input, offsetX, offsetY, availW int) []*Box {
	var runs []*Input
	gatherInlineRuns(children, &runs)
	texts := make([]string, len(runs))
	for i, run := range runs {
		texts[i] = transformText(run.Content, run.TextTransform)
	}
	perRun := breakInline(texts, availW)
	runIdx := 0
	return buildInlineBoxes(children, perRun, &runIdx, offsetX, offsetY)
}

// gatherInlineRuns appends the text runs under children depth-first,
// recursing through inline elements. Order matches buildInlineBoxes.
func gatherInlineRuns(children []*Input, runs *[]*Input) {
	for _, c := range children {
		if c == nil || c.HiddenFromLayout() {
			continue
		}
		if c.Kind == KindText {
			*runs = append(*runs, c)
			continue
		}
		gatherInlineRuns(c.Children, runs)
	}
}

// buildInlineBoxes converts the broken runs back into a Box forest that
// mirrors the Input structure. display:none children keep their nil
// placeholder convention.
func buildInlineBoxes(children []*Input, perRun [][]Fragment, runIdx *int, offsetX, offsetY int) []*Box {
	boxes := make([]*Box, len(children))
	for i, c := range children {
		if c == nil || c.HiddenFromLayout() {
			continue
		}
		if c.Kind == KindText {
			boxes[i] = fragmentBox(c, perRun[*runIdx], offsetX, offsetY)
			*runIdx++
			continue
		}
		boxes[i] = buildInlineElementBox(c, perRun, runIdx, offsetX, offsetY)
	}
	return boxes
}

// buildInlineElementBox builds an inline element's Box: children are
// built in container coordinates, the element takes their union rect,
// and the children are re-based onto the element's origin.
func buildInlineElementBox(c *Input, perRun [][]Fragment, runIdx *int, offsetX, offsetY int) *Box {
	childBoxes := buildInlineBoxes(c.Children, perRun, runIdx, offsetX, offsetY)
	box := &Box{
		Kind:       KindBox,
		Key:        c.Key,
		Style:      c.Style,
		HoverStyle: c.HoverStyle,
		Hovered:    c.Hovered,
		FocusStyle: c.FocusStyle,
		Focused:    c.Focused,
		Visibility: c.Visibility,
		CursorCol:  -1,
		CursorRow:  -1,
		X:          offsetX,
		Y:          offsetY,
		Children:   childBoxes,
	}
	unionChildRects(box, childBoxes)
	return box
}

// unionChildRects sizes box to the union of its children's rects and
// re-bases the children onto the box origin.
func unionChildRects(box *Box, children []*Box) {
	first := true
	maxX, maxY := 0, 0
	for _, cb := range children {
		if cb == nil || (cb.Width == 0 && cb.Height == 0) {
			continue
		}
		if first || cb.X < box.X {
			box.X = cb.X
		}
		if first || cb.Y < box.Y {
			box.Y = cb.Y
		}
		if first || cb.X+cb.Width > maxX {
			maxX = cb.X + cb.Width
		}
		if first || cb.Y+cb.Height > maxY {
			maxY = cb.Y + cb.Height
		}
		first = false
	}
	if first {
		return
	}
	box.Width = maxX - box.X
	box.Height = maxY - box.Y
	for _, cb := range children {
		if cb != nil {
			cb.X -= box.X
			cb.Y -= box.Y
		}
	}
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
