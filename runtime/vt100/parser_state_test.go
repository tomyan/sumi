package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestSplitWriteCSIAcrossCalls(t *testing.T) {
	// Given — a CUP sequence ESC[3;5H split across two Write calls
	screen := vt100.NewScreen(20, 10)

	// When — first call has ESC[, second has 3;5H then a character
	screen.Write([]byte("\x1b["))
	screen.Write([]byte("3;5HX"))

	// Then — X should be at row 2, col 4 (1-based → 0-based)
	cell := screen.Cell(2, 4)
	if cell.Ch != 'X' {
		t.Errorf("Cell(2,4).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestSplitWriteESCAcrossCalls(t *testing.T) {
	// Given — ESC arrives alone, then [ and rest in next call
	screen := vt100.NewScreen(20, 10)

	// When
	screen.Write([]byte("\x1b"))
	screen.Write([]byte("[2;1HA"))

	// Then — A at row 1, col 0
	cell := screen.Cell(1, 0)
	if cell.Ch != 'A' {
		t.Errorf("Cell(1,0).Ch = %c, want 'A'", cell.Ch)
	}
}

func TestSplitWriteOSCAcrossCalls(t *testing.T) {
	// Given — OSC sentinel split: first part has ESC]999;do, second has ne\x07
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b]999;do"))
	screen.Write([]byte("ne\x07"))

	// Then
	if !screen.SentinelSeen() {
		t.Error("SentinelSeen() = false, want true after split OSC")
	}
}

func TestSplitWriteSGRAcrossCalls(t *testing.T) {
	// Given — SGR sequence ESC[1m split: ESC[1 then m
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[1"))
	screen.Write([]byte("mX"))

	// Then — X should be bold
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
	if !cell.Style.Bold {
		t.Error("Cell(0,0).Style.Bold = false, want true")
	}
}

func TestSplitWriteCharsetDesignation(t *testing.T) {
	// Given — ESC( split from B
	screen := vt100.NewScreen(20, 5)

	// When — ESC( in first call, B then text in second
	screen.Write([]byte("\x1b("))
	screen.Write([]byte("BX"))

	// Then — X should be at (0,0), B should not appear
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}

func TestSplitWritePrivateMode(t *testing.T) {
	// Given — ESC[?25 split from h
	screen := vt100.NewScreen(20, 5)

	// When
	screen.Write([]byte("\x1b[?25"))
	screen.Write([]byte("hX"))

	// Then — private mode consumed, X at (0,0)
	cell := screen.Cell(0, 0)
	if cell.Ch != 'X' {
		t.Errorf("Cell(0,0).Ch = %c, want 'X'", cell.Ch)
	}
}
