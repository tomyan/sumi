package vt100

// parser state machine states
const (
	stateGround = iota
	stateEscape // saw ESC
	stateCSI    // saw ESC[
	stateParam  // reading CSI parameters
)

// Write implements io.Writer — feeds raw ANSI bytes into the screen.
func (s *Screen) Write(p []byte) (int, error) {
	state := stateGround
	var params []int
	cur := 0

	for i := 0; i < len(p); i++ {
		b := p[i]
		switch state {
		case stateGround:
			if b == 0x1b {
				state = stateEscape
			} else {
				s.putChar(rune(b))
			}

		case stateEscape:
			if b == '[' {
				state = stateCSI
				params = params[:0]
				cur = 0
			} else {
				// Unknown escape — ignore
				state = stateGround
			}

		case stateCSI:
			if isDigit(b) {
				cur = int(b - '0')
				state = stateParam
			} else {
				// No params — dispatch immediately
				s.dispatchCSI(params, b)
				state = stateGround
			}

		case stateParam:
			if isDigit(b) {
				cur = cur*10 + int(b-'0')
			} else if b == ';' {
				params = append(params, cur)
				cur = 0
			} else {
				params = append(params, cur)
				s.dispatchCSI(params, b)
				state = stateGround
			}
		}
	}
	return len(p), nil
}

// dispatchCSI handles a complete CSI sequence with params and final byte.
func (s *Screen) dispatchCSI(params []int, final byte) {
	switch final {
	case 'H': // CUP — cursor position
		row, col := 1, 1
		if len(params) >= 1 {
			row = params[0]
		}
		if len(params) >= 2 {
			col = params[1]
		}
		s.curRow = row - 1
		s.curCol = col - 1
	case 'J': // ED — erase display
		if len(params) >= 1 && params[0] == 2 {
			s.clear()
		}
	}
}

// putChar writes a character at the current cursor position and advances.
func (s *Screen) putChar(ch rune) {
	if s.curRow < 0 || s.curRow >= s.buf.Height() {
		return
	}
	if s.curCol >= s.buf.Width() {
		return
	}
	s.buf.SetStyledCell(s.curRow, s.curCol, ch, s.style)
	s.curCol++
}

// clear resets all cells to empty.
func (s *Screen) clear() {
	w, h := s.buf.Width(), s.buf.Height()
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			s.buf.SetCell(row, col, 0)
		}
	}
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
