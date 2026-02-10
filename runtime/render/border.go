package render

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
