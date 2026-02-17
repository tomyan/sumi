package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestResizeChangesDimensions(t *testing.T) {
	// Given
	screen := vt100.NewScreen(40, 20)

	// When
	screen.Resize(80, 24)

	// Then
	if screen.Width() != 80 {
		t.Errorf("Width() = %d, want 80", screen.Width())
	}
	if screen.Height() != 24 {
		t.Errorf("Height() = %d, want 24", screen.Height())
	}
}

func TestResizeClearsContent(t *testing.T) {
	// Given — write some content to the screen
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("Hello"))

	// When
	screen.Resize(30, 10)

	// Then — all cells should be empty
	for row := 0; row < screen.Height(); row++ {
		for col := 0; col < screen.Width(); col++ {
			cell := screen.Cell(row, col)
			if cell.Ch != 0 {
				t.Errorf("Cell(%d,%d).Ch = %c, want 0", row, col, cell.Ch)
			}
		}
	}
}

func TestResizeResetsCursor(t *testing.T) {
	// Given — move cursor via write
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("ABC"))

	// When
	screen.Resize(10, 3)

	// Then — writing to the resized screen starts at (0,0)
	screen.Write([]byte("X"))
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}
