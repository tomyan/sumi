package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestColorToANSICode(t *testing.T) {
	// Given
	tests := []struct {
		name string
		fg   int
		bg   int
	}{
		{"black", 30, 40},
		{"red", 31, 41},
		{"green", 32, 42},
		{"yellow", 33, 43},
		{"blue", 34, 44},
		{"magenta", 35, 45},
		{"cyan", 36, 46},
		{"white", 37, 47},
	}
	for _, tt := range tests {
		// When
		fgCode, fgOK := colorToFGCode(tt.name)

		// Then
		if !fgOK {
			t.Errorf("colorToFGCode(%q) returned not OK", tt.name)
		}
		if fgCode != tt.fg {
			t.Errorf("colorToFGCode(%q) = %d, want %d", tt.name, fgCode, tt.fg)
		}

		// When
		bgCode, bgOK := colorToBGCode(tt.name)

		// Then
		if !bgOK {
			t.Errorf("colorToBGCode(%q) returned not OK", tt.name)
		}
		if bgCode != tt.bg {
			t.Errorf("colorToBGCode(%q) = %d, want %d", tt.name, bgCode, tt.bg)
		}
	}
}

func TestColorToANSICodeEmptyIsDefault(t *testing.T) {
	// When
	_, ok := colorToFGCode("")

	// Then
	if ok {
		t.Error("colorToFGCode(\"\") should return not OK for empty/default color")
	}

	// When
	_, ok = colorToBGCode("")

	// Then
	if ok {
		t.Error("colorToBGCode(\"\") should return not OK for empty/default color")
	}
}

func TestDrawStyledBorderAppliesStyle(t *testing.T) {
	// Given
	b := NewBuffer(10, 6)
	s := Style{FG: Color{Name: "red"}, Bold: true}

	// When
	b.DrawStyledBorder(1, 2, 4, 3, "single", s)

	// Then
	// Check corners have the style
	corners := []struct {
		row, col int
		ch       rune
	}{
		{1, 2, '\u250c'},
		{1, 5, '\u2510'},
		{3, 2, '\u2514'},
		{3, 5, '\u2518'},
	}
	for _, c := range corners {
		cell := b.Cell(c.row, c.col)
		if cell.Ch != c.ch {
			t.Errorf("Cell(%d, %d).Ch = %c, want %c", c.row, c.col, cell.Ch, c.ch)
		}
		if cell.Style.FG.Name != "red" {
			t.Errorf("Cell(%d, %d).Style.FG.Name = %q, want %q", c.row, c.col, cell.Style.FG.Name, "red")
		}
		if !cell.Style.Bold {
			t.Errorf("Cell(%d, %d).Style.Bold = false, want true", c.row, c.col)
		}
	}

	// Check edge characters have the style
	// Top edge
	for col := 3; col <= 4; col++ {
		cell := b.Cell(1, col)
		if cell.Ch != '\u2500' {
			t.Errorf("top edge Cell(1, %d).Ch = %c, want \u2500", col, cell.Ch)
		}
		if cell.Style.FG.Name != "red" {
			t.Errorf("top edge Cell(1, %d).Style.FG.Name = %q, want %q", col, cell.Style.FG.Name, "red")
		}
	}

	// Vertical edge
	cell := b.Cell(2, 2)
	if cell.Ch != '\u2502' {
		t.Errorf("left edge Cell(2, 2).Ch = %c, want \u2502", cell.Ch)
	}
	if cell.Style.FG.Name != "red" {
		t.Errorf("left edge Cell(2, 2).Style.FG.Name = %q, want %q", cell.Style.FG.Name, "red")
	}
}

func TestDrawStyledBorderNoneIsNoOp(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)

	// When
	b.DrawStyledBorder(0, 0, 5, 3, "none", Style{Bold: true})

	// Then
	for row := 0; row < 5; row++ {
		for col := 0; col < 10; col++ {
			if c := b.Cell(row, col); c.Ch != 0 {
				t.Errorf("style none: Cell(%d, %d).Ch = %c, want 0", row, col, c.Ch)
			}
		}
	}
}

func TestDrawBorderStillWorks(t *testing.T) {
	// Given
	b := NewBuffer(10, 6)

	// When — Verify the existing DrawBorder still works with zero-value style
	b.DrawBorder(0, 0, 4, 3, "single")

	// Then
	c := b.Cell(0, 0)
	if c.Ch != '\u250c' {
		t.Errorf("Cell(0, 0).Ch = %c, want \u250c", c.Ch)
	}
	// Style should be zero value
	if c.Style.FG.Name != "" || c.Style.Bold {
		t.Error("DrawBorder should produce zero-value style")
	}
}

func TestRenderAllFGColors(t *testing.T) {
	// Given
	colors := map[string]string{
		"black":   "\x1b[30m",
		"red":     "\x1b[31m",
		"green":   "\x1b[32m",
		"yellow":  "\x1b[33m",
		"blue":    "\x1b[34m",
		"magenta": "\x1b[35m",
		"cyan":    "\x1b[36m",
		"white":   "\x1b[37m",
	}
	for name, expected := range colors {
		// Given
		b := NewBuffer(10, 5)
		b.SetStyledCell(0, 0, 'X', Style{FG: Color{Name: name}})
		var buf bytes.Buffer

		// When
		b.RenderTo(&buf)

		// Then
		got := buf.String()
		if !strings.Contains(got, expected) {
			t.Errorf("FG %s: output missing %q, got %q", name, expected, got)
		}
	}
}

func TestRenderAllBGColors(t *testing.T) {
	// Given
	colors := map[string]string{
		"black":   "\x1b[40m",
		"red":     "\x1b[41m",
		"green":   "\x1b[42m",
		"yellow":  "\x1b[43m",
		"blue":    "\x1b[44m",
		"magenta": "\x1b[45m",
		"cyan":    "\x1b[46m",
		"white":   "\x1b[47m",
	}
	for name, expected := range colors {
		// Given
		b := NewBuffer(10, 5)
		b.SetStyledCell(0, 0, 'X', Style{BG: Color{Name: name}})
		var buf bytes.Buffer

		// When
		b.RenderTo(&buf)

		// Then
		got := buf.String()
		if !strings.Contains(got, expected) {
			t.Errorf("BG %s: output missing %q, got %q", name, expected, got)
		}
	}
}
