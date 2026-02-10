package render

import "testing"

func TestWriteStyledTextClippedWithinClip(t *testing.T) {
	// Given
	buf := NewBuffer(20, 10)
	clip := &Clip{Top: 2, Left: 3, Bottom: 7, Right: 15}

	// When — write text fully within clip region
	buf.WriteStyledTextClipped(3, 5, "hello", Style{}, clip)

	// Then — all characters should be written
	for i, ch := range "hello" {
		got := buf.Cell(3, 5+i)
		if got.Ch != ch {
			t.Errorf("Cell(3, %d).Ch = %c, want %c", 5+i, got.Ch, ch)
		}
	}
}

func TestWriteStyledTextClippedOutsideClipRow(t *testing.T) {
	// Given
	buf := NewBuffer(20, 10)
	clip := &Clip{Top: 2, Left: 0, Bottom: 4, Right: 19}

	// When — write to a row outside the clip
	buf.WriteStyledTextClipped(5, 0, "hidden", Style{}, clip)

	// Then — nothing should be written
	for col := 0; col < 6; col++ {
		got := buf.Cell(5, col)
		if got.Ch != 0 {
			t.Errorf("Cell(5, %d).Ch = %c, want 0 (clipped)", col, got.Ch)
		}
	}
}

func TestWriteStyledTextClippedPartiallyClippedRight(t *testing.T) {
	// Given
	buf := NewBuffer(20, 10)
	clip := &Clip{Top: 0, Left: 0, Bottom: 9, Right: 7}

	// When — text starts inside clip but extends beyond
	buf.WriteStyledTextClipped(0, 5, "hello", Style{}, clip)

	// Then — only "hel" (cols 5,6,7) should be written
	for i, ch := range "hel" {
		got := buf.Cell(0, 5+i)
		if got.Ch != ch {
			t.Errorf("Cell(0, %d).Ch = %c, want %c", 5+i, got.Ch, ch)
		}
	}
	// Col 8 should be empty (clipped)
	if got := buf.Cell(0, 8); got.Ch != 0 {
		t.Errorf("Cell(0, 8).Ch = %c, want 0 (clipped)", got.Ch)
	}
}

func TestWriteStyledTextClippedPartiallyClippedLeft(t *testing.T) {
	// Given
	buf := NewBuffer(20, 10)
	clip := &Clip{Top: 0, Left: 3, Bottom: 9, Right: 19}

	// When — text starts before clip left edge
	buf.WriteStyledTextClipped(0, 1, "hello", Style{}, clip)

	// Then — cols 1,2 should be empty (clipped), cols 3,4,5 should have 'l','l','o'
	if got := buf.Cell(0, 1); got.Ch != 0 {
		t.Errorf("Cell(0, 1).Ch = %c, want 0 (clipped left)", got.Ch)
	}
	if got := buf.Cell(0, 2); got.Ch != 0 {
		t.Errorf("Cell(0, 2).Ch = %c, want 0 (clipped left)", got.Ch)
	}
	for i, ch := range "llo" {
		got := buf.Cell(0, 3+i)
		if got.Ch != ch {
			t.Errorf("Cell(0, %d).Ch = %c, want %c", 3+i, got.Ch, ch)
		}
	}
}

func TestWriteStyledTextClippedNilClipBehavesAsUnclipped(t *testing.T) {
	// Given
	buf := NewBuffer(20, 10)

	// When — nil clip means no clipping
	buf.WriteStyledTextClipped(0, 0, "hello", Style{}, nil)

	// Then — all characters should be written
	for i, ch := range "hello" {
		got := buf.Cell(0, i)
		if got.Ch != ch {
			t.Errorf("Cell(0, %d).Ch = %c, want %c", i, got.Ch, ch)
		}
	}
}

func TestSetStyledCellClippedInsideClip(t *testing.T) {
	// Given
	buf := NewBuffer(10, 10)
	clip := &Clip{Top: 2, Left: 2, Bottom: 8, Right: 8}

	// When
	buf.SetStyledCellClipped(5, 5, 'X', Style{}, clip)

	// Then
	got := buf.Cell(5, 5)
	if got.Ch != 'X' {
		t.Errorf("Cell(5, 5).Ch = %c, want X", got.Ch)
	}
}

func TestSetStyledCellClippedOutsideClip(t *testing.T) {
	// Given
	buf := NewBuffer(10, 10)
	clip := &Clip{Top: 2, Left: 2, Bottom: 8, Right: 8}

	// When — write outside the clip
	buf.SetStyledCellClipped(1, 5, 'X', Style{}, clip)

	// Then
	got := buf.Cell(1, 5)
	if got.Ch != 0 {
		t.Errorf("Cell(1, 5).Ch = %c, want 0 (clipped)", got.Ch)
	}
}

func TestSetStyledCellClippedNilClip(t *testing.T) {
	// Given
	buf := NewBuffer(10, 10)

	// When — nil clip means no clipping
	buf.SetStyledCellClipped(0, 0, 'X', Style{}, nil)

	// Then
	got := buf.Cell(0, 0)
	if got.Ch != 'X' {
		t.Errorf("Cell(0, 0).Ch = %c, want X", got.Ch)
	}
}
