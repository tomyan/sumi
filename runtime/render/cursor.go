package render

import (
	"fmt"
	"io"
)

// ShowCursor makes the terminal cursor visible and moves it to the given position.
// Coordinates are 0-based; ANSI cursor positioning is 1-based.
func ShowCursor(w io.Writer, row, col int) {
	fmt.Fprintf(w, "\x1b[?25h\x1b[%d;%dH", row+1, col+1)
}

// HideCursor hides the terminal cursor.
func HideCursor(w io.Writer) {
	fmt.Fprint(w, "\x1b[?25l")
}
