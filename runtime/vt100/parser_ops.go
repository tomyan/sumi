package vt100

import "github.com/tomyan/sumi/runtime/render"

// putChar writes a character at the current cursor position and advances.
func (s *Screen) putChar(ch rune) {
	if s.curRow < 0 || s.curRow >= s.buf.Height() {
		return
	}
	if s.curCol >= s.buf.Width() {
		s.curCol = 0
		if s.curRow == s.scrollBot {
			s.scrollRegionUp()
		} else {
			s.curRow++
			if s.curRow >= s.buf.Height() {
				s.curRow = s.buf.Height() - 1
			}
		}
	}
	s.buf.SetStyledCell(s.curRow, s.curCol, ch, s.style)
	s.curCol++
}

// fullReset resets the entire screen state (RIS).
func (s *Screen) fullReset() {
	w, h := s.buf.Width(), s.buf.Height()
	s.curRow = 0
	s.curCol = 0
	s.style = render.Style{}
	s.savedRow = 0
	s.savedCol = 0
	s.scrollTop = 0
	s.scrollBot = h - 1
	s.sentinel = false
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			s.buf.SetCell(row, col, 0)
		}
	}
}

// eraseDisplay clears part of the screen: 0=below, 1=above, 2=all.
func (s *Screen) eraseDisplay(mode int) {
	w, h := s.buf.Width(), s.buf.Height()
	switch mode {
	case 0:
		for col := s.curCol; col < w; col++ {
			s.buf.SetCell(s.curRow, col, 0)
		}
		for row := s.curRow + 1; row < h; row++ {
			for col := 0; col < w; col++ {
				s.buf.SetCell(row, col, 0)
			}
		}
	case 1:
		for row := 0; row < s.curRow; row++ {
			for col := 0; col < w; col++ {
				s.buf.SetCell(row, col, 0)
			}
		}
		for col := 0; col <= s.curCol && col < w; col++ {
			s.buf.SetCell(s.curRow, col, 0)
		}
	case 2:
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				s.buf.SetCell(row, col, 0)
			}
		}
	}
}

// eraseLine clears part of the current line: 0=right, 1=left, 2=entire.
func (s *Screen) eraseLine(mode int) {
	w := s.buf.Width()
	if s.curRow < 0 || s.curRow >= s.buf.Height() {
		return
	}
	switch mode {
	case 0:
		for col := s.curCol; col < w; col++ {
			s.buf.SetCell(s.curRow, col, 0)
		}
	case 1:
		for col := 0; col <= s.curCol && col < w; col++ {
			s.buf.SetCell(s.curRow, col, 0)
		}
	case 2:
		for col := 0; col < w; col++ {
			s.buf.SetCell(s.curRow, col, 0)
		}
	}
}

// scrollRegionUp shifts rows within the scroll region up by one.
func (s *Screen) scrollRegionUp() {
	w := s.buf.Width()
	for row := s.scrollTop; row < s.scrollBot; row++ {
		for col := 0; col < w; col++ {
			c := s.buf.Cell(row+1, col)
			s.buf.SetStyledCell(row, col, c.Ch, c.Style)
		}
	}
	for col := 0; col < w; col++ {
		s.buf.SetCell(s.scrollBot, col, 0)
	}
}

// scrollRegionDown shifts rows within the scroll region down by one.
func (s *Screen) scrollRegionDown() {
	w := s.buf.Width()
	for row := s.scrollBot; row > s.scrollTop; row-- {
		for col := 0; col < w; col++ {
			c := s.buf.Cell(row-1, col)
			s.buf.SetStyledCell(row, col, c.Ch, c.Style)
		}
	}
	for col := 0; col < w; col++ {
		s.buf.SetCell(s.scrollTop, col, 0)
	}
}

func (s *Screen) scrollUp()   { s.scrollRegionUp() }
func (s *Screen) scrollDown() { s.scrollRegionDown() }

