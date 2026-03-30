package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestBorderTopDrawsHorizontalLine(t *testing.T) {
	// Given a box with only border-top
	input := &Input{
		Kind:      KindBox,
		BorderTop: "single",
		FixedWidth: 10,
		Children: []*Input{{Kind: KindText, Content: "hello"}},
	}

	// When
	box := Layout(input, 20, 5)
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then row 0 should have a horizontal line
	if buf.Cell(0, 0).Ch != '─' {
		t.Errorf("expected ─ at (0,0), got %c", buf.Cell(0, 0).Ch)
	}
	if buf.Cell(0, 9).Ch != '─' {
		t.Errorf("expected ─ at (0,9), got %c", buf.Cell(0, 9).Ch)
	}
	// Content should be offset by 1 row
	if buf.Cell(1, 0).Ch != 'h' {
		t.Errorf("expected 'h' at (1,0), got %c", buf.Cell(1, 0).Ch)
	}
	// No left/right/bottom borders
	if buf.Cell(1, 9).Ch == '│' {
		t.Error("unexpected right border")
	}
}

func TestBorderBottomDrawsHorizontalLine(t *testing.T) {
	// Given a box with only border-bottom
	input := &Input{
		Kind:         KindBox,
		BorderBottom: "single",
		FixedWidth:   10,
		Children:     []*Input{{Kind: KindText, Content: "hello"}},
	}

	// When
	box := Layout(input, 20, 5)
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then row 0 has content (no top border)
	if buf.Cell(0, 0).Ch != 'h' {
		t.Errorf("expected 'h' at (0,0), got %c", buf.Cell(0, 0).Ch)
	}
	// Row 1 should be the bottom border
	if buf.Cell(1, 0).Ch != '─' {
		t.Errorf("expected ─ at (1,0), got %c", buf.Cell(1, 0).Ch)
	}
}

func TestBorderTopAndBottomNoSides(t *testing.T) {
	// Given a box with border-top and border-bottom
	input := &Input{
		Kind:         KindBox,
		BorderTop:    "single",
		BorderBottom: "single",
		FixedWidth:   10,
		Children:     []*Input{{Kind: KindText, Content: "hello"}},
	}

	// When
	box := Layout(input, 20, 5)
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then: top rule, content, bottom rule — no side borders
	if buf.Cell(0, 0).Ch != '─' {
		t.Errorf("expected ─ at (0,0), got %c", buf.Cell(0, 0).Ch)
	}
	if buf.Cell(1, 0).Ch != 'h' {
		t.Errorf("expected 'h' at (1,0), got %c", buf.Cell(1, 0).Ch)
	}
	if buf.Cell(2, 0).Ch != '─' {
		t.Errorf("expected ─ at (2,0), got %c", buf.Cell(2, 0).Ch)
	}
	// Height should be 3 (top + content + bottom)
	if box.Height != 3 {
		t.Errorf("box.Height = %d, want 3", box.Height)
	}
}
