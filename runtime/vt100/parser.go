package vt100

import "github.com/tomyan/sumi/runtime/render"

// parser state machine states
const (
	stateGround  = iota
	stateEscape  // saw ESC
	stateCSI     // saw ESC[
	stateParam   // reading CSI parameters
	statePrivate // saw ESC[? — private mode sequence
	stateOSC     // saw ESC] — operating system command
)

// Write implements io.Writer — feeds raw ANSI bytes into the screen.
func (s *Screen) Write(p []byte) (int, error) {
	state := stateGround
	var params []int
	cur := 0
	var oscBuf []byte

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
			switch b {
			case '[':
				state = stateCSI
				params = params[:0]
				cur = 0
			case ']':
				state = stateOSC
				oscBuf = oscBuf[:0]
			default:
				state = stateGround
			}

		case stateCSI:
			if b == '?' {
				state = statePrivate
			} else if isDigit(b) {
				cur = int(b - '0')
				state = stateParam
			} else {
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

		case statePrivate:
			// Consume everything until a letter (h/l/t etc)
			if b >= 0x40 && b <= 0x7e {
				state = stateGround
			}
			// digits and semicolons are silently consumed

		case stateOSC:
			if b == 0x07 { // BEL terminates OSC
				s.handleOSC(oscBuf)
				state = stateGround
			} else {
				oscBuf = append(oscBuf, b)
			}
		}
	}
	return len(p), nil
}

// handleOSC processes an OSC payload (everything between ESC] and BEL).
func (s *Screen) handleOSC(payload []byte) {
	if string(payload) == "999;done" {
		s.sentinel = true
	}
	// OSC 2 (title) and others are silently ignored.
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
	case 'm': // SGR — select graphic rendition
		s.applySGR(params)
	case 't': // window manipulation — consume and ignore
		// e.g. ESC[22;2t (save title), ESC[23;2t (restore title)
	}
}

// applySGR processes SGR parameters and updates the current style.
func (s *Screen) applySGR(params []int) {
	if len(params) == 0 {
		s.style = render.Style{}
		return
	}
	for _, code := range params {
		switch {
		case code == 0:
			s.style = render.Style{}
		case code == 1:
			s.style.Bold = true
		case code == 2:
			s.style.Dim = true
		case code == 3:
			s.style.Italic = true
		case code == 4:
			s.style.Underline = true
		case code == 7:
			s.style.Inverse = true
		case code == 9:
			s.style.Strikethrough = true
		case code >= 30 && code <= 37:
			s.style.FG = render.Color{Name: colorNames[code-30]}
		case code >= 40 && code <= 47:
			s.style.BG = render.Color{Name: colorNames[code-40]}
		}
	}
}

// colorNames maps ANSI offset (0-7) to color name.
var colorNames = [8]string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

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
