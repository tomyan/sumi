package render

import "strings"

// ToStyledText converts the buffer to styled markup text.
// Style changes produce <<attrs>> openers; returning to plain emits <</>>.
// Attributes are comma-separated, sorted: bg:COLOR, then fg COLOR, then
// boolean attributes alphabetically.
func (b *Buffer) ToStyledText() string {
	rows := make([]string, b.height)
	for r := 0; r < b.height; r++ {
		rows[r] = rowToStyledText(b.cells[r])
	}
	return trimTrailingEmptyRows(rows)
}

func rowToStyledText(cells []Cell) string {
	// Find last non-empty cell to avoid trailing styled spaces.
	lastContent := -1
	for i := len(cells) - 1; i >= 0; i-- {
		if cells[i].Ch != 0 {
			lastContent = i
			break
		}
	}
	if lastContent < 0 {
		return ""
	}

	var sb strings.Builder
	var current Style
	styled := false

	for i := 0; i <= lastContent; i++ {
		c := cells[i]
		ch := c.Ch
		if ch == 0 {
			ch = ' '
		}

		if c.Style != current {
			if styled {
				sb.WriteString("<</>>")
			}
			if !c.Style.IsZero() {
				sb.WriteString("<<")
				sb.WriteString(formatStyleAttrs(c.Style))
				sb.WriteString(">>")
				styled = true
			} else {
				styled = false
			}
			current = c.Style
		}

		sb.WriteRune(ch)
	}

	if styled {
		sb.WriteString("<</>>")
	}

	return sb.String()
}

func formatStyleAttrs(s Style) string {
	var attrs []string

	if s.BG.Name != "" {
		attrs = append(attrs, "bg:"+s.BG.Name)
	}
	if s.FG.Name != "" {
		attrs = append(attrs, s.FG.Name)
	}

	var bools []string
	if s.Bold {
		bools = append(bools, "bold")
	}
	if s.Dim {
		bools = append(bools, "dim")
	}
	if s.Inverse {
		bools = append(bools, "inverse")
	}
	if s.Italic {
		bools = append(bools, "italic")
	}
	if s.Strikethrough {
		bools = append(bools, "strikethrough")
	}
	if s.Underline {
		bools = append(bools, "underline")
	}
	attrs = append(attrs, bools...)

	return strings.Join(attrs, ",")
}