// insertLines inserts n blank lines at the cursor row.
func (s *Screen) insertLines(n int) {
	w, h := s.buf.Width(), s.buf.Height()
	for i := 0; i < n; i++ {
		for row := h - 1; row > s.curRow; row-- {
			for col := 0; col < w; col++ {
				c := s.buf.Cell(row-1, col)
				s.buf.SetStyledCell(row, col, c.Ch, c.Style)
			}
		}
		for col := 0; col < w; col++ {
			s.buf.SetCell(s.curRow, col, 0)
		}
	}
}

// deleteLines deletes n lines at the cursor row.
func (s *Screen) deleteLines(n int) {
	w, h := s.buf.Width(), s.buf.Height()
	for i := 0; i < n; i++ {
		for row := s.curRow; row < h-1; row++ {
			for col := 0; col < w; col++ {
				c := s.buf.Cell(row+1, col)
				s.buf.SetStyledCell(row, col, c.Ch, c.Style)
			}
		}
		for col := 0; col < w; col++ {
			s.buf.SetCell(h-1, col, 0)
		}
	}
}

// deleteChars deletes n characters at cursor, shifting the rest left.
func (s *Screen) deleteChars(n int) {
	w := s.buf.Width()
	if s.curRow < 0 || s.curRow >= s.buf.Height() {
		return
	}
	for col := s.curCol; col < w-n; col++ {
		c := s.buf.Cell(s.curRow, col+n)
		s.buf.SetStyledCell(s.curRow, col, c.Ch, c.Style)
	}
	for col := w - n; col < w; col++ {
		if col >= 0 {
			s.buf.SetCell(s.curRow, col, 0)
		}
	}
}

// insertChars inserts n blank characters at cursor, shifting the rest right.
func (s *Screen) insertChars(n int) {
	w := s.buf.Width()
	if s.curRow < 0 || s.curRow >= s.buf.Height() {
		return
	}
	for col := w - 1; col >= s.curCol+n; col-- {
		c := s.buf.Cell(s.curRow, col-n)
		s.buf.SetStyledCell(s.curRow, col, c.Ch, c.Style)
	}
	for col := s.curCol; col < s.curCol+n && col < w; col++ {
		s.buf.SetCell(s.curRow, col, 0)
	}
}

// decodeUTF8 decodes a UTF-8 sequence from p, returning the rune and byte count.
func decodeUTF8(p []byte) (rune, int) {
	if len(p) == 0 {
		return 0, 0
	}
	b := p[0]
	switch {
	case b < 0xC0:
		return rune(b), 1
	case b < 0xE0:
		if len(p) < 2 {
			return rune(b), 1
		}
		return rune(b&0x1F)<<6 | rune(p[1]&0x3F), 2
	case b < 0xF0:
		if len(p) < 3 {
			return rune(b), 1
		}
		return rune(b&0x0F)<<12 | rune(p[1]&0x3F)<<6 | rune(p[2]&0x3F), 3
	default:
		if len(p) < 4 {
			return rune(b), 1
		}
		return rune(b&0x07)<<18 | rune(p[1]&0x3F)<<12 | rune(p[2]&0x3F)<<6 | rune(p[3]&0x3F), 4
	}
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

func isIntermediate(b byte) bool { return b >= 0x20 && b <= 0x2F }

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// color256toRGB converts a 256-color index to RGB values.
func color256toRGB(idx int) (r, g, b uint8) {
	if idx < 0 || idx > 255 {
		return 0, 0, 0
	}
	if idx < 16 {
		return ansi16RGB[idx][0], ansi16RGB[idx][1], ansi16RGB[idx][2]
	}
	if idx < 232 {
		idx -= 16
		ri := idx / 36
		gi := (idx % 36) / 6
		bi := idx % 6
		return cubeVal(ri), cubeVal(gi), cubeVal(bi)
	}
	v := uint8(8 + (idx-232)*10)
	return v, v, v
}

func cubeVal(i int) uint8 {
	if i == 0 {
		return 0
	}
	return uint8(55 + i*40)
}

var ansi16RGB = [16][3]uint8{
	{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
	{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
	{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}
