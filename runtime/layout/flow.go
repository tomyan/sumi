package layout

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
// boxes index-aligned with the children. Adjacent block siblings'
// vertical margins collapse to their maximum (positive margins only);
// inline content between blocks resets the collapse context.
func layoutBlockFlow(children []*Input, offsetX, offsetY, availW, availH int) []*Box {
	boxes := make([]*Box, len(children))
	cursorY := 0
	prevMarginBottom := 0
	for i := 0; i < len(children); {
		if isInlineLevel(children[i]) {
			i = layoutInlineSegment(children, i, boxes, offsetX, offsetY, &cursorY, availW)
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

// layoutInlineSegment lays out the run of inline-level children starting
// at index start as one IFC, advancing the flow cursor by the segment's
// line count. Returns the index after the segment.
func layoutInlineSegment(children []*Input, start int, boxes []*Box, offsetX, offsetY int, cursorY *int, availW int) int {
	end := start
	for end < len(children) && isInlineLevel(children[end]) {
		end++
	}
	segTop := offsetY + *cursorY
	segBoxes := layoutInlineChildren(children[start:end], offsetX, segTop, availW)
	height := 0
	for j, segBox := range segBoxes {
		boxes[start+j] = segBox
		if segBox.Fragments != nil && segBox.Y+segBox.Height-segTop > height {
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
