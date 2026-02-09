package render

import "testing"

func TestDrawBorderCorners(t *testing.T) {
	b := NewBuffer(10, 6)
	b.DrawBorder(1, 2, 4, 3, "single")

	tests := []struct {
		row, col int
		want     rune
		name     string
	}{
		{1, 2, '┌', "top-left"},
		{1, 5, '┐', "top-right"},
		{3, 2, '└', "bottom-left"},
		{3, 5, '┘', "bottom-right"},
	}
	for _, tt := range tests {
		c := b.Cell(tt.row, tt.col)
		if c.Ch != tt.want {
			t.Errorf("%s: Cell(%d, %d).Ch = %c, want %c", tt.name, tt.row, tt.col, c.Ch, tt.want)
		}
	}
}

func TestDrawBorderHorizontalEdges(t *testing.T) {
	b := NewBuffer(10, 5)
	b.DrawBorder(0, 0, 6, 4, "single")

	// Top edge: cols 1..4 should be '─'
	for col := 1; col <= 4; col++ {
		c := b.Cell(0, col)
		if c.Ch != '─' {
			t.Errorf("top edge: Cell(0, %d).Ch = %c, want ─", col, c.Ch)
		}
	}
	// Bottom edge: cols 1..4 should be '─'
	for col := 1; col <= 4; col++ {
		c := b.Cell(3, col)
		if c.Ch != '─' {
			t.Errorf("bottom edge: Cell(3, %d).Ch = %c, want ─", col, c.Ch)
		}
	}
}

func TestDrawBorderVerticalEdges(t *testing.T) {
	b := NewBuffer(10, 5)
	b.DrawBorder(0, 0, 6, 4, "single")

	// Left edge: rows 1..2 should be '│'
	for row := 1; row <= 2; row++ {
		c := b.Cell(row, 0)
		if c.Ch != '│' {
			t.Errorf("left edge: Cell(%d, 0).Ch = %c, want │", row, c.Ch)
		}
	}
	// Right edge: rows 1..2 should be '│'
	for row := 1; row <= 2; row++ {
		c := b.Cell(row, 5)
		if c.Ch != '│' {
			t.Errorf("right edge: Cell(%d, 5).Ch = %c, want │", row, c.Ch)
		}
	}
}

func TestDrawBorderSmallDimensionsNoOp(t *testing.T) {
	b := NewBuffer(10, 5)

	// Width < 2
	b.DrawBorder(0, 0, 1, 5, "single")
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("width<2: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}

	// Height < 2
	b.DrawBorder(0, 0, 5, 1, "single")
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("height<2: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}

	// Width == 0
	b.DrawBorder(0, 0, 0, 5, "single")
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("width==0: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}

	// Negative dimensions
	b.DrawBorder(0, 0, -3, -2, "single")
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("negative: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderClipsOutOfBounds(t *testing.T) {
	// Border starts at (-1, -1) with size 5x4 — should clip without panic
	b := NewBuffer(6, 5)
	b.DrawBorder(-1, -1, 5, 4, "single")

	// Top-left corner at (-1, -1) is clipped
	// Top-right corner at (-1, 3) is clipped
	// Bottom-left corner at (2, -1) is clipped
	// Bottom-right corner at (2, 3) should be visible
	c := b.Cell(2, 3)
	if c.Ch != '┘' {
		t.Errorf("clipped bottom-right: Cell(2, 3).Ch = %c, want ┘", c.Ch)
	}

	// Top edge at row -1 is clipped, but left edge at col -1 is clipped
	// Visible portion: right edge at col 3, rows 0..1
	for row := 0; row <= 1; row++ {
		c := b.Cell(row, 3)
		if c.Ch != '│' {
			t.Errorf("clipped right edge: Cell(%d, 3).Ch = %c, want │", row, c.Ch)
		}
	}
	// Bottom edge at row 2, cols 0..2
	for col := 0; col <= 2; col++ {
		c := b.Cell(2, col)
		if c.Ch != '─' {
			t.Errorf("clipped bottom edge: Cell(2, %d).Ch = %c, want ─", col, c.Ch)
		}
	}
}

func TestDrawBorderClipsRightAndBottom(t *testing.T) {
	// Border extends beyond right and bottom edges
	b := NewBuffer(4, 3)
	b.DrawBorder(1, 2, 5, 5, "single")

	// Top-left corner at (1, 2) should be visible
	c := b.Cell(1, 2)
	if c.Ch != '┌' {
		t.Errorf("Cell(1, 2).Ch = %c, want ┌", c.Ch)
	}
	// Top edge at (1, 3) should be visible
	c = b.Cell(1, 3)
	if c.Ch != '─' {
		t.Errorf("Cell(1, 3).Ch = %c, want ─", c.Ch)
	}
	// Left edge at (2, 2) should be visible
	c = b.Cell(2, 2)
	if c.Ch != '│' {
		t.Errorf("Cell(2, 2).Ch = %c, want │", c.Ch)
	}
}

func TestDrawBorderStyleNoneIsNoOp(t *testing.T) {
	b := NewBuffer(10, 5)
	b.DrawBorder(0, 0, 5, 3, "none")

	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("style none: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderStyleEmptyIsNoOp(t *testing.T) {
	b := NewBuffer(10, 5)
	b.DrawBorder(0, 0, 5, 3, "")

	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("style empty: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderMinimumSize(t *testing.T) {
	// 2x2 is the minimum — just four corners
	b := NewBuffer(5, 5)
	b.DrawBorder(1, 1, 2, 2, "single")

	if c := b.Cell(1, 1); c.Ch != '┌' {
		t.Errorf("Cell(1,1).Ch = %c, want ┌", c.Ch)
	}
	if c := b.Cell(1, 2); c.Ch != '┐' {
		t.Errorf("Cell(1,2).Ch = %c, want ┐", c.Ch)
	}
	if c := b.Cell(2, 1); c.Ch != '└' {
		t.Errorf("Cell(2,1).Ch = %c, want └", c.Ch)
	}
	if c := b.Cell(2, 2); c.Ch != '┘' {
		t.Errorf("Cell(2,2).Ch = %c, want ┘", c.Ch)
	}
}

func TestDrawBorderInteriorUntouched(t *testing.T) {
	b := NewBuffer(10, 8)
	b.DrawBorder(0, 0, 5, 4, "single")

	// Interior cells (rows 1..2, cols 1..3) should be untouched (zero)
	for row := 1; row <= 2; row++ {
		for col := 1; col <= 3; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("interior: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}
