package layout

import "github.com/tomyan/sumi/runtime/render"

// Block flow (display: block): block-level children stack vertically and
// fill the available width; consecutive inline-level children form an
// inline formatting context. Flex attributes (direction/gap/justify/
// align) do not apply in block flow.

// isInlineLevel reports whether a child participates in inline flow:
// plain text runs, and display:inline elements whose content is itself
// inline-level. Text with editing state, non-normal white-space, or
// explicit sizing stacks as block-level instead.
func isInlineLevel(c *Input) bool {
	if c.Kind == KindText {
		return c.Display != "block" && c.WhiteSpace == "" && !c.ContentEditable &&
			c.FixedWidth == 0 && c.FixedHeight == 0
	}
	if c.Display == "inline-block" {
		return true
	}
	if c.Display != "inline" {
		return false
	}
	for _, gc := range c.Children {
		if gc != nil && gc.Display != "none" && !isInlineLevel(gc) {
			return false
		}
	}
	return true
}

// layoutBlockFlow lays out a block container's flow children, returning
// boxes index-aligned with the children. display:contents elements are
// flattened first so their children participate directly in this flow,
// then reassembled into union-rect placeholder boxes. Adjacent block
// siblings' vertical margins collapse to their maximum (positive
// margins only); inline content between blocks resets the collapse
// context.
func layoutBlockFlow(children []*Input, offsetX, offsetY, availW, availH int) []*Box {
	var flat []*Input
	flattenContents(children, &flat)
	flatBoxes := layoutFlatBlockFlow(flat, offsetX, offsetY, availW, availH)
	idx := 0
	return reassembleContents(children, flatBoxes, &idx)
}

func layoutFlatBlockFlow(children []*Input, offsetX, offsetY, availW, availH int) []*Box {
	boxes := make([]*Box, len(children))
	cursorY := 0
	prevMarginBottom := 0
	for i := 0; i < len(children); {
		if isInlineLevel(children[i]) {
			i = layoutInlineSegment(children, i, boxes, offsetX, offsetY, &cursorY, availW, availH)
			prevMarginBottom = 0
			continue
		}
		child := children[i]
		if top := child.Margin.Top; prevMarginBottom > 0 && top > 0 {
			cursorY -= prevMarginBottom + top - maxInt(prevMarginBottom, top)
		}
		layoutBlockChild(child, boxes, i, offsetX, offsetY, &cursorY, availW, availH)
		prevMarginBottom = child.Margin.Bottom
		i++
	}
	return boxes
}

// flattenContents appends the flow participants, replacing each
// display:contents element with its own (recursively flattened)
// children.
func flattenContents(children []*Input, out *[]*Input) {
	for _, c := range children {
		if c == nil || c.HiddenFromLayout() {
			continue
		}
		if c.Display == "contents" {
			flattenContents(c.Children, out)
			continue
		}
		*out = append(*out, c)
	}
}

// reassembleContents rebuilds the original child structure from the
// flat flow boxes, wrapping each contents element's children in a
// union-rect placeholder that paints nothing of its own.
func reassembleContents(children []*Input, flatBoxes []*Box, idx *int) []*Box {
	boxes := make([]*Box, len(children))
	for i, c := range children {
		if c == nil || c.HiddenFromLayout() {
			continue
		}
		if c.Display != "contents" {
			boxes[i] = flatBoxes[*idx]
			*idx++
			continue
		}
		sub := reassembleContents(c.Children, flatBoxes, idx)
		style := c.Style
		style.BG = render.Color{} // contents generates no box: nothing painted
		box := &Box{
			Kind:      KindBox,
			Key:       c.Key,
			Style:     style,
			UnionBox:  true,
			CursorCol: -1,
			CursorRow: -1,
			Children:  sub,
		}
		unionChildRects(box, sub)
		boxes[i] = box
	}
	return boxes
}

// layoutInlineSegment lays out the run of inline-level children starting
// at index start as one IFC, advancing the flow cursor by the segment's
// line count. Returns the index after the segment.
func layoutInlineSegment(children []*Input, start int, boxes []*Box, offsetX, offsetY int, cursorY *int, availW, availH int) int {
	end := start
	for end < len(children) && isInlineLevel(children[end]) {
		end++
	}
	segTop := offsetY + *cursorY
	segBoxes := layoutInlineChildren(children[start:end], offsetX, segTop, availW, availH)
	height := 0
	for j, segBox := range segBoxes {
		boxes[start+j] = segBox
		if segBox != nil && segBox.Height > 0 && segBox.Y+segBox.Height-segTop > height {
			height = segBox.Y + segBox.Height - segTop
		}
	}
	*cursorY += height
	return end
}

// layoutBlockChild places one block-level child at the flow cursor,
// applying its margins (auto horizontal margins centre it).
func layoutBlockChild(child *Input, boxes []*Box, i, offsetX, offsetY int, cursorY *int, availW, availH int) {
	m := child.Margin
	*cursorY += m.Top
	childBox := layoutNode(child, maxInt(availW-m.horizontal(), 0), availH)
	childBox.X = offsetX + m.Left
	if m.autoCentreX() && childBox.Width < availW {
		childBox.X = offsetX + (availW-childBox.Width)/2
	}
	childBox.Y = offsetY + *cursorY
	*cursorY += childBox.Height + m.Bottom
	boxes[i] = childBox
}
