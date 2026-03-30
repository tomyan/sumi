package layout

import (
	"strings"
	"unicode/utf8"
)

// wrapText breaks text into lines that fit within the given width (in runes).
// It wraps at word boundaries (spaces) when possible, falling back
// to character-level breaks for words longer than width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if utf8.RuneCountInString(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)

	var line strings.Builder
	lineLen := 0

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)

		// Handle words longer than width by rune-breaking them
		for wordLen > width {
			if lineLen > 0 {
				lines = append(lines, line.String())
				line.Reset()
				lineLen = 0
			}
			runes := []rune(word)
			lines = append(lines, string(runes[:width]))
			word = string(runes[width:])
			wordLen = utf8.RuneCountInString(word)
		}
		if word == "" {
			continue
		}

		if lineLen == 0 {
			line.WriteString(word)
			lineLen = wordLen
		} else if lineLen+1+wordLen <= width {
			line.WriteByte(' ')
			line.WriteString(word)
			lineLen += 1 + wordLen
		} else {
			lines = append(lines, line.String())
			line.Reset()
			line.WriteString(word)
			lineLen = wordLen
		}
	}
	if lineLen > 0 {
		lines = append(lines, line.String())
	}

	return lines
}
