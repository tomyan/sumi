package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestBoxScrollYShiftsChildrenDuringRender(t *testing.T) {
	// Given — a scrollable box with ScrollY=2, children at Y=0,1,2,3,4
	box := &Box{
		X: 0, Y: 0, Width: 20, Height: 3,
		ScrollY: 2,
		Clip:    &render.Clip{Top: 0, Left: 0, Bottom: 2, Right: 19},
		Children: []*Box{
			{X: 0, Y: 0, Width: 5, Height: 1, Content: "line0"},
			{X: 0, Y: 1, Width: 5, Height: 1, Content: "line1"},
			{X: 0, Y: 2, Width: 5, Height: 1, Content: "line2"},
			{X: 0, Y: 3, Width: 5, Height: 1, Content: "line3"},
			{X: 0, Y: 4, Width: 5, Height: 1, Content: "line4"},
		},
	}

	// When — render with scroll offset
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then — with ScrollY=2, "line2" should appear at row 0
	// line0 and line1 are scrolled above the clip
	if ch := buf.Cell(0, 0); ch.Ch != 'l' {
		t.Errorf("Cell(0,0).Ch = %c, want 'l' (from 'line2')", ch.Ch)
	}
	// "line3" should appear at row 1
	if ch := buf.Cell(1, 4); ch.Ch != '3' {
		t.Errorf("Cell(1,4).Ch = %c, want '3' (from 'line3')", ch.Ch)
	}
	// "line4" should appear at row 2
	if ch := buf.Cell(2, 4); ch.Ch != '4' {
		t.Errorf("Cell(2,4).Ch = %c, want '4' (from 'line4')", ch.Ch)
	}
	// Row 3 should be empty (only 3 visible lines: line2, line3, line4)
	if ch := buf.Cell(3, 0); ch.Ch != 0 {
		t.Errorf("Cell(3,0).Ch = %c, want 0 (below viewport)", ch.Ch)
	}
}

func TestRenderTreeWithoutScrollY(t *testing.T) {
	// Given — box without scroll
	box := &Box{
		X: 0, Y: 0, Width: 20, Height: 3,
		Children: []*Box{
			{X: 0, Y: 0, Width: 5, Height: 1, Content: "hello"},
		},
	}

	// When
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, box, nil)

	// Then — text renders at its natural position
	if ch := buf.Cell(0, 0); ch.Ch != 'h' {
		t.Errorf("Cell(0,0).Ch = %c, want 'h'", ch.Ch)
	}
}
