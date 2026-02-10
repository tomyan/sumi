package layout

import "strings"

// wrapText breaks text into lines that fit within the given width.
// It wraps at word boundaries (spaces) when possible, falling back
// to character-level breaks for words longer than width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)

	var line strings.Builder
	for _, word := range words {
		// Handle words longer than width by char-breaking them
		for len(word) > width {
			if line.Len() > 0 {
				lines = append(lines, line.String())
				line.Reset()
			}
			lines = append(lines, word[:width])
			word = word[width:]
		}
		if word == "" {
			continue
		}

		if line.Len() == 0 {
			line.WriteString(word)
		} else if line.Len()+1+len(word) <= width {
			line.WriteByte(' ')
			line.WriteString(word)
		} else {
			lines = append(lines, line.String())
			line.Reset()
			line.WriteString(word)
		}
	}
	if line.Len() > 0 {
		lines = append(lines, line.String())
	}

	return lines
}
