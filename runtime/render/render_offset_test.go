package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderToOffsetShiftsCursorPositions(t *testing.T) {
	// Given — a 3x2 buffer with text at (0,0) and (1,0)
	buf := NewBuffer(3, 2)
	buf.WriteText(0, 0, "Hi!")
	buf.WriteText(1, 0, "Go")

	// When — render with offset (5, 3) → row+5, col+3
	var out bytes.Buffer
	buf.RenderToOffset(&out, 5, 3)

	// Then — cursor should be at row 6, col 4 for first cell (0-based + offset + 1-based)
	result := out.String()
	// (0,0) → row 0+5+1=6, col 0+3+1=4 → \x1b[6;4H
	if !strings.Contains(result, "\x1b[6;4H") {
		t.Errorf("expected cursor at [6;4H, got:\n%s", result)
	}
	// (1,0) → row 1+5+1=7, col 0+3+1=4 → \x1b[7;4H
	if !strings.Contains(result, "\x1b[7;4H") {
		t.Errorf("expected cursor at [7;4H, got:\n%s", result)
	}
}

func TestRenderToOffsetPreservesCharacters(t *testing.T) {
	// Given
	buf := NewBuffer(5, 1)
	buf.WriteText(0, 0, "Hello")

	// When
	var out bytes.Buffer
	buf.RenderToOffset(&out, 0, 0)

	// Then — characters appear in output
	result := out.String()
	if !strings.Contains(result, "H") || !strings.Contains(result, "o") {
		t.Errorf("expected characters in output, got:\n%s", result)
	}
}

func TestRenderToOffsetPreservesStyles(t *testing.T) {
	// Given
	buf := NewBuffer(3, 1)
	buf.WriteStyledText(0, 0, "abc", Style{FG: Color{Name: "green"}})

	// When
	var out bytes.Buffer
	buf.RenderToOffset(&out, 2, 1)

	// Then — should contain SGR for green (code 32) and shifted cursor
	result := out.String()
	if !strings.Contains(result, ";32m") {
		t.Errorf("expected green SGR code, got:\n%s", result)
	}
	// (0,0) → row 0+2+1=3, col 0+1+1=2 → \x1b[3;2H
	if !strings.Contains(result, "\x1b[3;2H") {
		t.Errorf("expected cursor at [3;2H, got:\n%s", result)
	}
}

func TestRenderToOffsetZeroOffsetMatchesRenderTo(t *testing.T) {
	// Given
	buf := NewBuffer(4, 2)
	buf.WriteText(0, 0, "test")
	buf.WriteText(1, 0, "data")

	// When
	var normal, offset bytes.Buffer
	buf.RenderTo(&normal)
	buf.RenderToOffset(&offset, 0, 0)

	// Then — output should be identical
	if normal.String() != offset.String() {
		t.Errorf("zero-offset should match RenderTo:\nnormal: %q\noffset: %q", normal.String(), offset.String())
	}
}

func TestRenderToOffsetSkipsEmptyCells(t *testing.T) {
	// Given — a 5x1 buffer with only position 2 filled
	buf := NewBuffer(5, 1)
	buf.SetCell(0, 2, 'X')

	// When
	var out bytes.Buffer
	buf.RenderToOffset(&out, 10, 5)

	// Then — only one cursor movement, for the X at (0,2) → [11;8H
	result := out.String()
	if !strings.Contains(result, "\x1b[11;8H") {
		t.Errorf("expected cursor at [11;8H, got:\n%s", result)
	}
	if strings.Count(result, "\x1b[") > 2 { // one cursor + possibly one reset
		t.Errorf("expected minimal cursor movements for sparse buffer, got:\n%s", result)
	}
}
