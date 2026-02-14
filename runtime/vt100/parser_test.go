package vt100_test

import (
	"bytes"
	"testing"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/vt100"
)

func TestParsePlainTextRoundTrip(t *testing.T) {
	// Given — a buffer with plain text at various positions
	buf := render.NewBuffer(10, 3)
	buf.WriteText(0, 0, "Hello")
	buf.WriteText(1, 2, "World")
	buf.WriteText(2, 5, "!")

	// When — render to ANSI bytes and parse back
	var ansi bytes.Buffer
	buf.RenderTo(&ansi)

	screen := vt100.NewScreen(10, 3)
	_, err := screen.Write(ansi.Bytes())

	// Then — parsed cells match original
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertCellsMatch(t, buf, screen)
}

func TestParseCursorHome(t *testing.T) {
	// Given — ESC[H moves cursor to (0,0), then text is written
	screen := vt100.NewScreen(5, 2)

	// When — write CUP to (1,3), write 'A', then home, write 'B'
	_, err := screen.Write([]byte("\x1b[1;3HA\x1b[HB"))

	// Then — 'A' at (0,2), 'B' at (0,0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := screen.Cell(0, 2).Ch; got != 'A' {
		t.Errorf("cell (0,2): got %q, want 'A'", got)
	}
	if got := screen.Cell(0, 0).Ch; got != 'B' {
		t.Errorf("cell (0,0): got %q, want 'B'", got)
	}
}

func TestParseClearScreen(t *testing.T) {
	// Given — screen with existing content
	screen := vt100.NewScreen(5, 2)
	screen.Write([]byte("\x1b[1;1HHello"))

	// When — clear screen
	screen.Write([]byte("\x1b[2J"))

	// Then — all cells are empty
	for row := 0; row < 2; row++ {
		for col := 0; col < 5; col++ {
			if got := screen.Cell(row, col).Ch; got != 0 {
				t.Errorf("cell (%d,%d): got %q, want 0", row, col, got)
			}
		}
	}
}

func TestParseTextAdvancesCursor(t *testing.T) {
	// Given — cursor at a position, write multiple characters
	screen := vt100.NewScreen(10, 1)

	// When — position cursor at (0,0) and write "AB"
	screen.Write([]byte("\x1b[1;1HAB"))

	// Then — 'A' at (0,0), 'B' at (0,1)
	if got := screen.Cell(0, 0).Ch; got != 'A' {
		t.Errorf("cell (0,0): got %q, want 'A'", got)
	}
	if got := screen.Cell(0, 1).Ch; got != 'B' {
		t.Errorf("cell (0,1): got %q, want 'B'", got)
	}
}

func TestParseTextWrapsToNowhere(t *testing.T) {
	// Given — cursor near end of line, text would overflow
	screen := vt100.NewScreen(3, 1)

	// When — write text that goes past the edge
	screen.Write([]byte("\x1b[1;2HAB"))

	// Then — 'A' at (0,1), 'B' discarded (cursor stops at width)
	if got := screen.Cell(0, 1).Ch; got != 'A' {
		t.Errorf("cell (0,1): got %q, want 'A'", got)
	}
	if got := screen.Cell(0, 2).Ch; got != 'B' {
		t.Errorf("cell (0,2): got %q, want 'B'", got)
	}
}

// assertCellsMatch compares two buffers cell-by-cell.
func assertCellsMatch(t *testing.T, expected *render.Buffer, screen *vt100.Screen) {
	t.Helper()
	for row := 0; row < expected.Height(); row++ {
		for col := 0; col < expected.Width(); col++ {
			exp := expected.Cell(row, col)
			got := screen.Cell(row, col)
			if exp.Ch != got.Ch {
				t.Errorf("char mismatch at (%d,%d): got %q, want %q", row, col, got.Ch, exp.Ch)
			}
		}
	}
}
