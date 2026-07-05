package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestRenderTreeDrawsVerticalScrollbar(t *testing.T) {
	// Given — a box that needs a scrollbar with content taller than viewport
	box := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		ContentHeight:  20,
		NeedsScrollbar: true,
		ScrollY:        0,
		Clip:           &render.Clip{Top: 0, Left: 0, Bottom: 4, Right: 9},
		Children: []*Box{
			{X: 0, Y: 0, Width: 9, Height: 1, Content: "hello"},
		},
	}
	buf := render.NewBuffer(10, 5)

	// When
	RenderTree(buf, box, nil)

	// Then — scrollbar should be drawn at the right edge of clip (column 9)
	thumbFound := false
	trackFound := false
	for row := 0; row < 5; row++ {
		ch := buf.Cell(row, 9).Ch
		if ch == '█' {
			thumbFound = true
		}
		if ch == '░' {
			trackFound = true
		}
	}
	if !thumbFound {
		t.Error("expected thumb character (█) in vertical scrollbar")
	}
	if !trackFound {
		t.Error("expected track character (░) in vertical scrollbar")
	}
}

func TestRenderTreeNoScrollbarWhenNotNeeded(t *testing.T) {
	// Given — a box that does NOT need a scrollbar
	box := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		NeedsScrollbar: false,
		Clip:           &render.Clip{Top: 0, Left: 0, Bottom: 4, Right: 9},
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "hello"},
		},
	}
	buf := render.NewBuffer(10, 5)

	// When
	RenderTree(buf, box, nil)

	// Then — no scrollbar characters at the right edge
	for row := 0; row < 5; row++ {
		ch := buf.Cell(row, 9).Ch
		if ch == '█' || ch == '░' {
			t.Errorf("unexpected scrollbar character at row %d, col 9: %c", row, ch)
		}
	}
}

func TestRenderTreeAppliesScrollXOffset(t *testing.T) {
	// Given — a box with ScrollX=5, child at X=2
	box := &Box{
		X: 0, Y: 0, Width: 10, Height: 3,
		ScrollX: 5,
		Clip:    &render.Clip{Top: 0, Left: 0, Bottom: 2, Right: 9},
		Children: []*Box{
			{X: 7, Y: 0, Width: 5, Height: 1, Content: "ABCDE"},
		},
	}
	buf := render.NewBuffer(10, 3)

	// When
	RenderTree(buf, box, nil)

	// Then — child at X=7 shifted left by 5 → visible at X=2
	if ch := buf.Cell(0, 2).Ch; ch != 'A' {
		t.Errorf("Cell(0, 2).Ch = %c, want 'A' (child shifted by ScrollX)", ch)
	}
}

func TestRenderTreeAppliesBothScrollOffsets(t *testing.T) {
	// Given — a box with ScrollX=3 and ScrollY=2, child at X=5, Y=4
	box := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		ScrollX: 3,
		ScrollY: 2,
		Clip:    &render.Clip{Top: 0, Left: 0, Bottom: 4, Right: 9},
		Children: []*Box{
			{X: 5, Y: 4, Width: 3, Height: 1, Content: "Hi!"},
		},
	}
	buf := render.NewBuffer(10, 5)

	// When
	RenderTree(buf, box, nil)

	// Then — child shifted: X=5-3=2, Y=4-2=2
	if ch := buf.Cell(2, 2).Ch; ch != 'H' {
		t.Errorf("Cell(2, 2).Ch = %c, want 'H' (shifted by ScrollX+ScrollY)", ch)
	}
}

func TestRenderTreeScrollbarNarrowsContentClip(t *testing.T) {
	// Given — a box that needs a scrollbar with a child spanning full width
	box := &Box{
		X: 0, Y: 0, Width: 10, Height: 5,
		ContentHeight:  20,
		NeedsScrollbar: true,
		ScrollY:        0,
		Clip:           &render.Clip{Top: 0, Left: 0, Bottom: 4, Right: 9},
		Children: []*Box{
			{X: 0, Y: 0, Width: 10, Height: 1, Content: "1234567890"},
		},
	}
	buf := render.NewBuffer(10, 5)

	// When
	RenderTree(buf, box, nil)

	// Then — content at column 9 should be scrollbar, not text
	ch := buf.Cell(0, 9).Ch
	if ch != '█' && ch != '░' {
		t.Errorf("expected scrollbar at col 9, got %c", ch)
	}
	// Content should be clipped to columns 0-8
	ch = buf.Cell(0, 8).Ch
	if ch != '9' {
		t.Errorf("expected '9' at col 8 (content clipped), got %c", ch)
	}
}

// C14/C15: a box with Cells blits the buffer at its content origin.
func TestRenderTreeBlitsCells(t *testing.T) {
	// Given — a 3x2 cell grid inside a bordered box
	cells := render.NewBuffer(3, 2)
	cells.SetStyledCell(0, 0, 'A', render.Style{FG: render.Color{Name: "red"}})
	cells.SetStyledCell(1, 2, 'Z', render.Style{})
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindBox, Border: "single", FixedWidth: 5, FixedHeight: 4, Cells: cells},
	}}

	// When
	box := Layout(tree, 10, 6)
	buf := render.NewBuffer(10, 6)
	RenderTree(buf, box, nil)

	// Then — cells land inside the border
	if got := buf.Cell(1, 1); got.Ch != 'A' || got.Style.FG.Name != "red" {
		t.Errorf("cell(1,1) = %c %v, want styled A", got.Ch, got.Style.FG)
	}
	if got := buf.Cell(2, 3); got.Ch != 'Z' {
		t.Errorf("cell(2,3) = %c, want Z", got.Ch)
	}
}

// Cells larger than the box clip to its content area.
func TestRenderTreeClipsCellsToContentArea(t *testing.T) {
	// Given — a 10-wide grid in a 4-wide borderless box
	cells := render.NewBuffer(10, 1)
	for i := 0; i < 10; i++ {
		cells.SetStyledCell(0, i, rune('0'+i), render.Style{})
	}
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1, Cells: cells},
	}}

	// When
	box := Layout(tree, 20, 3)
	buf := render.NewBuffer(20, 3)
	RenderTree(buf, box, nil)

	// Then
	if got := buf.Cell(0, 3); got.Ch != '3' {
		t.Errorf("cell(0,3) = %c, want 3", got.Ch)
	}
	if got := buf.Cell(0, 4); got.Ch == '4' {
		t.Errorf("cell(0,4) leaked outside the box: %c", got.Ch)
	}
}
