package render

import "io"

// EnterAlternateScreen switches to the alternate screen buffer and hides the cursor.
func EnterAlternateScreen(w io.Writer) {
	io.WriteString(w, "\x1b[?1049h\x1b[?25l")
}

// ExitAlternateScreen shows the cursor and returns to the normal screen buffer.
func ExitAlternateScreen(w io.Writer) {
	io.WriteString(w, "\x1b[?25h\x1b[?1049l")
}

// ClearScreen erases all content and moves the cursor to the top-left.
func ClearScreen(w io.Writer) {
	io.WriteString(w, "\x1b[2J\x1b[H")
}
