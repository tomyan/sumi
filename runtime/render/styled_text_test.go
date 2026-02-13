package render

import "testing"

func TestToStyledTextPlainContent(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteText(0, 0, "Hello")

	// When
	result := buf.ToStyledText()

	// Then
	want := "Hello"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextFGColor(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Hi", Style{FG: Color{Name: "green"}})

	// When
	result := buf.ToStyledText()

	// Then
	want := "<<green>>Hi<</>>"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextBGColor(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Hi", Style{BG: Color{Name: "red"}})

	// When
	result := buf.ToStyledText()

	// Then
	want := "<<bg:red>>Hi<</>>"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextBoldAndColor(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Hi", Style{FG: Color{Name: "green"}, Bold: true})

	// When
	result := buf.ToStyledText()

	// Then
	want := "<<green,bold>>Hi<</>>"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextStyleChangesMidRow(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "AB", Style{FG: Color{Name: "red"}})
	buf.WriteText(0, 2, "CD")

	// When
	result := buf.ToStyledText()

	// Then
	want := "<<red>>AB<</>>CD"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextAttrsSortedBGThenFGThenAlpha(t *testing.T) {
	// Given — all boolean attrs + both colors
	style := Style{
		FG:            Color{Name: "cyan"},
		BG:            Color{Name: "yellow"},
		Bold:          true,
		Dim:           true,
		Italic:        true,
		Underline:     true,
		Strikethrough: true,
		Inverse:       true,
	}
	buf := NewBuffer(5, 1)
	buf.WriteStyledText(0, 0, "X", style)

	// When
	result := buf.ToStyledText()

	// Then — bg:COLOR, fg COLOR, then booleans alphabetically
	want := "<<bg:yellow,cyan,bold,dim,inverse,italic,strikethrough,underline>>X<</>>"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextMultipleRows(t *testing.T) {
	// Given
	buf := NewBuffer(10, 2)
	buf.WriteStyledText(0, 0, "Red", Style{FG: Color{Name: "red"}})
	buf.WriteText(1, 0, "Plain")

	// When
	result := buf.ToStyledText()

	// Then
	want := "<<red>>Red<</>>\nPlain"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextTrailingStyledSpacesTrimmed(t *testing.T) {
	// Given — styled text followed by styled empty cells
	buf := NewBuffer(10, 1)
	buf.WriteStyledText(0, 0, "Hi", Style{FG: Color{Name: "blue"}})

	// When
	result := buf.ToStyledText()

	// Then — trailing empty cells should not produce styled spaces
	want := "<<blue>>Hi<</>>"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}

func TestToStyledTextEmptyBuffer(t *testing.T) {
	// Given
	buf := NewBuffer(5, 3)

	// When
	result := buf.ToStyledText()

	// Then
	if result != "" {
		t.Errorf("ToStyledText() = %q, want %q", result, "")
	}
}

func TestToStyledTextStyleSpansGap(t *testing.T) {
	// Given — styled cell, gap, styled cell with same style
	buf := NewBuffer(10, 1)
	style := Style{FG: Color{Name: "green"}}
	buf.SetStyledCell(0, 0, 'A', style)
	buf.SetStyledCell(0, 3, 'B', style)

	// When
	result := buf.ToStyledText()

	// Then — the gap (empty cells with zero style) should close and reopen
	want := "<<green>>A<</>>  <<green>>B<</>>"
	if result != want {
		t.Errorf("ToStyledText() = %q, want %q", result, want)
	}
}
