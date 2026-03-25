package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestESCMReverseIndexScrollsDown(t *testing.T) {
	// Given — cursor at row 0, screen has content
	screen := vt100.NewScreen(10, 3)
	screen.Write([]byte("AAA\r\nBBB\r\nCCC"))

	// When — move to row 0 then reverse index
	screen.Write([]byte("\x1b[1;1H\x1bM"))

	// Then — row 0 should be blank (scrolled down), row 1 should have AAA
	if cellStr(screen, 0, 3) != "   " {
		t.Errorf("row 0 = %q, want blank", cellStr(screen, 0, 3))
	}
	if cellStr(screen, 1, 3) != "AAA" {
		t.Errorf("row 1 = %q, want AAA", cellStr(screen, 1, 3))
	}
}

func TestESCMReverseIndexNonTopRow(t *testing.T) {
	// Given — cursor at row 2
	screen := vt100.NewScreen(10, 5)
	screen.Write([]byte("\x1b[3;1H"))

	// When — reverse index from non-top row just moves up
	screen.Write([]byte("\x1bMX"))

	// Then — X at row 1
	cell := screen.Cell(1, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(1,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestESCDIndexScrollsUp(t *testing.T) {
	// Given — cursor at bottom row, screen has content
	screen := vt100.NewScreen(10, 3)
	screen.Write([]byte("AAA\r\nBBB\r\nCCC"))

	// When — index (cursor down) at bottom row
	screen.Write([]byte("\x1bD"))

	// Then — scrolled up: row 0 = BBB, row 1 = CCC, row 2 = blank
	if cellStr(screen, 0, 3) != "BBB" {
		t.Errorf("row 0 = %q, want BBB", cellStr(screen, 0, 3))
	}
	if cellStr(screen, 1, 3) != "CCC" {
		t.Errorf("row 1 = %q, want CCC", cellStr(screen, 1, 3))
	}
}

func TestESCENextLine(t *testing.T) {
	// Given
	screen := vt100.NewScreen(10, 5)
	screen.Write([]byte("ABC"))

	// When — ESC E: CR + LF
	screen.Write([]byte("\x1bEX"))

	// Then — X at row 1, col 0
	cell := screen.Cell(1, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(1,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestESC7And8SaveRestoreCursor(t *testing.T) {
	// Given — write at position then save
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[3;5H\x1b7"))

	// When — move elsewhere, then restore
	screen.Write([]byte("\x1b[1;1H\x1b8X"))

	// Then — X at saved position row 2, col 4
	cell := screen.Cell(2, 4)
	if cell.Ch != 'X' {
		t.Errorf("Cell(2,4).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestESCcFullReset(t *testing.T) {
	// Given — styled content on screen
	screen := vt100.NewScreen(10, 3)
	screen.Write([]byte("\x1b[1m\x1b[31mHello"))

	// When — full reset
	screen.Write([]byte("\x1bcX"))

	// Then — screen cleared, cursor at 0,0, style reset
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
	if cell.Style.Bold || cell.Style.FG.Name != "" {
		t.Errorf("Style = %+v, want zero", cell.Style)
	}
	// Old content should be gone
	if screen.Cell(0, 4).Ch != 0 {
		t.Error("old content should be cleared")
	}
}

// cellStr extracts n characters starting at (row, 0).
func cellStr(screen *vt100.Screen, row, n int) string {
	s := make([]byte, n)
	for i := 0; i < n; i++ {
		c := screen.Cell(row, i)
		if c.Ch == 0 {
			s[i] = ' '
		} else {
			s[i] = byte(c.Ch)
		}
	}
	return string(s)
}
