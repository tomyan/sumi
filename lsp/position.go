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

// positionToOffset converts an LSP Position (zero-based line, UTF-16
// character offset within the line) to a byte offset within text. Positions
// past the end of a line clamp to the line's newline; positions past the end
// of the document clamp to len(text).
func positionToOffset(text string, pos Position) int {
	line, offset := 0, 0
	for line < pos.Line {
		nl := strings.IndexByte(text[offset:], '\n')
		if nl < 0 {
			return len(text)
		}
		offset += nl + 1
		line++
	}
	return offset + utf16OffsetToByte(lineAt(text, offset), pos.Character)
}

// lineAt returns text from offset up to (excluding) the next newline.
func lineAt(text string, offset int) string {
	rest := text[offset:]
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		return rest[:nl]
	}
	return rest
}

// utf16OffsetToByte returns the byte offset within line corresponding to the
// given UTF-16 character offset, clamping to len(line) when it runs past.
func utf16OffsetToByte(line string, char int) int {
	units := 0
	for i, r := range line {
		if units >= char {
			return i
		}
		if r > 0xFFFF {
			units += 2
		} else {
			units++
		}
	}
	return len(line)
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
