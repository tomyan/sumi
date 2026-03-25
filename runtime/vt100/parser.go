package vt100

// parser state machine states
const (
	stateGround  = iota
	stateEscape  // saw ESC
	stateCSI     // saw ESC[
	stateParam   // reading CSI parameters
	statePrivate // saw ESC[? — private mode sequence
	stateOSC     // saw ESC] — operating system command
	stateCharset // saw ESC( or ESC) etc — expecting one designator byte
	stateCSIIgn  // consuming an unrecognized CSI sequence until final byte
	stateOSCEsc  // saw ESC inside OSC — expecting \ for ST
)

// Write implements io.Writer — feeds raw ANSI bytes into the screen.
// Parser state persists across calls so split sequences are handled correctly.
func (s *Screen) Write(p []byte) (int, error) {
	for i := 0; i < len(p); i++ {
		b := p[i]
		switch s.pState {
		case stateGround:
			s.handleGround(b, p, &i)
		case stateEscape:
			s.handleEscape(b)
		case stateCharset:
			s.pState = stateGround
		case stateCSI:
			s.handleCSI(b)
		case stateParam:
			s.handleParam(b)
		case stateCSIIgn:
			if b >= 0x40 && b <= 0x7e {
				s.pState = stateGround
			}
		case statePrivate:
			if b >= 0x40 && b <= 0x7e {
				s.pState = stateGround
			}
		case stateOSC:
			s.handleOSCByte(b)
		case stateOSCEsc:
			s.handleOSCEsc(b, &i)
		}
	}
	return len(p), nil
}

func (s *Screen) handleGround(b byte, p []byte, i *int) {
	switch {
	case b == 0x1b:
		s.pState = stateEscape
	case b == '\r':
		s.curCol = 0
	case b == '\n':
		if s.curRow == s.scrollBot {
			s.scrollRegionUp()
		} else {
			s.curRow++
			if s.curRow >= s.buf.Height() {
				s.curRow = s.buf.Height() - 1
			}
		}
	case b == '\b':
		if s.curCol > 0 {
			s.curCol--
		}
	case b == '\t':
		next := (s.curCol + 8) &^ 7
		if next > s.buf.Width() {
			next = s.buf.Width()
		}
		s.curCol = next
	case b < 0x20:
		// ignore other control characters
	case b < 0x80:
		s.putChar(rune(b))
	default:
		r, size := decodeUTF8(p[*i:])
		if size > 0 {
			s.putChar(r)
			*i += size - 1
		}
	}
}

func (s *Screen) handleEscape(b byte) {
	switch b {
	case '[':
		s.pState = stateCSI
		s.pParams = s.pParams[:0]
		s.pCur = 0
	case ']':
		s.pState = stateOSC
		s.pOSCBuf = s.pOSCBuf[:0]
	case '(', ')', '*', '+':
		s.pState = stateCharset
	case '=', '>':
		s.pState = stateGround
	case 'M': // RI — Reverse Index
		if s.curRow == s.scrollTop {
			s.scrollRegionDown()
		} else if s.curRow > 0 {
			s.curRow--
		}
		s.pState = stateGround
	case 'D': // IND — Index
		if s.curRow == s.scrollBot {
			s.scrollRegionUp()
		} else {
			s.curRow++
			if s.curRow >= s.buf.Height() {
				s.curRow = s.buf.Height() - 1
			}
		}
		s.pState = stateGround
	case 'E': // NEL — Next Line
		s.curCol = 0
		if s.curRow == s.scrollBot {
			s.scrollRegionUp()
		} else {
			s.curRow++
			if s.curRow >= s.buf.Height() {
				s.curRow = s.buf.Height() - 1
			}
		}
		s.pState = stateGround
	case '7': // DECSC — Save Cursor
		s.savedRow = s.curRow
		s.savedCol = s.curCol
		s.pState = stateGround
	case '8': // DECRC — Restore Cursor
		s.curRow = s.savedRow
		s.curCol = s.savedCol
		s.pState = stateGround
	case 'c': // RIS — Full Reset
		s.fullReset()
		s.pState = stateGround
	default:
		s.pState = stateGround
	}
}

func (s *Screen) handleCSI(b byte) {
	if b == '?' || b == '>' || b == '=' {
		s.pState = statePrivate
	} else if isIntermediate(b) {
		s.pState = stateCSIIgn
	} else if isDigit(b) {
		s.pCur = int(b - '0')
		s.pState = stateParam
	} else {
		s.dispatchCSI(s.pParams, b)
		s.pState = stateGround
	}
}

func (s *Screen) handleParam(b byte) {
	if isDigit(b) {
		s.pCur = s.pCur*10 + int(b-'0')
	} else if b == ';' {
		s.pParams = append(s.pParams, s.pCur)
		s.pCur = 0
	} else if isIntermediate(b) {
		s.pState = stateCSIIgn
	} else {
		s.pParams = append(s.pParams, s.pCur)
		s.dispatchCSI(s.pParams, b)
		s.pState = stateGround
	}
}

func (s *Screen) handleOSCByte(b byte) {
	if b == 0x07 {
		s.handleOSC(s.pOSCBuf)
		s.pState = stateGround
	} else if b == 0x1b {
		s.pState = stateOSCEsc
	} else {
		s.pOSCBuf = append(s.pOSCBuf, b)
	}
}

func (s *Screen) handleOSCEsc(b byte, i *int) {
	if b == '\\' {
		s.handleOSC(s.pOSCBuf)
		s.pState = stateGround
	} else {
		s.pOSCBuf = nil
		s.pState = stateEscape
		*i--
	}
}

// handleOSC processes an OSC payload (everything between ESC] and BEL/ST).
func (s *Screen) handleOSC(payload []byte) {
	if string(payload) == "999;done" {
		s.sentinel = true
	}
}
