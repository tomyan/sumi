package render

// DrawBorder draws a single-line Unicode box border at (row, col) with the given
// width and height. Out-of-bounds portions are clipped. Width or height < 2, or
// style "" / "none" results in a no-op.
func (b *Buffer) DrawBorder(row, col, width, height int, style string) {
	if style == "" || style == "none" {
		return
	}
	if width < 2 || height < 2 {
		return
	}

	right := col + width - 1
	bottom := row + height - 1

	// Corners
	b.SetCell(row, col, '┌')
	b.SetCell(row, right, '┐')
	b.SetCell(bottom, col, '└')
	b.SetCell(bottom, right, '┘')

	// Top and bottom horizontal edges
	for c := col + 1; c < right; c++ {
		b.SetCell(row, c, '─')
		b.SetCell(bottom, c, '─')
	}

	// Left and right vertical edges
	for r := row + 1; r < bottom; r++ {
		b.SetCell(r, col, '│')
		b.SetCell(r, right, '│')
	}
}
