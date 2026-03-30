package layout

import "unicode/utf8"

// cursorToVisual converts a flat character offset into visual (row, col)
// coordinates within wrapped lines. If lines is nil (no wrapping), the
// cursor is on row 0 at the given offset.
func cursorToVisual(offset int, lines []string, wrapW int) (row, col int) {
	if lines == nil {
		return 0, offset
	}
	// Walk through wrapped lines. Each line's rune count is the number
	// of characters consumed from the flat string.
	pos := 0
	for i, line := range lines {
		lineLen := utf8.RuneCountInString(line)
		if pos+lineLen > offset {
			return i, offset - pos
		}
		pos += lineLen
	}
	// Cursor at or past the end — place at end of last line.
	lastLine := len(lines) - 1
	lastLineLen := utf8.RuneCountInString(lines[lastLine])
	return lastLine, lastLineLen
}
