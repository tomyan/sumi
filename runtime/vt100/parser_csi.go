package vt100

import "github.com/tomyan/sumi/runtime/render"

// dispatchCSI handles a complete CSI sequence with params and final byte.
func (s *Screen) dispatchCSI(params []int, final byte) {
	n := 1
	if len(params) >= 1 && params[0] > 0 {
		n = params[0]
	}

	switch final {
	case 'H', 'f': // CUP — cursor position
		row, col := 1, 1
		if len(params) >= 1 {
			row = params[0]
		}
		if len(params) >= 2 {
			col = params[1]
		}
		if row < 1 {
			row = 1
		}
		if col < 1 {
			col = 1
		}
		s.curRow = row - 1
		s.curCol = col - 1
	case 'A': // CUU — cursor up
		s.curRow -= n
		if s.curRow < 0 {
			s.curRow = 0
		}
	case 'B': // CUD — cursor down
		s.curRow += n
		if s.curRow >= s.buf.Height() {
			s.curRow = s.buf.Height() - 1
		}
	case 'C': // CUF — cursor forward
		s.curCol += n
		if s.curCol >= s.buf.Width() {
			s.curCol = s.buf.Width() - 1
		}
	case 'D': // CUB — cursor backward
		s.curCol -= n
		if s.curCol < 0 {
			s.curCol = 0
		}
	case 'J': // ED — erase display
		mode := 0
		if len(params) >= 1 {
			mode = params[0]
		}
		s.eraseDisplay(mode)
	case 'K': // EL — erase in line
		mode := 0
		if len(params) >= 1 {
			mode = params[0]
		}
		s.eraseLine(mode)
	case 'L': // IL — insert lines
		s.insertLines(n)
	case 'M': // DL — delete lines
		s.deleteLines(n)
	case 'P': // DCH — delete characters
		s.deleteChars(n)
	case '@': // ICH — insert characters
		s.insertChars(n)
	case 'r': // DECSTBM — set scrolling region
		top, bot := 1, s.buf.Height()
		if len(params) >= 1 && params[0] > 0 {
			top = params[0]
		}
		if len(params) >= 2 && params[1] > 0 {
			bot = params[1]
		}
		s.scrollTop = top - 1
		s.scrollBot = bot - 1
		s.curRow = 0
		s.curCol = 0
	case 'm': // SGR — select graphic rendition
		s.applySGR(params)
	case 't': // window manipulation — consume and ignore
	case 'G': // CHA — cursor character absolute
		s.curCol = n - 1
		if s.curCol < 0 {
			s.curCol = 0
		}
	case 'd': // VPA — vertical position absolute
		s.curRow = n - 1
		if s.curRow < 0 {
			s.curRow = 0
		}
	case 'S': // SU — scroll up
		for i := 0; i < n; i++ {
			s.scrollUp()
		}
	case 'T': // SD — scroll down
		for i := 0; i < n; i++ {
			s.scrollDown()
		}
	case 's': // SCP — Save Cursor Position
		s.savedRow = s.curRow
		s.savedCol = s.curCol
	case 'u': // RCP — Restore Cursor Position
		s.curRow = s.savedRow
		s.curCol = s.savedCol
	case 'c': // DA — device attributes, ignore
	case 'n': // DSR — device status report, ignore
	case 'X': // ECH — erase characters
		w := s.buf.Width()
		for j := 0; j < n && s.curCol+j < w; j++ {
			s.buf.SetCell(s.curRow, s.curCol+j, 0)
		}
	}
}

// applySGR processes SGR parameters and updates the current style.
func (s *Screen) applySGR(params []int) {
	if len(params) == 0 {
		s.style = render.Style{}
		return
	}
	for i := 0; i < len(params); i++ {
		code := params[i]
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
		case code == 22:
			s.style.Bold = false
			s.style.Dim = false
		case code == 23:
			s.style.Italic = false
		case code == 24:
			s.style.Underline = false
		case code == 27:
			s.style.Inverse = false
		case code == 29:
			s.style.Strikethrough = false
		case code >= 30 && code <= 37:
			s.style.FG = render.Color{Name: colorNames[code-30]}
		case code == 39:
			s.style.FG = render.Color{}
		case code == 38:
			// Extended FG: 38;2;r;g;b (24-bit) or 38;5;n (256-color)
			if i+1 < len(params) && params[i+1] == 2 && i+4 < len(params) {
				s.style.FG = render.Color{
					IsRGB: true,
					R:     uint8(clamp(params[i+2])),
					G:     uint8(clamp(params[i+3])),
					B:     uint8(clamp(params[i+4])),
				}
				i += 4
			} else if i+1 < len(params) && params[i+1] == 5 && i+2 < len(params) {
				r, g, b := color256toRGB(params[i+2])
				s.style.FG = render.Color{IsRGB: true, R: r, G: g, B: b}
				i += 2
			}
		case code >= 40 && code <= 47:
			s.style.BG = render.Color{Name: colorNames[code-40]}
		case code == 49:
			s.style.BG = render.Color{}
		case code == 48:
			// Extended BG: 48;2;r;g;b (24-bit) or 48;5;n (256-color)
			if i+1 < len(params) && params[i+1] == 2 && i+4 < len(params) {
				s.style.BG = render.Color{
					IsRGB: true,
					R:     uint8(clamp(params[i+2])),
					G:     uint8(clamp(params[i+3])),
					B:     uint8(clamp(params[i+4])),
				}
				i += 4
			} else if i+1 < len(params) && params[i+1] == 5 && i+2 < len(params) {
				r, g, b := color256toRGB(params[i+2])
				s.style.BG = render.Color{IsRGB: true, R: r, G: g, B: b}
				i += 2
			}
		case code >= 90 && code <= 97:
			s.style.FG = render.Color{Name: brightColorNames[code-90]}
		case code >= 100 && code <= 107:
			s.style.BG = render.Color{Name: brightColorNames[code-100]}
		}
	}
}

// brightColorNames maps bright ANSI offset (0-7) to color name.
var brightColorNames = [8]string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

// colorNames maps ANSI offset (0-7) to color name.
var colorNames = [8]string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
