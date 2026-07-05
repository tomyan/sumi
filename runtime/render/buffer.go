package render

import "io"

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

// Resize changes the buffer dimensions, clearing all cells.
func (b *Buffer) Resize(width, height int) {
	if width == b.width && height == b.height {
		b.Clear()
		return
	}
	b.width = width
	b.height = height
	b.cells = make([][]Cell, height)
	for i := range b.cells {
		b.cells[i] = make([]Cell, width)
	}
}

// ResizePreserve changes the buffer dimensions, keeping existing cell content
// that fits within the new bounds. New cells are zeroed. Cells beyond the new
// dimensions are discarded.
func (b *Buffer) ResizePreserve(width, height int) {
	if width == b.width && height == b.height {
		return
	}
	newCells := make([][]Cell, height)
	for row := range newCells {
		newCells[row] = make([]Cell, width)
		if row < b.height {
			copyW := b.width
			if copyW > width {
				copyW = width
			}
			copy(newCells[row][:copyW], b.cells[row][:copyW])
		}
	}
	b.width = width
	b.height = height
	b.cells = newCells
}

// Clear zeroes all cells without reallocating.
func (b *Buffer) Clear() {
	for row := range b.cells {
		for col := range b.cells[row] {
			b.cells[row][col] = Cell{}
		}
	}
}

// RenderDiff writes only the cells that differ between b (desired) and prev (current screen).
// After rendering, prev is updated to match b. prevW/prevH are the dimensions of
// what was previously on the terminal (may differ from prev's current dimensions
// if prev was resized before this call).
func (b *Buffer) RenderDiff(w io.Writer, prev *Buffer, prevTermW, prevTermH int) {
	buf := make([]byte, 0, b.width*b.height*2)
	var prevStyle Style
	styled := false
	prevRow, prevCol := -1, -1

	// Write changed cells.
	for row := 0; row < b.height; row++ {
		for col := 0; col < b.width; col++ {
			desired := b.cells[row][col]
			var current Cell
			if row < prev.height && col < prev.width {
				current = prev.cells[row][col]
			}
			if desired == current {
				prevCol = -1 // break adjacency
				continue
			}
			ch := desired.Ch
			if ch == 0 {
				ch = ' ' // clear cell
			}
			if row != prevRow || col != prevCol {
				buf = appendCUP(buf, row+1, col+1)
			}
			if desired.Style != prevStyle || (styled != !desired.Style.IsZero()) {
				if desired.Style.IsZero() {
					buf = append(buf, "\x1b[0m"...)
					styled = false
				} else {
					buf = appendSGR(buf, desired.Style)
					styled = true
				}
				prevStyle = desired.Style
			}
			buf = appendRune(buf, ch)
			prevRow = row
			prevCol = col + 1
		}
	}

	// Clear stale cells from the terminal that are beyond the new dimensions.
	if prevTermH > b.height || prevTermW > b.width {
		if styled {
			buf = append(buf, "\x1b[0m"...)
			styled = false
		}
		for row := b.height; row < prevTermH; row++ {
			buf = appendCUP(buf, row+1, 1)
			buf = append(buf, "\x1b[2K"...) // clear entire line
		}
		if prevTermW > b.width {
			for row := 0; row < b.height; row++ {
				buf = appendCUP(buf, row+1, b.width+1)
				buf = append(buf, "\x1b[0K"...) // clear to end of line
			}
		}
	}

	if styled {
		buf = append(buf, "\x1b[0m"...)
	}

	if len(buf) > 0 {
		writeSynchronized(w, buf)
	}

	// Update prev to match desired.
	if prev.width != b.width || prev.height != b.height {
		prev.Resize(b.width, b.height)
	}
	for row := 0; row < b.height; row++ {
		copy(prev.cells[row], b.cells[row])
	}
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

// SetStyledCell sets the character and style at (row, col), compositing
// translucent colours over the existing cell. Out-of-bounds is a no-op.
func (b *Buffer) SetStyledCell(row, col int, ch rune, style Style) {
	if row < 0 || row >= b.height || col < 0 || col >= b.width {
		return
	}
	b.cells[row][col] = Cell{Ch: ch, Style: b.compositeStyle(row, col, style)}
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

// SetStyledCellClipped sets a cell only if it falls within the clip region.
// A nil clip means no clipping (equivalent to SetStyledCell).
func (b *Buffer) SetStyledCellClipped(row, col int, ch rune, style Style, clip *Clip) {
	if clip != nil && !clip.Contains(row, col) {
		return
	}
	b.SetStyledCell(row, col, ch, style)
}

// WriteStyledTextClipped writes a styled string, skipping characters outside the clip region.
// A nil clip means no clipping (equivalent to WriteStyledText).
func (b *Buffer) WriteStyledTextClipped(row, col int, text string, style Style, clip *Clip) {
	if clip != nil && (row < clip.Top || row > clip.Bottom) {
		return
	}
	for _, ch := range text {
		if col >= b.width {
			break
		}
		if clip != nil && col > clip.Right {
			break
		}
		if clip != nil && col < clip.Left {
			col++
			continue
		}
		b.SetStyledCell(row, col, ch, style)
		col++
	}
}
