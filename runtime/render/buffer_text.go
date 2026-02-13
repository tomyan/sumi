package render

import "strings"

// ToPlainText converts the buffer to a plain text string.
// Empty cells (Ch==0) become spaces. Trailing spaces per row and
// trailing empty rows are trimmed. Rows are joined with newlines.
func (b *Buffer) ToPlainText() string {
	rows := make([]string, b.height)
	for r := 0; r < b.height; r++ {
		rows[r] = rowToPlainText(b.cells[r])
	}
	return trimTrailingEmptyRows(rows)
}

func rowToPlainText(cells []Cell) string {
	var sb strings.Builder
	for _, c := range cells {
		if c.Ch == 0 {
			sb.WriteByte(' ')
		} else {
			sb.WriteRune(c.Ch)
		}
	}
	return strings.TrimRight(sb.String(), " ")
}

func trimTrailingEmptyRows(rows []string) string {
	last := len(rows) - 1
	for last >= 0 && rows[last] == "" {
		last--
	}
	if last < 0 {
		return ""
	}
	return strings.Join(rows[:last+1], "\n")
}
