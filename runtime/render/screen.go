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
