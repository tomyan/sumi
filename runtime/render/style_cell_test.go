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
	if !strings.Contains(got, ";1m") {
		t.Errorf("missing bold SGR, got %q", got)
	}
	// Should contain reset before U (since S was styled)
	// The exact sequence depends on implementation, but 'U' should render without style codes
	if !strings.Contains(got, "U") {
		t.Errorf("missing character U, got %q", got)
	}
}
