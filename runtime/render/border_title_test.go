package render

import "testing"

func TestDrawBorderTitle(t *testing.T) {
	// Given — a 12-wide box with border drawn
	b := NewBuffer(12, 3)
	b.DrawBorder(0, 0, 12, 3, "single")

	// When
	b.DrawBorderTitle(0, 0, 12, "Title")

	// Then — top edge should read: ┌─ Title ───┐
	// Position 0: ┌ (unchanged corner)
	if c := b.Cell(0, 0); c.Ch != '┌' {
		t.Errorf("Cell(0,0) = %c, want ┌", c.Ch)
	}
	// Position 1: ─ (space before title)
	if c := b.Cell(0, 1); c.Ch != '─' {
		t.Errorf("Cell(0,1) = %c, want ─", c.Ch)
	}
	// Position 2: ' ' (space)
	if c := b.Cell(0, 2); c.Ch != ' ' {
		t.Errorf("Cell(0,2) = %c, want space", c.Ch)
	}
	// Positions 3-7: "Title"
	title := "Title"
	for i, ch := range title {
		if c := b.Cell(0, 3+i); c.Ch != ch {
			t.Errorf("Cell(0,%d) = %c, want %c", 3+i, c.Ch, ch)
		}
	}
	// Position 8: ' ' (space after title)
	if c := b.Cell(0, 8); c.Ch != ' ' {
		t.Errorf("Cell(0,8) = %c, want space", c.Ch)
	}
	// Positions 9-10: ─ (remaining fill)
	for col := 9; col <= 10; col++ {
		if c := b.Cell(0, col); c.Ch != '─' {
			t.Errorf("Cell(0,%d) = %c, want ─", col, c.Ch)
		}
	}
	// Position 11: ┐ (unchanged corner)
	if c := b.Cell(0, 11); c.Ch != '┐' {
		t.Errorf("Cell(0,11) = %c, want ┐", c.Ch)
	}
}

func TestDrawBorderTitleTruncation(t *testing.T) {
	// Given — a 8-wide box, title "VeryLongTitle" won't fit (width-4 = 4 chars max)
	b := NewBuffer(8, 3)
	b.DrawBorder(0, 0, 8, 3, "single")

	// When
	b.DrawBorderTitle(0, 0, 8, "VeryLongTitle")

	// Then — title truncated to "Very": ┌─ Very ┐
	title := "Very"
	for i, ch := range title {
		if c := b.Cell(0, 3+i); c.Ch != ch {
			t.Errorf("Cell(0,%d) = %c, want %c", 3+i, c.Ch, ch)
		}
	}
}

func TestDrawBorderTitleEmpty(t *testing.T) {
	// Given
	b := NewBuffer(10, 3)
	b.DrawBorder(0, 0, 10, 3, "single")

	// When — empty title is a no-op
	b.DrawBorderTitle(0, 0, 10, "")

	// Then — top edge unchanged: all ─ between corners
	for col := 1; col <= 8; col++ {
		if c := b.Cell(0, col); c.Ch != '─' {
			t.Errorf("Cell(0,%d) = %c, want ─", col, c.Ch)
		}
	}
}
