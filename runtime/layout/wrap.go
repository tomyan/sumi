package layout

import "unicode/utf8"

// wrapText breaks text into lines that fit within the given width (in runes).
// Preserves all whitespace exactly. Prefers breaking at space boundaries but
// falls back to character-level breaks for long words. Explicit newlines
// always start a new line.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if utf8.RuneCountInString(text) <= width {
		return []string{text}
	}

	runes := []rune(text)
	var lines []string
	lineStart := 0

	for lineStart < len(runes) {
		// Handle explicit newlines.
		if runes[lineStart] == '\n' {
			lines = append(lines, "")
			lineStart++
			continue
		}

		// Find the end of this line.
		remaining := len(runes) - lineStart
		if remaining <= width {
			// Rest fits on one line.
			lines = append(lines, string(runes[lineStart:]))
			break
		}

		// Check for newline within the width.
		lineEnd := lineStart + width
		for i := lineStart; i < lineEnd && i < len(runes); i++ {
			if runes[i] == '\n' {
				lines = append(lines, string(runes[lineStart:i]))
				lineStart = i + 1
				goto nextLine
			}
		}

		// Look backwards from the width boundary for a space to break at.
		{
			breakAt := -1
			for i := lineEnd - 1; i > lineStart; i-- {
				if runes[i] == ' ' {
					breakAt = i
					break
				}
			}
			if breakAt > lineStart {
				// Break after the space (space stays at end of current line).
				lines = append(lines, string(runes[lineStart:breakAt+1]))
				lineStart = breakAt + 1
			} else {
				// No space found — hard break at width.
				lines = append(lines, string(runes[lineStart:lineEnd]))
				lineStart = lineEnd
			}
		}
	nextLine:
	}

	if lineStart == len(runes) && (len(lines) == 0 || len(runes) > 0 && runes[len(runes)-1] == '\n') {
		lines = append(lines, "")
	}

	return lines
}
