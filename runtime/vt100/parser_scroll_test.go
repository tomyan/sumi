package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/vt100"
)

func TestScrollRegionSetAndReset(t *testing.T) {
	// Given — 10 row screen, set scroll region to rows 2-5 (1-based)
	screen := vt100.NewScreen(10, 10)
	screen.Write([]byte("AAAAAAAAAA\r\nBBBBBBBBBB\r\nCCCCCCCCCC\r\nDDDDDDDDDD\r\nEEEEEEEEEE\r\nFFFFFFFFFF"))

	// When — set scroll region rows 3-5 and scroll up via LF at bottom of region
	screen.Write([]byte("\x1b[3;5r"))       // set region
	screen.Write([]byte("\x1b[5;1H"))       // move to row 5 (bottom of region)
	screen.Write([]byte("\n"))              // LF at bottom of region → scroll within region

	// Then — rows outside region unchanged
	if cellStr(screen, 0, 3) != "AAA" {
		t.Errorf("row 0 = %q, want AAA (outside region)", cellStr(screen, 0, 3))
	}
	if cellStr(screen, 1, 3) != "BBB" {
		t.Errorf("row 1 = %q, want BBB (outside region)", cellStr(screen, 1, 3))
	}
	// Row 2 (was C) should now have D (scrolled up within region)
	if cellStr(screen, 2, 3) != "DDD" {
		t.Errorf("row 2 = %q, want DDD (scrolled up)", cellStr(screen, 2, 3))
	}
	// Row 3 (was D) should now have E
	if cellStr(screen, 3, 3) != "EEE" {
		t.Errorf("row 3 = %q, want EEE (scrolled up)", cellStr(screen, 3, 3))
	}
	// Row 4 (bottom of region) should be blank
	if cellStr(screen, 4, 3) != "   " {
		t.Errorf("row 4 = %q, want blank (new line)", cellStr(screen, 4, 3))
	}
	// Row 5 (outside region) should be unchanged
	if cellStr(screen, 5, 3) != "FFF" {
		t.Errorf("row 5 = %q, want FFF (outside region)", cellStr(screen, 5, 3))
	}
}

func TestScrollRegionReverseIndex(t *testing.T) {
	// Given — set scroll region rows 2-4 (1-based), cursor at top of region
	screen := vt100.NewScreen(10, 6)
	screen.Write([]byte("AAA\r\nBBB\r\nCCC\r\nDDD\r\nEEE\r\nFFF"))
	screen.Write([]byte("\x1b[2;4r"))       // set region rows 2-4
	screen.Write([]byte("\x1b[2;1H"))       // cursor at top of region (row 2)

	// When — reverse index at top of region → scroll down within region
	screen.Write([]byte("\x1bM"))

	// Then — row 0 unchanged
	if cellStr(screen, 0, 3) != "AAA" {
		t.Errorf("row 0 = %q, want AAA", cellStr(screen, 0, 3))
	}
	// Row 1 (top of region) should be blank (new line from scroll down)
	if cellStr(screen, 1, 3) != "   " {
		t.Errorf("row 1 = %q, want blank", cellStr(screen, 1, 3))
	}
	// Row 2 should have BBB (was row 1, pushed down)
	if cellStr(screen, 2, 3) != "BBB" {
		t.Errorf("row 2 = %q, want BBB", cellStr(screen, 2, 3))
	}
	// Row 3 should have CCC (was row 2, pushed down)
	if cellStr(screen, 3, 3) != "CCC" {
		t.Errorf("row 3 = %q, want CCC", cellStr(screen, 3, 3))
	}
	// Row 4 unchanged (outside region bottom)
	if cellStr(screen, 4, 3) != "EEE" {
		t.Errorf("row 4 = %q, want EEE", cellStr(screen, 4, 3))
	}
}

func TestScrollRegionResetWithNoParams(t *testing.T) {
	// Given — set a scroll region, then reset it
	screen := vt100.NewScreen(10, 5)
	screen.Write([]byte("AAA\r\nBBB\r\nCCC\r\nDDD\r\nEEE"))
	screen.Write([]byte("\x1b[2;3r"))       // set region
	screen.Write([]byte("\x1b[r"))          // reset to full screen

	// When — cursor at bottom, LF should scroll the full screen
	screen.Write([]byte("\x1b[5;1H\n"))

	// Then — row 0 should have BBB (full-screen scroll)
	if cellStr(screen, 0, 3) != "BBB" {
		t.Errorf("row 0 = %q, want BBB (full screen scroll)", cellStr(screen, 0, 3))
	}
}
