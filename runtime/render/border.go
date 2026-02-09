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
