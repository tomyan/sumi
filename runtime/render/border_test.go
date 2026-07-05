package render

import "testing"

func TestDrawBorderCorners(t *testing.T) {
	// Given
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
		// When
		c := b.Cell(tt.row, tt.col)

		// Then
		if c.Ch != tt.want {
			t.Errorf("%s: Cell(%d, %d).Ch = %c, want %c", tt.name, tt.row, tt.col, c.Ch, tt.want)
		}
	}
}

func TestDrawBorderHorizontalEdges(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.DrawBorder(0, 0, 6, 4, "single")

	// Then
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
	// Given
	b := NewBuffer(10, 5)
	b.DrawBorder(0, 0, 6, 4, "single")

	// Then
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
	// Given
	b := NewBuffer(10, 5)

	// When — Width < 2
	b.DrawBorder(0, 0, 1, 5, "single")

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("width<2: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}

	// When — Height < 2
	b.DrawBorder(0, 0, 5, 1, "single")

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("height<2: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}

	// When — Width == 0
	b.DrawBorder(0, 0, 0, 5, "single")

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("width==0: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}

	// When — Negative dimensions
	b.DrawBorder(0, 0, -3, -2, "single")

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("negative: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderClipsOutOfBounds(t *testing.T) {
	// Given — Border starts at (-1, -1) with size 5x4 — should clip without panic
	b := NewBuffer(6, 5)

	// When
	b.DrawBorder(-1, -1, 5, 4, "single")

	// Then
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
	// Given — Border extends beyond right and bottom edges
	b := NewBuffer(4, 3)

	// When
	b.DrawBorder(1, 2, 5, 5, "single")

	// Then
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
	// Given
	b := NewBuffer(10, 5)

	// When
	b.DrawBorder(0, 0, 5, 3, "none")

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("style none: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderStyleEmptyIsNoOp(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)

	// When
	b.DrawBorder(0, 0, 5, 3, "")

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("style empty: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderMinimumSize(t *testing.T) {
	// Given — 2x2 is the minimum — just four corners
	b := NewBuffer(5, 5)

	// When
	b.DrawBorder(1, 1, 2, 2, "single")

	// Then
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
	// Given
	b := NewBuffer(10, 8)

	// When
	b.DrawBorder(0, 0, 5, 4, "single")

	// Then — Interior cells (rows 1..2, cols 1..3) should be untouched (zero)
	for row := 1; row <= 2; row++ {
		for col := 1; col <= 3; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("interior: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

// F1: border styles select glyph families.
func TestDrawStyledBorderStyles(t *testing.T) {
	cases := []struct {
		style          string
		tl, tr, bl, br rune
		h, v           rune
	}{
		{"double", '╔', '╗', '╚', '╝', '═', '║'},
		{"rounded", '╭', '╮', '╰', '╯', '─', '│'},
		{"heavy", '┏', '┓', '┗', '┛', '━', '┃'},
		{"ascii", '+', '+', '+', '+', '-', '|'},
		{"unknown-style", '┌', '┐', '└', '┘', '─', '│'}, // falls back to single
	}
	for _, tc := range cases {
		t.Run(tc.style, func(t *testing.T) {
			// Given / When
			b := NewBuffer(4, 3)
			b.DrawStyledBorder(0, 0, 4, 3, tc.style, Style{})

			// Then
			got := [][2]rune{
				{b.Cell(0, 0).Ch, tc.tl}, {b.Cell(0, 3).Ch, tc.tr},
				{b.Cell(2, 0).Ch, tc.bl}, {b.Cell(2, 3).Ch, tc.br},
				{b.Cell(0, 1).Ch, tc.h}, {b.Cell(1, 0).Ch, tc.v},
			}
			for i, g := range got {
				if g[0] != g[1] {
					t.Errorf("glyph %d = %c, want %c", i, g[0], g[1])
				}
			}
		})
	}
}

// F1: the border title's horizontal filler matches the border style.
func TestBorderTitleUsesStyleGlyphs(t *testing.T) {
	// Given / When
	b := NewBuffer(12, 3)
	b.DrawStyledBorder(0, 0, 12, 3, "double", Style{})
	b.DrawStyledBorderTitle(0, 0, 12, "Hi", Style{})

	// Then — corners stay double and the title is embedded
	if b.Cell(0, 0).Ch != '╔' {
		t.Errorf("corner = %c, want ╔", b.Cell(0, 0).Ch)
	}
	if b.Cell(0, 3).Ch != 'H' || b.Cell(0, 4).Ch != 'i' {
		t.Errorf("title not rendered: %c%c", b.Cell(0, 3).Ch, b.Cell(0, 4).Ch)
	}
}

// F1b: block-edge border styles.
func TestDrawBlockBorders(t *testing.T) {
	// Given / When — an eighth-cell-inner border
	b := NewBuffer(4, 3)
	b.DrawStyledBorder(0, 0, 4, 3, "eighth-cell-inner", Style{})

	// Then — thin edges hug the content; corners stay blank
	if got := b.Cell(0, 1).Ch; got != '▁' {
		t.Errorf("top edge = %q, want ▁", got)
	}
	if got := b.Cell(2, 1).Ch; got != '▔' {
		t.Errorf("bottom edge = %q, want ▔", got)
	}
	if got := b.Cell(1, 0).Ch; got != '▕' {
		t.Errorf("left edge = %q, want ▕", got)
	}
	if got := b.Cell(1, 3).Ch; got != '▏' {
		t.Errorf("right edge = %q, want ▏", got)
	}
	if got := b.Cell(0, 0).Ch; got != 0 && got != ' ' {
		t.Errorf("inner corner = %q, want blank", got)
	}

	// When — full-cell extends through the corners
	b2 := NewBuffer(4, 3)
	b2.DrawStyledBorder(0, 0, 4, 3, "full-cell", Style{})

	// Then
	if got := b2.Cell(0, 0).Ch; got != '█' {
		t.Errorf("full-cell corner = %q, want █", got)
	}
	if got := b2.Cell(1, 0).Ch; got != '█' {
		t.Errorf("full-cell side = %q, want █", got)
	}

	// When — half-cell-outer
	b3 := NewBuffer(4, 3)
	b3.DrawStyledBorder(0, 0, 4, 3, "half-cell-outer", Style{})

	// Then — top half-blocks extend through corners
	if got := b3.Cell(0, 0).Ch; got != '▀' {
		t.Errorf("half-outer corner = %q, want ▀", got)
	}
	if got := b3.Cell(1, 0).Ch; got != '▌' {
		t.Errorf("half-outer left = %q, want ▌", got)
	}
}
