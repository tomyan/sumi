package edit

import "strings"

// LineCol returns the cursor's (row, col) within Value's lines.
func (s *State) LineCol() (row, col int) {
	col = s.Cursor
	for _, line := range strings.Split(s.Value, "\n") {
		length := len([]rune(line))
		if col <= length {
			return row, col
		}
		col -= length + 1 // the newline
		row++
	}
	return row, 0
}

// CursorUp moves the cursor one line up, keeping the column where the
// target line allows. At the first line it stays put.
func (s *State) CursorUp() {
	s.moveLine(-1)
}

// CursorDown moves the cursor one line down, keeping the column where
// the target line allows. At the last line it stays put.
func (s *State) CursorDown() {
	s.moveLine(1)
}

func (s *State) moveLine(delta int) {
	lines := strings.Split(s.Value, "\n")
	row, col := s.LineCol()
	target := row + delta
	if target < 0 || target >= len(lines) {
		return
	}
	if max := len([]rune(lines[target])); col > max {
		col = max
	}
	cursor := 0
	for i := 0; i < target; i++ {
		cursor += len([]rune(lines[i])) + 1
	}
	s.Cursor = cursor + col
	s.lastYank = false
}
