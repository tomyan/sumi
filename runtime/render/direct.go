package render

import (
	"fmt"
	"io"
	"strings"
)

// WriteAt moves the cursor to (row, col) and writes text with the given style.
// Coordinates are 0-based; ANSI cursor positioning is 1-based.
func WriteAt(w io.Writer, row, col int, text string, style Style) {
	fmt.Fprintf(w, "\x1b[%d;%dH", row+1, col+1)
	if sgr := buildSGR(style); sgr != "" {
		fmt.Fprint(w, sgr)
		fmt.Fprint(w, text)
		fmt.Fprint(w, "\x1b[0m")
	} else {
		fmt.Fprint(w, text)
	}
}

// ClearRegion fills a rectangular region with spaces using cursor positioning.
// Coordinates are 0-based.
func ClearRegion(w io.Writer, row, col, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	spaces := strings.Repeat(" ", width)
	for r := 0; r < height; r++ {
		fmt.Fprintf(w, "\x1b[%d;%dH%s", row+r+1, col+1, spaces)
	}
}

// DrawBorderAt draws a border directly to a writer using cursor positioning.
// Coordinates are 0-based. borderStyle "" or "none" is a no-op.
func DrawBorderAt(w io.Writer, row, col, width, height int, borderStyle string, style Style) {
	if borderStyle == "" || borderStyle == "none" {
		return
	}
	if width < 2 || height < 2 {
		return
	}

	right := col + width - 1
	bottom := row + height - 1

	writeCharAt(w, row, col, '┌', style)
	writeCharAt(w, row, right, '┐', style)
	writeCharAt(w, bottom, col, '└', style)
	writeCharAt(w, bottom, right, '┘', style)

	for c := col + 1; c < right; c++ {
		writeCharAt(w, row, c, '─', style)
		writeCharAt(w, bottom, c, '─', style)
	}
	for r := row + 1; r < bottom; r++ {
		writeCharAt(w, r, col, '│', style)
		writeCharAt(w, r, right, '│', style)
	}
}

// DrawBorderTitleAt draws a border title directly to a writer using cursor positioning.
func DrawBorderTitleAt(w io.Writer, row, col, width int, title string, style Style) {
	if title == "" || width < 6 {
		return
	}
	maxLen := width - 4
	runes := []rune(title)
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	writeCharAt(w, row, col+2, ' ', style)
	for i, ch := range runes {
		writeCharAt(w, row, col+3+i, ch, style)
	}
	writeCharAt(w, row, col+3+len(runes), ' ', style)
}

// writeCharAt writes a single character at (row, col) with style.
func writeCharAt(w io.Writer, row, col int, ch rune, style Style) {
	fmt.Fprintf(w, "\x1b[%d;%dH", row+1, col+1)
	if sgr := buildSGR(style); sgr != "" {
		fmt.Fprintf(w, "%s%c\x1b[0m", sgr, ch)
	} else {
		fmt.Fprintf(w, "%c", ch)
	}
}
