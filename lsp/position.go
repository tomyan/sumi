package lsp

import "strings"

// offsetToPosition converts a byte offset within text to an LSP Position:
// a zero-based line and a UTF-16 character offset within that line.
func offsetToPosition(text string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	line, lineStart := 0, 0
	for i := 0; i < offset; i++ {
		if text[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	return Position{Line: line, Character: utf16Len(text[lineStart:offset])}
}

// utf16Len counts the number of UTF-16 code units in s. Runes outside the
// Basic Multilingual Plane occupy two units (a surrogate pair).
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// firstLine returns text up to the first newline (or all of it if none).
func firstLine(text string) string {
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		return text[:i]
	}
	return text
}
