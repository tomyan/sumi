package render

// CollapsedEdges flags which edges of a box are shared with an adjacent border.
// When an edge is collapsed, its corners use junction characters instead of
// normal box-drawing corners.
type CollapsedEdges struct {
	Top, Right, Bottom, Left bool
}

// IsZero returns true if no edges are collapsed.
func (c CollapsedEdges) IsZero() bool {
	return !c.Top && !c.Right && !c.Bottom && !c.Left
}

// DrawBorder draws a single-line Unicode box border at (row, col) with the given
// width and height. Out-of-bounds portions are clipped. Width or height < 2, or
// style "" / "none" results in a no-op.
func (b *Buffer) DrawBorder(row, col, width, height int, borderStyle string) {
	b.DrawStyledBorder(row, col, width, height, borderStyle, Style{})
}

// DrawStyledBorder draws a styled single-line Unicode box border at (row, col)
// with the given width and height. Out-of-bounds portions are clipped.
// Width or height < 2, or borderStyle "" / "none" results in a no-op.
func (b *Buffer) DrawStyledBorder(row, col, width, height int, borderStyle string, style Style) {
	if borderStyle == "" || borderStyle == "none" {
		return
	}
	if width < 2 || height < 2 {
		return
	}

	right := col + width - 1
	bottom := row + height - 1

	// Corners
	b.SetStyledCell(row, col, '┌', style)
	b.SetStyledCell(row, right, '┐', style)
	b.SetStyledCell(bottom, col, '└', style)
	b.SetStyledCell(bottom, right, '┘', style)

	// Top and bottom horizontal edges
	for c := col + 1; c < right; c++ {
		b.SetStyledCell(row, c, '─', style)
		b.SetStyledCell(bottom, c, '─', style)
	}

	// Left and right vertical edges
	for r := row + 1; r < bottom; r++ {
		b.SetStyledCell(r, col, '│', style)
		b.SetStyledCell(r, right, '│', style)
	}
}

// DrawBorderTitle renders a title string on the top edge of a border.
// The pattern is: ┌─ Title ───┐ (title starts at col+3, preceded by "─ " and followed by " ─…").
// Titles longer than width-4 are truncated. Empty title or width < 6 is a no-op.
func (b *Buffer) DrawBorderTitle(row, col, width int, title string) {
	b.DrawStyledBorderTitle(row, col, width, title, Style{})
}

// DrawStyledBorderTitle renders a styled title string on the top edge of a border.
func (b *Buffer) DrawStyledBorderTitle(row, col, width int, title string, style Style) {
	if title == "" || width < 6 {
		return
	}
	maxLen := width - 4 // "─ " before + " " after + corners
	runes := []rune(title)
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	// Write space before title at col+2
	b.SetStyledCell(row, col+2, ' ', style)
	// Write title characters starting at col+3
	for i, ch := range runes {
		b.SetStyledCell(row, col+3+i, ch, style)
	}
	// Write space after title
	b.SetStyledCell(row, col+3+len(runes), ' ', style)
}

// DrawCollapsedBorder draws a border with junction characters at corners
// where edges are collapsed (shared with adjacent borders).
func (b *Buffer) DrawCollapsedBorder(row, col, width, height int, borderStyle string, style Style, collapsed CollapsedEdges) {
	if borderStyle == "" || borderStyle == "none" {
		return
	}
	if width < 2 || height < 2 {
		return
	}

	right := col + width - 1
	bottom := row + height - 1

	// Corners use JunctionChar based on which edges connect
	b.SetStyledCell(row, col, cornerChar(collapsed.Top, collapsed.Left, true, true), style)
	b.SetStyledCell(row, right, cornerChar(collapsed.Top, collapsed.Right, true, false), style)
	b.SetStyledCell(bottom, col, cornerChar(collapsed.Bottom, collapsed.Left, false, true), style)
	b.SetStyledCell(bottom, right, cornerChar(collapsed.Bottom, collapsed.Right, false, false), style)

	// Top and bottom horizontal edges
	for c := col + 1; c < right; c++ {
		b.mergeJunction(row, c, false, true, false, true, style)
		b.mergeJunction(bottom, c, false, true, false, true, style)
	}

	// Left and right vertical edges
	for r := row + 1; r < bottom; r++ {
		b.mergeJunction(r, col, true, false, true, false, style)
		b.mergeJunction(r, right, true, false, true, false, style)
	}
}

// mergeJunction writes a border character at (row, col), merging with any existing
// junction character. If the cell already has a box-drawing character, the
// directions are OR'd together to produce the correct junction.
func (b *Buffer) mergeJunction(row, col int, up, right, down, left bool, style Style) {
	if existing := b.Cell(row, col); existing.Ch != 0 {
		eu, er, ed, el := junctionDirs(existing.Ch)
		up = up || eu
		right = right || er
		down = down || ed
		left = left || el
	}
	b.SetStyledCell(row, col, JunctionChar(up, right, down, left), style)
}

// junctionDirs returns the directional connections of a box-drawing character.
func junctionDirs(ch rune) (up, right, down, left bool) {
	if key, ok := reverseJunctionTable[ch]; ok {
		return key&1 != 0, key&2 != 0, key&4 != 0, key&8 != 0
	}
	return false, false, false, false
}

// cornerChar returns the junction character for a corner.
// collapsedV/collapsedH indicate whether the vertical/horizontal neighbor edge is collapsed.
// isTop/isLeft indicate which corner we're computing.
func cornerChar(collapsedV, collapsedH, isTop, isLeft bool) rune {
	// A normal corner always has connections in two directions.
	// Collapsed edges add connections in the opposite direction.
	up := !isTop || collapsedV
	down := isTop || collapsedV
	right := isLeft || collapsedH
	left := !isLeft || collapsedH
	return JunctionChar(up, right, down, left)
}
