package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderBoldCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'A', Style{Bold: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[1m") {
		t.Errorf("output missing bold SGR \\x1b[1m, got %q", got)
	}
	if !strings.Contains(got, "A") {
		t.Errorf("output missing character A, got %q", got)
	}
}

func TestRenderGreenFG(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'G', Style{FG: Color{Name: "green"}})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[32m") {
		t.Errorf("output missing green FG SGR \\x1b[32m, got %q", got)
	}
}

func TestRenderBGColor(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'B', Style{BG: Color{Name: "blue"}})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[44m") {
		t.Errorf("output missing blue BG SGR \\x1b[44m, got %q", got)
	}
}

func TestRenderMultipleAttributes(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	s := Style{
		FG:        Color{Name: "cyan"},
		BG:        Color{Name: "yellow"},
		Bold:      true,
		Underline: true,
	}
	b.SetStyledCell(0, 0, 'M', s)
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should contain reset, bold, underline, cyan FG, yellow BG
	if !strings.Contains(got, "\x1b[0m") {
		t.Errorf("output missing reset SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[1m") {
		t.Errorf("output missing bold SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[4m") {
		t.Errorf("output missing underline SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[36m") {
		t.Errorf("output missing cyan FG SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[43m") {
		t.Errorf("output missing yellow BG SGR, got %q", got)
	}
}

func TestRenderUnstyledCellNoSGR(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetCell(0, 0, 'A') // no style, uses existing method
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should NOT contain any SGR escape
	if strings.Contains(got, "\x1b[") && !strings.Contains(got, "\x1b[1;1H") {
		// Allow cursor positioning escapes but no SGR
		// More precise: strip cursor moves and check for remaining escapes
	}
	// The output should be exactly the cursor move + char, no SGR codes
	expected := "\x1b[1;1HA"
	if got != expected {
		t.Errorf("unstyled cell output = %q, want %q", got, expected)
	}
}

func TestRenderDimCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'D', Style{Dim: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[2m") {
		t.Errorf("output missing dim SGR \\x1b[2m, got %q", got)
	}
}

func TestRenderItalicCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'I', Style{Italic: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[3m") {
		t.Errorf("output missing italic SGR \\x1b[3m, got %q", got)
	}
}

func TestRenderStrikethroughCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'S', Style{Strikethrough: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[9m") {
		t.Errorf("output missing strikethrough SGR \\x1b[9m, got %q", got)
	}
}

func TestRenderInverseCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'V', Style{Inverse: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[7m") {
		t.Errorf("output missing inverse SGR \\x1b[7m, got %q", got)
	}
}

func TestRenderTrailingReset(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'A', Style{Bold: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should end with a reset
	if !strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("output should end with reset SGR, got %q", got)
	}
}

func TestRenderNoTrailingResetForUnstyled(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetCell(0, 0, 'A') // unstyled
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should NOT end with reset when no styled cells exist
	if strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("unstyled output should not end with reset, got %q", got)
	}
}
