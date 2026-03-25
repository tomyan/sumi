package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestCSIWithIntermediateSpace(t *testing.T) {
	// Given — ESC[2 q (cursor shape) should be consumed without corruption
	screen := vt100.NewScreen(20, 5)

	// When — cursor shape then a character
	screen.Write([]byte("\x1b[2 qX"))

	// Then — X at (0,0), the space+q sequence consumed cleanly
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestCSIWithGreaterThanPrefix(t *testing.T) {
	// Given — ESC[>c (secondary device attributes) should be consumed
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[>cX"))

	// Then — X at (0,0)
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestCSIWithEqualsPrefix(t *testing.T) {
	// Given — ESC[=c (tertiary device attributes) should be consumed
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[=cX"))

	// Then — X at (0,0)
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestCSIWithDollarIntermediate(t *testing.T) {
	// Given — ESC[0$p (DECRQM) should be consumed
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[0$pX"))

	// Then — X at (0,0)
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestCSIWithExclamationIntermediate(t *testing.T) {
	// Given — ESC[!p (DECSTR soft reset) should be consumed
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[!pX"))

	// Then — X at (0,0)
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestCSIIntermediateDoesNotCorruptFollowing(t *testing.T) {
	// Given — unknown CSI followed by a real CUP
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[2 q\x1b[2;3HY"))

	// Then — Y at row 1, col 2
	cell := screen.Cell(1, 2)
	if cell.Ch != 'Y' {
		t.Errorf("Cell(1,2).Ch = %c, want 'Y'", cell.Ch)
	}
}
