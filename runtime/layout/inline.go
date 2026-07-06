package layout

// Fragment is one line-rectangle of an inline run. Coordinates are
// relative to the owning Box's origin (they shift with it through
// absolutePositions untouched).
type Fragment struct {
	X, Y int
	Text string
}

// inlineItem is one participant in an inline formatting context: a text
// run, or an inline-block atom (an unbreakable pre-laid-out box).
type inlineItem struct {
	input *Input
	text  string // collapsed source text (text runs only)
	atom  bool
	box   *Box // atom's laid-out box (atoms only)
}

// layoutInlineChildren lays out inline-level children as one inline
// formatting context: text runs are gathered depth-first through
// display:inline elements (inline-block boxes join as atoms),
// whitespace collapses, lines break across run boundaries, and each
// run's Box receives box-relative Fragments plus a bounding rect.
// Inline elements keep their pairwise child structure with a union
// bounding rect.
func layoutInlineChildren(children []*Input, offsetX, offsetY, availW, availH int, align string) []*Box {
	var items []inlineItem
	gatherInlineItems(children, &items, availW, availH)
	perItem := breakInline(items, availW, align)
	itemIdx := 0
	return buildInlineBoxes(children, items, perItem, &itemIdx, offsetX, offsetY)
}

// gatherInlineItems appends the IFC participants under children
// depth-first, recursing through inline elements. Order matches
// buildInlineBoxes.
func gatherInlineItems(children []*Input, items *[]inlineItem, availW, availH int) {
	for _, c := range children {
		if c == nil || c.HiddenFromLayout() {
			continue
		}
		if c.Kind == KindText {
			*items = append(*items, inlineItem{input: c, text: transformText(c.Content, c.TextTransform)})
			continue
		}
		if c.Display == "inline-block" {
			*items = append(*items, inlineItem{input: c, atom: true, box: layoutNode(c, availW, availH)})
			continue
		}
		gatherInlineItems(c.Children, items, availW, availH)
	}
}

// buildInlineBoxes converts the broken runs back into a Box forest that
// mirrors the Input structure. display:none children keep their nil
// placeholder convention.
func buildInlineBoxes(children []*Input, items []inlineItem, perItem [][]Fragment, itemIdx *int, offsetX, offsetY int) []*Box {
	boxes := make([]*Box, len(children))
	for i, c := range children {
		if c == nil || c.HiddenFromLayout() {
			continue
		}
		if c.Kind == KindText {
			boxes[i] = fragmentBox(c, perItem[*itemIdx], offsetX, offsetY)
			*itemIdx++
			continue
		}
		if c.Display == "inline-block" {
			boxes[i] = placeAtomBox(items[*itemIdx], perItem[*itemIdx], offsetX, offsetY)
			*itemIdx++
			continue
		}
		boxes[i] = buildInlineElementBox(c, items, perItem, itemIdx, offsetX, offsetY)
	}
	return boxes
}

// placeAtomBox positions an atom's pre-laid-out box at its line slot
// (top-aligned on the line).
func placeAtomBox(item inlineItem, frags []Fragment, offsetX, offsetY int) *Box {
	box := item.box
	if len(frags) > 0 {
		box.X = offsetX + frags[0].X
		box.Y = offsetY + frags[0].Y
	}
	return box
}

// buildInlineElementBox builds an inline element's Box: children are
// built in container coordinates, the element takes their union rect,
// and the children are re-based onto the element's origin.
func buildInlineElementBox(c *Input, items []inlineItem, perItem [][]Fragment, itemIdx *int, offsetX, offsetY int) *Box {
	childBoxes := buildInlineBoxes(c.Children, items, perItem, itemIdx, offsetX, offsetY)
	box := &Box{
		Kind:          KindBox,
		Key:           c.Key,
		Style:         c.Style,
		HoverStyle:    c.HoverStyle,
		Hovered:       c.Hovered,
		FocusStyle:    c.FocusStyle,
		Focused:       c.Focused,
		Visibility:    c.Visibility,
		UnionBox:      true,
		Transitions:   c.Transitions,
		AnimationSpec: c.AnimationSpec,
		CursorCol:     -1,
		CursorRow:     -1,
		X:             offsetX,
		Y:             offsetY,
		Children:      childBoxes,
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
		Kind:          KindText,
		Key:           child.Key,
		Style:         child.Style,
		HoverStyle:    child.HoverStyle,
		Hovered:       child.Hovered,
		FocusStyle:    child.FocusStyle,
		Focused:       child.Focused,
		Visibility:    child.Visibility,
		Transitions:   child.Transitions,
		AnimationSpec: child.AnimationSpec,
		CursorCol:     -1,
		CursorRow:     -1,
		X:             offsetX,
		Y:             offsetY,
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
