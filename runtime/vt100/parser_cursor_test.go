package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestCSISaveRestoreCursor(t *testing.T) {
	// Given — position cursor, save with CSI s
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[3;7H\x1b[s"))

	// When — move elsewhere, restore with CSI u, write
	screen.Write([]byte("\x1b[1;1H\x1b[uX"))

	// Then — X at row 2, col 6
	cell := screen.Cell(2, 6)
	if cell.Ch != 'X' {
		t.Errorf("Cell(2,6).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestCSISaveRestoreSharesWithDECSC(t *testing.T) {
	// Given — save with DECSC at one position
	screen := vt100.NewScreen(20, 5)
	screen.Write([]byte("\x1b[2;3H\x1b7"))

	// When — restore with CSI u
	screen.Write([]byte("\x1b[1;1H\x1b[uX"))

	// Then — X at the saved position (row 1, col 2)
	cell := screen.Cell(1, 2)
	if cell.Ch != 'X' {
		t.Errorf("Cell(1,2).Ch = %c, want 'X'", cell.Ch)
	}
}
