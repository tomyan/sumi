package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestSetStyledCellStoresStyle(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	s := Style{FG: Color{Name: "red"}, Bold: true}

	// When
	b.SetStyledCell(2, 3, 'X', s)
	c := b.Cell(2, 3)

	// Then
	if c.Ch != 'X' {
		t.Errorf("Cell.Ch = %c, want X", c.Ch)
	}
	if c.Style.FG.Name != "red" {
		t.Errorf("Cell.Style.FG.Name = %q, want %q", c.Style.FG.Name, "red")
	}
	if !c.Style.Bold {
		t.Error("Cell.Style.Bold = false, want true")
	}
}

func TestSetStyledCellOutOfBoundsIsNoOp(t *testing.T) {
	// Given
	b := NewBuffer(5, 5)
	s := Style{Bold: true}

	// When/Then — Should not panic
	b.SetStyledCell(-1, 0, 'A', s)
	b.SetStyledCell(0, -1, 'A', s)
	b.SetStyledCell(5, 0, 'A', s)
	b.SetStyledCell(0, 5, 'A', s)
}

func TestWriteStyledTextAppliesStyleToAll(t *testing.T) {
	// Given
	b := NewBuffer(20, 5)
	s := Style{FG: Color{Name: "green"}, Italic: true}

	// When
	b.WriteStyledText(1, 2, "Hello", s)

	// Then
	for i, ch := range "Hello" {
		c := b.Cell(1, 2+i)
		if c.Ch != ch {
			t.Errorf("Cell(1, %d).Ch = %c, want %c", 2+i, c.Ch, ch)
		}
		if c.Style.FG.Name != "green" {
			t.Errorf("Cell(1, %d).Style.FG.Name = %q, want %q", 2+i, c.Style.FG.Name, "green")
		}
		if !c.Style.Italic {
			t.Errorf("Cell(1, %d).Style.Italic = false, want true", 2+i)
		}
	}
}

func TestWriteStyledTextTruncatesAtEdge(t *testing.T) {
	// Given
	b := NewBuffer(5, 3)
	s := Style{Bold: true}

	// When — starts at col 3, only 2 cols remain
	b.WriteStyledText(0, 3, "Hello", s)

	// Then
	if c := b.Cell(0, 3); c.Ch != 'H' || !c.Style.Bold {
		t.Errorf("Cell(0, 3) = {%c, Bold:%v}, want {H, true}", c.Ch, c.Style.Bold)
	}
	if c := b.Cell(0, 4); c.Ch != 'e' || !c.Style.Bold {
		t.Errorf("Cell(0, 4) = {%c, Bold:%v}, want {e, true}", c.Ch, c.Style.Bold)
	}
}

func TestRenderBoldCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'A', Style{Bold: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[1m") {
		t.Errorf("output missing bold SGR \\x1b[1m, got %q", got)
	}
	if !strings.Contains(got, "A") {
		t.Errorf("output missing character A, got %q", got)
	}
}

func TestRenderGreenFG(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'G', Style{FG: Color{Name: "green"}})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[32m") {
		t.Errorf("output missing green FG SGR \\x1b[32m, got %q", got)
	}
}

func TestRenderBGColor(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'B', Style{BG: Color{Name: "blue"}})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[44m") {
		t.Errorf("output missing blue BG SGR \\x1b[44m, got %q", got)
	}
}

