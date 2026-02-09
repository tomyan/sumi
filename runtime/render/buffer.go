package render

// Cell represents a single terminal cell.
type Cell struct {
	Ch    rune
	Style Style
}

// Buffer is a 2D grid of terminal cells.
type Buffer struct {
	width  int
	height int
	cells  [][]Cell
}

// NewBuffer creates a buffer with the given dimensions.
func NewBuffer(width, height int) *Buffer {
	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
	}
	return &Buffer{width: width, height: height, cells: cells}
}

// Width returns the buffer width in columns.
func (b *Buffer) Width() int { return b.width }

// Height returns the buffer height in rows.
func (b *Buffer) Height() int { return b.height }

// SetCell sets the character at (row, col). Out-of-bounds is a no-op.
func (b *Buffer) SetCell(row, col int, ch rune) {
	if row < 0 || row >= b.height || col < 0 || col >= b.width {
		return
	}
	b.cells[row][col].Ch = ch
}

// Cell returns the cell at (row, col). Out-of-bounds returns a zero Cell.
func (b *Buffer) Cell(row, col int) Cell {
	if row < 0 || row >= b.height || col < 0 || col >= b.width {
		return Cell{}
	}
	return b.cells[row][col]
}

// WriteText writes a string starting at (row, col), truncating at the buffer edge.
func (b *Buffer) WriteText(row, col int, text string) {
	for _, ch := range text {
		if col >= b.width {
			break
		}
		b.SetCell(row, col, ch)
		col++
	}
}

// SetStyledCell sets the character and style at (row, col). Out-of-bounds is a no-op.
func (b *Buffer) SetStyledCell(row, col int, ch rune, style Style) {
	if row < 0 || row >= b.height || col < 0 || col >= b.width {
		return
	}
	b.cells[row][col] = Cell{Ch: ch, Style: style}
}

// WriteStyledText writes a styled string starting at (row, col), truncating at the buffer edge.
func (b *Buffer) WriteStyledText(row, col int, text string, style Style) {
	for _, ch := range text {
		if col >= b.width {
			break
		}
		b.SetStyledCell(row, col, ch, style)
		col++
	}
}
