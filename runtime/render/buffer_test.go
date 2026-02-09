package render

import "testing"

func TestNewBufferDimensions(t *testing.T) {
	// When
	b := NewBuffer(80, 24)

	// Then
	if b.Width() != 80 {
		t.Errorf("Width() = %d, want 80", b.Width())
	}
	if b.Height() != 24 {
		t.Errorf("Height() = %d, want 24", b.Height())
	}
}

func TestDefaultCellIsZeroValue(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)

	// When
	c := b.Cell(0, 0)

	// Then
	if c.Ch != 0 {
		t.Errorf("default Cell.Ch = %d, want 0", c.Ch)
	}
}

func TestSetCellGetCellRoundTrip(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)

	// When
	b.SetCell(2, 3, 'X')
	c := b.Cell(2, 3)

	// Then
	if c.Ch != 'X' {
		t.Errorf("Cell.Ch = %c, want X", c.Ch)
	}
}

func TestWriteTextWritesAcrossColumns(t *testing.T) {
	// Given
	b := NewBuffer(20, 5)

	// When
	b.WriteText(1, 2, "Hello")

	// Then
	expected := []rune{'H', 'e', 'l', 'l', 'o'}
	for i, want := range expected {
		got := b.Cell(1, 2+i)
		if got.Ch != want {
			t.Errorf("Cell(1, %d).Ch = %c, want %c", 2+i, got.Ch, want)
		}
	}
}

func TestSetCellOutOfBoundsIsNoOp(t *testing.T) {
	// Given
	b := NewBuffer(5, 5)

	// When/Then — These should not panic
	b.SetCell(-1, 0, 'A')
	b.SetCell(0, -1, 'A')
	b.SetCell(5, 0, 'A')
	b.SetCell(0, 5, 'A')
	b.SetCell(100, 100, 'A')
}

func TestCellOutOfBoundsReturnsZeroCell(t *testing.T) {
	// Given
	b := NewBuffer(5, 5)
	tests := []struct {
		row, col int
	}{
		{-1, 0},
		{0, -1},
		{5, 0},
		{0, 5},
		{100, 100},
	}
	for _, tt := range tests {
		// When
		c := b.Cell(tt.row, tt.col)

		// Then
		if c.Ch != 0 {
			t.Errorf("Cell(%d, %d).Ch = %d, want 0", tt.row, tt.col, c.Ch)
		}
	}
}

func TestWriteTextTruncatedAtBufferWidth(t *testing.T) {
	// Given
	b := NewBuffer(5, 3)

	// When — starts at col 3, only 2 cols remain
	b.WriteText(0, 3, "Hello")

	// Then
	// 'H' at col 3, 'e' at col 4 should be written
	if c := b.Cell(0, 3); c.Ch != 'H' {
		t.Errorf("Cell(0, 3).Ch = %c, want H", c.Ch)
	}
	if c := b.Cell(0, 4); c.Ch != 'e' {
		t.Errorf("Cell(0, 4).Ch = %c, want e", c.Ch)
	}
	// Nothing beyond buffer width
	// (Cell returns zero for out-of-bounds, already tested)
}
