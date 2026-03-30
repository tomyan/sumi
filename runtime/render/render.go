package render

import (
	"io"
	"strconv"
)

// RenderTo renders the buffer to a writer using ANSI cursor-addressed output.
// Only non-empty cells (Ch != 0) are rendered. Styled cells emit ANSI SGR
// sequences before the character.
func (b *Buffer) RenderTo(w io.Writer) {
	b.RenderToOffset(w, 0, 0)
}

// RenderToOffset renders the buffer with all cursor positions shifted by the
// given row and column offsets. Batches all output into a single write.
// Skips redundant cursor moves for adjacent cells and redundant SGR changes.
func (b *Buffer) RenderToOffset(w io.Writer, rowOffset, colOffset int) {
	// Pre-allocate: ~20 bytes per non-empty cell is a reasonable estimate.
	buf := make([]byte, 0, b.width*b.height*4)
	prevRow, prevCol := -1, -1
	var prevStyle Style
	styled := false

	for row := 0; row < b.height; row++ {
		for col := 0; col < b.width; col++ {
			c := b.cells[row][col]
			if c.Ch == 0 {
				prevCol = -1 // break adjacency
				continue
			}

			// Cursor positioning: skip if this cell is right after the previous one.
			if row != prevRow || col != prevCol {
				buf = appendCUP(buf, row+1+rowOffset, col+1+colOffset)
			}

			// Style: only emit SGR when style changes.
			if c.Style != prevStyle || (styled != !c.Style.IsZero()) {
				if c.Style.IsZero() {
					buf = append(buf, "\x1b[0m"...)
					styled = false
				} else {
					buf = appendSGR(buf, c.Style)
					styled = true
				}
				prevStyle = c.Style
			}

			// Character.
			buf = appendRune(buf, c.Ch)
			prevRow = row
			prevCol = col + 1 // next expected position
		}
	}
	if styled {
		buf = append(buf, "\x1b[0m"...)
	}
	w.Write(buf)
}

// RenderWithClear clears the screen and renders the entire buffer in a single
// buffered write, avoiding the visible flash of separate clear + render.
func (b *Buffer) RenderWithClear(w io.Writer) {
	buf := make([]byte, 0, b.width*b.height*4+10)
	buf = append(buf, "\x1b[2J\x1b[H"...) // clear screen + cursor home
	var prevStyle Style
	styled := false
	prevRow, prevCol := -1, -1

	for row := 0; row < b.height; row++ {
		for col := 0; col < b.width; col++ {
			c := b.cells[row][col]
			if c.Ch == 0 {
				prevCol = -1
				continue
			}
			if row != prevRow || col != prevCol {
				buf = appendCUP(buf, row+1, col+1)
			}
			if c.Style != prevStyle || (styled != !c.Style.IsZero()) {
				if c.Style.IsZero() {
					buf = append(buf, "\x1b[0m"...)
					styled = false
				} else {
					buf = appendSGR(buf, c.Style)
					styled = true
				}
				prevStyle = c.Style
			}
			buf = appendRune(buf, c.Ch)
			prevRow = row
			prevCol = col + 1
		}
	}
	if styled {
		buf = append(buf, "\x1b[0m"...)
	}
	w.Write(buf)
}

// appendCUP appends a CUrsor Position escape sequence: ESC[row;colH
func appendCUP(buf []byte, row, col int) []byte {
	buf = append(buf, "\x1b["...)
	buf = strconv.AppendInt(buf, int64(row), 10)
	buf = append(buf, ';')
	buf = strconv.AppendInt(buf, int64(col), 10)
	buf = append(buf, 'H')
	return buf
}

// appendSGR appends the full SGR escape sequence for a style (reset + attributes).
func appendSGR(buf []byte, s Style) []byte {
	buf = append(buf, "\x1b[0"...)
	if s.Bold {
		buf = append(buf, ";1"...)
	}
	if s.Dim {
		buf = append(buf, ";2"...)
	}
	if s.Italic {
		buf = append(buf, ";3"...)
	}
	if s.Underline {
		buf = append(buf, ";4"...)
	}
	if s.Inverse {
		buf = append(buf, ";7"...)
	}
	if s.Strikethrough {
		buf = append(buf, ";9"...)
	}
	if s.FG.IsRGB {
		buf = append(buf, ";38;2;"...)
		buf = strconv.AppendInt(buf, int64(s.FG.R), 10)
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(s.FG.G), 10)
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(s.FG.B), 10)
	} else if code, ok := colorToFGCode(s.FG.Name); ok {
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(code), 10)
	}
	if s.BG.IsRGB {
		buf = append(buf, ";48;2;"...)
		buf = strconv.AppendInt(buf, int64(s.BG.R), 10)
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(s.BG.G), 10)
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(s.BG.B), 10)
	} else if code, ok := colorToBGCode(s.BG.Name); ok {
		buf = append(buf, ';')
		buf = strconv.AppendInt(buf, int64(code), 10)
	}
	buf = append(buf, 'm')
	return buf
}

// appendRune appends a UTF-8 encoded rune to the byte slice.
func appendRune(buf []byte, r rune) []byte {
	var tmp [4]byte
	n := encodeRune(tmp[:], r)
	return append(buf, tmp[:n]...)
}

// encodeRune writes the UTF-8 encoding of r into p and returns the number of bytes written.
func encodeRune(p []byte, r rune) int {
	switch {
	case r < 0x80:
		p[0] = byte(r)
		return 1
	case r < 0x800:
		p[0] = byte(0xC0 | (r >> 6))
		p[1] = byte(0x80 | (r & 0x3F))
		return 2
	case r < 0x10000:
		p[0] = byte(0xE0 | (r >> 12))
		p[1] = byte(0x80 | ((r >> 6) & 0x3F))
		p[2] = byte(0x80 | (r & 0x3F))
		return 3
	default:
		p[0] = byte(0xF0 | (r >> 18))
		p[1] = byte(0x80 | ((r >> 12) & 0x3F))
		p[2] = byte(0x80 | ((r >> 6) & 0x3F))
		p[3] = byte(0x80 | (r & 0x3F))
		return 4
	}
}
