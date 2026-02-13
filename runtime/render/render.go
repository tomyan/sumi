package render

import (
	"fmt"
	"io"
	"strings"
)

// RenderTo renders the buffer to a writer using ANSI cursor-addressed output.
// Only non-empty cells (Ch != 0) are rendered. Styled cells emit ANSI SGR
// sequences before the character.
func (b *Buffer) RenderTo(w io.Writer) {
	b.RenderToOffset(w, 0, 0)
}

// RenderToOffset renders the buffer with all cursor positions shifted by the
// given row and column offsets. This allows rendering a buffer at an arbitrary
// position on screen (e.g. below a header in preview mode).
func (b *Buffer) RenderToOffset(w io.Writer, rowOffset, colOffset int) {
	inStyled := false
	for row := 0; row < b.height; row++ {
		for col := 0; col < b.width; col++ {
			c := b.cells[row][col]
			if c.Ch == 0 {
				continue
			}
			fmt.Fprintf(w, "\x1b[%d;%dH", row+1+rowOffset, col+1+colOffset)
			if sgr := buildSGR(c.Style); sgr != "" {
				inStyled = true
				fmt.Fprint(w, sgr)
			} else if inStyled {
				fmt.Fprint(w, "\x1b[0m")
				inStyled = false
			}
			fmt.Fprintf(w, "%c", c.Ch)
		}
	}
	if inStyled {
		fmt.Fprint(w, "\x1b[0m")
	}
}

// buildSGR returns the ANSI SGR escape sequence for a style.
// Returns an empty string if the style has no attributes set.
func buildSGR(s Style) string {
	if s.IsZero() {
		return ""
	}
	var b strings.Builder
	b.WriteString("\x1b[0m")
	appendAttrCodes(&b, s)
	appendColorCodes(&b, s)
	return b.String()
}

// appendAttrCodes appends SGR codes for text attributes (bold, dim, etc.).
func appendAttrCodes(b *strings.Builder, s Style) {
	attrs := []struct {
		set  bool
		code int
	}{
		{s.Bold, 1},
		{s.Dim, 2},
		{s.Italic, 3},
		{s.Underline, 4},
		{s.Inverse, 7},
		{s.Strikethrough, 9},
	}
	for _, a := range attrs {
		if a.set {
			fmt.Fprintf(b, "\x1b[%dm", a.code)
		}
	}
}

// appendColorCodes appends SGR codes for foreground and background colors.
func appendColorCodes(b *strings.Builder, s Style) {
	if code, ok := colorToFGCode(s.FG.Name); ok {
		fmt.Fprintf(b, "\x1b[%dm", code)
	}
	if code, ok := colorToBGCode(s.BG.Name); ok {
		fmt.Fprintf(b, "\x1b[%dm", code)
	}
}