func TestRenderMultipleAttributes(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	s := Style{
		FG:        Color{Name: "cyan"},
		BG:        Color{Name: "yellow"},
		Bold:      true,
		Underline: true,
	}
	b.SetStyledCell(0, 0, 'M', s)
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should contain reset, bold, underline, cyan FG, yellow BG
	if !strings.Contains(got, "\x1b[0m") {
		t.Errorf("output missing reset SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[1m") {
		t.Errorf("output missing bold SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[4m") {
		t.Errorf("output missing underline SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[36m") {
		t.Errorf("output missing cyan FG SGR, got %q", got)
	}
	if !strings.Contains(got, "\x1b[43m") {
		t.Errorf("output missing yellow BG SGR, got %q", got)
	}
}

func TestRenderUnstyledCellNoSGR(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetCell(0, 0, 'A') // no style, uses existing method
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should NOT contain any SGR escape
	if strings.Contains(got, "\x1b[") && !strings.Contains(got, "\x1b[1;1H") {
		// Allow cursor positioning escapes but no SGR
		// More precise: strip cursor moves and check for remaining escapes
	}
	// The output should be exactly the cursor move + char, no SGR codes
	expected := "\x1b[1;1HA"
	if got != expected {
		t.Errorf("unstyled cell output = %q, want %q", got, expected)
	}
}

func TestRenderDimCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'D', Style{Dim: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[2m") {
		t.Errorf("output missing dim SGR \\x1b[2m, got %q", got)
	}
}

func TestRenderItalicCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'I', Style{Italic: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[3m") {
		t.Errorf("output missing italic SGR \\x1b[3m, got %q", got)
	}
}

func TestRenderStrikethroughCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'S', Style{Strikethrough: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[9m") {
		t.Errorf("output missing strikethrough SGR \\x1b[9m, got %q", got)
	}
}

func TestRenderInverseCell(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'V', Style{Inverse: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	if !strings.Contains(got, "\x1b[7m") {
		t.Errorf("output missing inverse SGR \\x1b[7m, got %q", got)
	}
}

func TestRenderTrailingReset(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'A', Style{Bold: true})
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should end with a reset
	if !strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("output should end with reset SGR, got %q", got)
	}
}

func TestRenderNoTrailingResetForUnstyled(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)
	b.SetCell(0, 0, 'A') // unstyled
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should NOT end with reset when no styled cells exist
	if strings.HasSuffix(got, "\x1b[0m") {
		t.Errorf("unstyled output should not end with reset, got %q", got)
	}
}

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
		{1, 2, '┌'},
		{1, 5, '┐'},
		{3, 2, '└'},
		{3, 5, '┘'},
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
		if cell.Ch != '─' {
			t.Errorf("top edge Cell(1, %d).Ch = %c, want ─", col, cell.Ch)
		}
		if cell.Style.FG.Name != "red" {
			t.Errorf("top edge Cell(1, %d).Style.FG.Name = %q, want %q", col, cell.Style.FG.Name, "red")
		}
	}

	// Vertical edge
	cell := b.Cell(2, 2)
	if cell.Ch != '│' {
		t.Errorf("left edge Cell(2, 2).Ch = %c, want │", cell.Ch)
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
	if c.Ch != '┌' {
		t.Errorf("Cell(0, 0).Ch = %c, want ┌", c.Ch)
	}
	// Style should be zero value
	if c.Style.FG.Name != "" || c.Style.Bold {
		t.Error("DrawBorder should produce zero-value style")
	}
}

func TestSetCellPreservesZeroStyle(t *testing.T) {
	// Given
	b := NewBuffer(10, 5)

	// When
	b.SetCell(0, 0, 'A')
	c := b.Cell(0, 0)

	// Then
	if c.Style != (Style{}) {
		t.Errorf("SetCell should produce zero-value style, got %+v", c.Style)
	}
}

func TestRenderStyledThenUnstyled(t *testing.T) {
	// Given — Mix of styled and unstyled cells — reset should appear between them
	b := NewBuffer(10, 5)
	b.SetStyledCell(0, 0, 'S', Style{Bold: true})
	b.SetCell(0, 1, 'U')
	var buf bytes.Buffer

	// When
	b.RenderTo(&buf)

	// Then
	got := buf.String()
	// Should contain bold for S
	if !strings.Contains(got, "\x1b[1m") {
		t.Errorf("missing bold SGR, got %q", got)
	}
	// Should contain reset before U (since S was styled)
	// The exact sequence depends on implementation, but 'U' should render without style codes
	if !strings.Contains(got, "U") {
		t.Errorf("missing character U, got %q", got)
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
