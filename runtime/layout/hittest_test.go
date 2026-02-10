package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestHitTestScrollFindsScrollableBox(t *testing.T) {
	// Given — a tree with one scrollable box
	tree := &Box{
		X: 0, Y: 0, Width: 80, Height: 24,
		Children: []*Box{
			{
				X: 5, Y: 5, Width: 20, Height: 10,
				Clip:          &render.Clip{Top: 5, Left: 5, Bottom: 14, Right: 24},
				ContentHeight: 30,
				Children:      []*Box{},
			},
		},
	}

	// When — click inside the scrollable box
	idx := HitTestScroll(tree, 10, 7)

	// Then — should find the scrollable child (index 0)
	if idx != 0 {
		t.Errorf("HitTestScroll = %d, want 0", idx)
	}
}

func TestHitTestScrollMissesScrollableBox(t *testing.T) {
	// Given
	tree := &Box{
		X: 0, Y: 0, Width: 80, Height: 24,
		Children: []*Box{
			{
				X: 5, Y: 5, Width: 20, Height: 10,
				Clip:          &render.Clip{Top: 5, Left: 5, Bottom: 14, Right: 24},
				ContentHeight: 30,
			},
		},
	}

	// When — click outside the scrollable box
	idx := HitTestScroll(tree, 1, 1)

	// Then
	if idx != -1 {
		t.Errorf("HitTestScroll = %d, want -1", idx)
	}
}

func TestHitTestScrollMultipleBoxes(t *testing.T) {
	// Given — two scrollable boxes
	tree := &Box{
		X: 0, Y: 0, Width: 80, Height: 24,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 20, Height: 10,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 9, Right: 19},
				ContentHeight: 30,
			},
			{
				X: 30, Y: 0, Width: 20, Height: 10,
				Clip:          &render.Clip{Top: 0, Left: 30, Bottom: 9, Right: 49},
				ContentHeight: 25,
			},
		},
	}

	// When — click in the second box
	idx := HitTestScroll(tree, 35, 5)

	// Then
	if idx != 1 {
		t.Errorf("HitTestScroll = %d, want 1", idx)
	}
}

func TestHitTestScrollNestedScrollableBox(t *testing.T) {
	// Given — nested scrollable box inside another
	tree := &Box{
		X: 0, Y: 0, Width: 80, Height: 24,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 40, Height: 20,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 19, Right: 39},
				ContentHeight: 50,
				Children: []*Box{
					{
						X: 5, Y: 5, Width: 20, Height: 10,
						Clip:          &render.Clip{Top: 5, Left: 5, Bottom: 14, Right: 24},
						ContentHeight: 30,
					},
				},
			},
		},
	}

	// When — click inside the inner scrollable box
	idx := HitTestScroll(tree, 10, 7)

	// Then — should find the deepest scrollable box (index 1)
	if idx != 1 {
		t.Errorf("HitTestScroll = %d, want 1 (deepest)", idx)
	}
}

func TestHitTestScrollNonScrollableBoxIgnored(t *testing.T) {
	// Given — a non-scrollable box (no ContentHeight set)
	tree := &Box{
		X: 0, Y: 0, Width: 80, Height: 24,
		Children: []*Box{
			{X: 0, Y: 0, Width: 20, Height: 10},
		},
	}

	// When
	idx := HitTestScroll(tree, 5, 5)

	// Then
	if idx != -1 {
		t.Errorf("HitTestScroll = %d, want -1 (non-scrollable)", idx)
	}
}
