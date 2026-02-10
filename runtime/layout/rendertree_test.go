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
