package render

import (
	"bytes"
	"testing"
)

func TestWriteAtPlainText(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	WriteAt(&buf, 0, 0, "hi", Style{})

	// Then — cursor moves to (1,1) in ANSI coords, writes "hi"
	got := buf.String()
	want := "\x1b[1;1Hhi"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteAtStyledText(t *testing.T) {
	// Given
	var buf bytes.Buffer
	style := Style{FG: Color{Name: "red"}}

	// When
	WriteAt(&buf, 2, 5, "ok", style)

	// Then — cursor to (3,6), SGR for red, text, reset
	got := buf.String()
	want := "\x1b[3;6H\x1b[0m\x1b[31mok\x1b[0m"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestClearRegion(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	ClearRegion(&buf, 1, 2, 3, 2)

	// Then — fills 2 rows × 3 cols with spaces starting at (2,3)
	got := buf.String()
	want := "\x1b[2;3H   \x1b[3;3H   "
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestClearRegionZeroSize(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	ClearRegion(&buf, 0, 0, 0, 0)

	// Then — nothing written
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestDrawBorderAtSingle(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	DrawBorderAt(&buf, 0, 0, 3, 3, "single", Style{})

	// Then — draws a 3×3 border with corners and edges
	got := buf.String()
	// Top: ┌─┐
	// Mid: │ │
	// Bot: └─┘
	if !containsAll(got, "┌", "─", "┐", "│", "└", "┘") {
		t.Errorf("border missing expected characters: %q", got)
	}
}

func TestDrawBorderAtStyled(t *testing.T) {
	// Given
	var buf bytes.Buffer
	style := Style{FG: Color{Name: "cyan"}}

	// When
	DrawBorderAt(&buf, 0, 0, 3, 3, "single", style)

	// Then — should contain SGR codes for cyan
	got := buf.String()
	if !contains(got, "\x1b[36m") {
		t.Errorf("expected cyan SGR code in output: %q", got)
	}
}

func TestDrawBorderAtNone(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	DrawBorderAt(&buf, 0, 0, 3, 3, "none", Style{})

	// Then — nothing written
	if buf.Len() != 0 {
		t.Errorf("expected empty output for none border, got %q", buf.String())
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && bytes.Contains([]byte(s), []byte(sub))
}
