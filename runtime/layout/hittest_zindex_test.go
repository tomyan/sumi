package layout

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestHitTestScrollHigherZIndexWins(t *testing.T) {
	// Given — two overlapping scrollable boxes at the same position,
	// the one with higher z-index should be hit
	tree := &Box{
		X: 0, Y: 0, Width: 30, Height: 20,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 20, Height: 10,
				ContentHeight: 50,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 9, Right: 19},
				ZIndex:        1,
			},
			{
				X: 0, Y: 0, Width: 20, Height: 10,
				ContentHeight: 50,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 9, Right: 19},
				ZIndex:        2,
			},
		},
	}

	// When — hit test at a point inside both boxes
	idx := HitTestScroll(tree, 5, 5)

	// Then — second box (z-index:2) should win (index 1 in flat list)
	if idx != 1 {
		t.Errorf("expected hit index 1 (higher z-index), got %d", idx)
	}
}

func TestHitTestScrollHigherZIndexWinsEvenWhenFirst(t *testing.T) {
	// Given — higher z-index box is FIRST in document order
	tree := &Box{
		X: 0, Y: 0, Width: 30, Height: 20,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 20, Height: 10,
				ContentHeight: 50,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 9, Right: 19},
				ZIndex:        5,
			},
			{
				X: 0, Y: 0, Width: 20, Height: 10,
				ContentHeight: 50,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 9, Right: 19},
				ZIndex:        1,
			},
		},
	}

	// When
	idx := HitTestScroll(tree, 5, 5)

	// Then — first box (z-index:5) should win even though it's first
	if idx != 0 {
		t.Errorf("expected hit index 0 (higher z-index), got %d", idx)
	}
}

func TestHitTestScrollSameZDeepestWins(t *testing.T) {
	// Given — nested scrollable boxes with same z-index,
	// deepest (most nested) box should win (existing behavior)
	tree := &Box{
		X: 0, Y: 0, Width: 40, Height: 20,
		Children: []*Box{
			{
				X: 0, Y: 0, Width: 30, Height: 15,
				ContentHeight: 100,
				Clip:          &render.Clip{Top: 0, Left: 0, Bottom: 14, Right: 29},
				Children: []*Box{
					{
						X: 2, Y: 2, Width: 20, Height: 10,
						ContentHeight: 50,
						Clip:          &render.Clip{Top: 2, Left: 2, Bottom: 11, Right: 21},
					},
				},
			},
		},
	}

	// When — hit test at a point inside both boxes
	idx := HitTestScroll(tree, 5, 5)

	// Then — inner box wins (index 1, after outer at index 0)
	if idx != 1 {
		t.Errorf("expected hit index 1 (deepest), got %d", idx)
	}
}
