package render

import (
	"bytes"
	"testing"
)

func TestShowCursorEmitsANSI(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	ShowCursor(&buf, 5, 3)

	// Then — expects ESC[?25h (show cursor) then ESC[row;colH (position)
	// row/col are 0-based input, 1-based ANSI output
	expected := "\x1b[?25h\x1b[6;4H"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestHideCursorEmitsANSI(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	HideCursor(&buf)

	// Then
	expected := "\x1b[?25l"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestShowCursorAtOrigin(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	ShowCursor(&buf, 0, 0)

	// Then
	expected := "\x1b[?25h\x1b[1;1H"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}
