package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestRenderStickyStaysVisibleWhenScrolled(t *testing.T) {
	// Given — a scrollable parent with a sticky child (top:0) and regular children.
	// When scrolled down, the sticky child should clamp to the top of the clip.
	parent := &Box{
		X: 0, Y: 0, Width: 20, Height: 5,
		ScrollY:       3, // scrolled down 3 rows
		ContentHeight: 20,
		Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 4, Right: 19},
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 10, Height: 1,
				Content:  "STICKY",
				Position: "sticky",
				Top:      0,
			},
			{
				X: 0, Y: 1, Width: 10, Height: 1,
				Content: "normal1",
			},
			{
				X: 0, Y: 4, Width: 10, Height: 1,
				Content: "normal2",
			},
		},
	}

	// When
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, parent, nil)

	// Then — sticky child stays at visible row 0 (clip.Top + child.Top)
	got := bufRowString(buf, 0, 6)
	if got != "STICKY" {
		t.Errorf("expected 'STICKY' at row 0, got %q", got)
	}
}

func TestRenderStickyFlowsNormally(t *testing.T) {
	// Given — not scrolled yet, sticky child is at its normal flow position
	parent := &Box{
		X: 0, Y: 0, Width: 20, Height: 5,
		ScrollY:       0, // no scroll
		ContentHeight: 20,
		Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 4, Right: 19},
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 10, Height: 1,
				Content:  "STICKY",
				Position: "sticky",
				Top:      0,
			},
			{
				X: 0, Y: 1, Width: 10, Height: 1,
				Content: "normal",
			},
		},
	}

	// When
	buf := render.NewBuffer(20, 5)
	RenderTree(buf, parent, nil)

	// Then — sticky child at its normal position (row 0)
	got := bufRowString(buf, 0, 6)
	if got != "STICKY" {
		t.Errorf("expected 'STICKY' at row 0, got %q", got)
	}
}

func TestLayoutStickyAffectsFlowLikeNormal(t *testing.T) {
	// Given — sticky element is in flow for layout purposes
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "sticky", Position: "sticky", Top: 0},
			{Kind: KindText, Content: "after"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — sticky child contributes to parent size (in flow)
	if box.Children[0].Y != 0 {
		t.Errorf("expected sticky child Y=0, got %d", box.Children[0].Y)
	}
	if box.Children[1].Y != 1 {
		t.Errorf("expected next child Y=1, got %d", box.Children[1].Y)
	}
	if box.Height != 2 {
		t.Errorf("expected parent height=2, got %d", box.Height)
	}
}
