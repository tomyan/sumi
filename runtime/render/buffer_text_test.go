package render

import "testing"

func TestToPlainTextEmptyBuffer(t *testing.T) {
	// Given
	buf := NewBuffer(5, 3)

	// When
	result := buf.ToPlainText()

	// Then
	if result != "" {
		t.Errorf("ToPlainText() = %q, want %q", result, "")
	}
}

func TestToPlainTextSingleRow(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteText(0, 0, "Hello")

	// When
	result := buf.ToPlainText()

	// Then
	want := "Hello"
	if result != want {
		t.Errorf("ToPlainText() = %q, want %q", result, want)
	}
}

func TestToPlainTextMultipleRows(t *testing.T) {
	// Given
	buf := NewBuffer(10, 3)
	buf.WriteText(0, 0, "Hello")
	buf.WriteText(1, 0, "World")

	// When
	result := buf.ToPlainText()

	// Then
	want := "Hello\nWorld"
	if result != want {
		t.Errorf("ToPlainText() = %q, want %q", result, want)
	}
}

func TestToPlainTextEmptyCellsBecomeSpaces(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.SetCell(0, 0, 'A')
	buf.SetCell(0, 4, 'B')

	// When
	result := buf.ToPlainText()

	// Then
	want := "A   B"
	if result != want {
		t.Errorf("ToPlainText() = %q, want %q", result, want)
	}
}

func TestToPlainTextTrailingSpacesTrimmed(t *testing.T) {
	// Given
	buf := NewBuffer(10, 1)
	buf.WriteText(0, 0, "Hi")

	// When
	result := buf.ToPlainText()

	// Then
	want := "Hi"
	if result != want {
		t.Errorf("ToPlainText() = %q, want %q", result, want)
	}
}

func TestToPlainTextTrailingEmptyRowsTrimmed(t *testing.T) {
	// Given
	buf := NewBuffer(10, 5)
	buf.WriteText(0, 0, "Top")

	// When
	result := buf.ToPlainText()

	// Then
	want := "Top"
	if result != want {
		t.Errorf("ToPlainText() = %q, want %q", result, want)
	}
}

func TestToPlainTextMiddleEmptyRowPreserved(t *testing.T) {
	// Given
	buf := NewBuffer(10, 3)
	buf.WriteText(0, 0, "Top")
	buf.WriteText(2, 0, "Bottom")

	// When
	result := buf.ToPlainText()

	// Then
	want := "Top\n\nBottom"
	if result != want {
		t.Errorf("ToPlainText() = %q, want %q", result, want)
	}
}
