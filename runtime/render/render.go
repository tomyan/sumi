package render

import (
	"fmt"
	"io"
)

// RenderTo renders the buffer to a writer using ANSI cursor-addressed output.
// Only non-empty cells (Ch != 0) are rendered. Styled cells emit ANSI SGR
// sequences before the character.
func (b *Buffer) RenderTo(w io.Writer) {
	hasStyled := false
	for row := 0; row < b.height; row++ {
		for col := 0; col < b.width; col++ {
			c := b.cells[row][col]
			if c.Ch == 0 {
				continue
			}
			fmt.Fprintf(w, "\x1b[%d;%dH", row+1, col+1)
			if !c.Style.IsZero() {
				hasStyled = true
				fmt.Fprint(w, "\x1b[0m")
				if c.Style.Bold {
					fmt.Fprint(w, "\x1b[1m")
				}
				if c.Style.Dim {
					fmt.Fprint(w, "\x1b[2m")
				}
				if c.Style.Italic {
					fmt.Fprint(w, "\x1b[3m")
				}
				if c.Style.Underline {
					fmt.Fprint(w, "\x1b[4m")
				}
				if c.Style.Inverse {
					fmt.Fprint(w, "\x1b[7m")
				}
				if c.Style.Strikethrough {
					fmt.Fprint(w, "\x1b[9m")
				}
				if code, ok := colorToFGCode(c.Style.FG.Name); ok {
					fmt.Fprintf(w, "\x1b[%dm", code)
				}
				if code, ok := colorToBGCode(c.Style.BG.Name); ok {
					fmt.Fprintf(w, "\x1b[%dm", code)
				}
			}
			fmt.Fprintf(w, "%c", c.Ch)
		}
	}
	if hasStyled {
		fmt.Fprint(w, "\x1b[0m")
	}
}
