package layout

import (
	"testing"
)

func TestWrapTextWordBreak(t *testing.T) {
	// Given
	text := "hello world"
	width := 6

	// When
	lines := wrapText(text, width)

	// Then
	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(lines))
	}
	if lines[0] != "hello " {
		t.Errorf("lines[0] = %q, want %q", lines[0], "hello ")
	}
	if lines[1] != "world" {
		t.Errorf("lines[1] = %q, want %q", lines[1], "world")
	}
}

func TestWrapTextNoWrapNeeded(t *testing.T) {
	// Given
	text := "short"
	width := 80

	// When
	lines := wrapText(text, width)

	// Then
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	if lines[0] != "short" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "short")
	}
}

func TestWrapTextCharBreakNoSpaces(t *testing.T) {
	// Given — no spaces, must char-break
	text := "abcdef"
	width := 3

	// When
	lines := wrapText(text, width)

	// Then
	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(lines))
	}
	if lines[0] != "abc" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "abc")
	}
	if lines[1] != "def" {
		t.Errorf("lines[1] = %q, want %q", lines[1], "def")
	}
}

func TestWrapTextExactFit(t *testing.T) {
	// Given — text exactly fits width
	text := "hello"
	width := 5

	// When
	lines := wrapText(text, width)

	// Then
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	if lines[0] != "hello" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "hello")
	}
}

func TestWrapTextLongWordFallsBackToCharBreak(t *testing.T) {
	// Given — word longer than width
	text := "abcdefgh ij"
	width := 5

	// When
	lines := wrapText(text, width)

	// Then
	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3", len(lines))
	}
	if lines[0] != "abcde" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "abcde")
	}
	if lines[1] != "fgh " {
		t.Errorf("lines[1] = %q, want %q", lines[1], "fgh ")
	}
	if lines[2] != "ij" {
		t.Errorf("lines[2] = %q, want %q", lines[2], "ij")
	}
}

func TestLayoutTextWrapsWhenExceedsAvailWidth(t *testing.T) {
	// Given
	input := &Input{
		Kind:    KindText,
		Content: "hello world",
	}

	// When
	box := Layout(input, 6, 24)

	// Then
	if box.Width != 6 {
		t.Errorf("Width = %d, want 6", box.Width)
	}
	if box.Height != 2 {
		t.Errorf("Height = %d, want 2", box.Height)
	}
	if len(box.Lines) != 2 {
		t.Fatalf("len(Lines) = %d, want 2", len(box.Lines))
	}
	if box.Lines[0] != "hello " {
		t.Errorf("Lines[0] = %q, want %q", box.Lines[0], "hello ")
	}
	if box.Lines[1] != "world" {
		t.Errorf("Lines[1] = %q, want %q", box.Lines[1], "world")
	}
}

func TestLayoutTextNoWrapWhenFits(t *testing.T) {
	// Given
	input := &Input{
		Kind:    KindText,
		Content: "short",
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Width != 5 {
		t.Errorf("Width = %d, want 5", box.Width)
	}
	if box.Height != 1 {
		t.Errorf("Height = %d, want 1", box.Height)
	}
	if box.Lines != nil {
		t.Errorf("Lines = %v, want nil (no wrapping)", box.Lines)
	}
}

func TestLayoutTextCharBreakNoSpaces(t *testing.T) {
	// Given
	input := &Input{
		Kind:    KindText,
		Content: "abcdef",
	}

	// When
	box := Layout(input, 3, 24)

	// Then
	if box.Height != 2 {
		t.Errorf("Height = %d, want 2", box.Height)
	}
	if box.Width != 3 {
		t.Errorf("Width = %d, want 3", box.Width)
	}
}

func TestLayoutColumnWithWrappedText(t *testing.T) {
	// Given — text wraps inside a column container
	input := &Input{
		Kind:      KindBox,
		Direction: "column",
		Children: []*Input{
			{Kind: KindText, Content: "hello world"},
			{Kind: KindText, Content: "ok"},
		},
	}

	// When
	box := Layout(input, 6, 24)

	// Then — first child wraps to 2 lines, second child starts at Y=2
	first := box.Children[0]
	if first.Height != 2 {
		t.Errorf("first child Height = %d, want 2", first.Height)
	}
	second := box.Children[1]
	if second.Y != 2 {
		t.Errorf("second child Y = %d, want 2", second.Y)
	}
}
