package render

import (
	"fmt"
	"io"
)

// RenderTo renders the buffer to a writer using ANSI cursor-addressed output.
// Only non-empty cells (Ch != 0) are rendered.
func (b *Buffer) RenderTo(w io.Writer) {
	for row := 0; row < b.height; row++ {
		for col := 0; col < b.width; col++ {
			c := b.cells[row][col]
			if c.Ch != 0 {
				fmt.Fprintf(w, "\x1b[%d;%dH%c", row+1, col+1, c.Ch)
			}
		}
	}
}
