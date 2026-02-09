package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmptyBufferRendersNothing(t *testing.T) {
	b := NewBuffer(10, 5)
	var buf bytes.Buffer
	b.RenderTo(&buf)

	if buf.Len() != 0 {
		t.Errorf("empty buffer rendered %d bytes, want 0", buf.Len())
	}
}

func TestRenderSingleCell(t *testing.T) {
	b := NewBuffer(10, 5)
	b.SetCell(0, 0, 'A')

	var buf bytes.Buffer
	b.RenderTo(&buf)

	got := buf.String()
	// Should contain cursor move to row 1, col 1 (1-indexed) and the char 'A'
	if !strings.Contains(got, "\x1b[1;1H") {
		t.Errorf("output missing cursor move \\x1b[1;1H, got %q", got)
	}
	if !strings.Contains(got, "A") {
		t.Errorf("output missing character A, got %q", got)
	}
}

func TestRenderTextHi(t *testing.T) {
	b := NewBuffer(10, 5)
	b.WriteText(0, 0, "Hi")

	var buf bytes.Buffer
	b.RenderTo(&buf)

	got := buf.String()
	// Should contain cursor moves and both characters
	if !strings.Contains(got, "\x1b[1;1H") {
		t.Errorf("output missing cursor move for H, got %q", got)
	}
	if !strings.Contains(got, "H") {
		t.Errorf("output missing character H, got %q", got)
	}
	if !strings.Contains(got, "i") {
		t.Errorf("output missing character i, got %q", got)
	}
}

func TestRenderSparseBuffer(t *testing.T) {
	b := NewBuffer(80, 24)
	b.SetCell(0, 0, 'A')
	b.SetCell(10, 40, 'B')
	b.SetCell(23, 79, 'C')

	var buf bytes.Buffer
	b.RenderTo(&buf)

	got := buf.String()

	// Should contain cursor moves for all three cells (1-indexed)
	if !strings.Contains(got, "\x1b[1;1H") {
		t.Errorf("missing cursor move for A at (0,0)")
	}
	if !strings.Contains(got, "\x1b[11;41H") {
		t.Errorf("missing cursor move for B at (10,40)")
	}
	if !strings.Contains(got, "\x1b[24;80H") {
		t.Errorf("missing cursor move for C at (23,79)")
	}
	if !strings.Contains(got, "A") || !strings.Contains(got, "B") || !strings.Contains(got, "C") {
		t.Errorf("missing characters, got %q", got)
	}
}
