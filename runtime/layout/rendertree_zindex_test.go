package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestRenderHigherZIndexOnTop(t *testing.T) {
	// Given — two overlapping boxes, higher z-index paints on top
	parent := &Box{
		X: 0, Y: 0, Width: 20, Height: 5,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 5, Height: 1,
				Content: "AAAAA",
				ZIndex:  1,
			},
			{
				X: 0, Y: 0, Width: 5, Height: 1,
				Content: "BBBBB",
				ZIndex:  2,
			},
		},
	}

	// When
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, parent, nil)

	// Then — B (z-index:2) paints on top of A (z-index:1)
	got := bufRowString(buf, 0, 5)
	if got != "BBBBB" {
		t.Errorf("expected 'BBBBB' on top, got %q", got)
	}
}

func TestRenderEqualZIndexUsesDocumentOrder(t *testing.T) {
	// Given — same z-index, later sibling paints on top
	parent := &Box{
		X: 0, Y: 0, Width: 20, Height: 5,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 5, Height: 1,
				Content: "FIRST",
				ZIndex:  0,
			},
			{
				X: 0, Y: 0, Width: 5, Height: 1,
				Content: "LATER",
				ZIndex:  0,
			},
		},
	}

	// When
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, parent, nil)

	// Then — later sibling wins (stable sort preserves document order)
	got := bufRowString(buf, 0, 5)
	if got != "LATER" {
		t.Errorf("expected 'LATER' on top, got %q", got)
	}
}

func TestLayoutZIndexDoesNotAffectLayout(t *testing.T) {
	// Given — z-index is purely visual, no layout effect
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "A", ZIndex: 10},
			{Kind: KindText, Content: "B", ZIndex: 5},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — positions are normal flow order regardless of z-index
	if box.Children[0].Y != 0 {
		t.Errorf("expected first child Y=0, got %d", box.Children[0].Y)
	}
	if box.Children[1].Y != 1 {
		t.Errorf("expected second child Y=1, got %d", box.Children[1].Y)
	}
}
