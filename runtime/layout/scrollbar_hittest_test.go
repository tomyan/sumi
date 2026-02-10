package layout

import "testing"

func TestScrollbarHitAboveThumb(t *testing.T) {
	// Given — thumb at position 3, size 2 (rows 3-4 out of 10)
	// Click at row 1 is above thumb
	result := ScrollbarHit(1, 3, 2, 10)

	// Then
	if result != ScrollbarAboveThumb {
		t.Errorf("got %d, want ScrollbarAboveThumb", result)
	}
}

func TestScrollbarHitBelowThumb(t *testing.T) {
	// Given — thumb at position 3, size 2 (rows 3-4 out of 10)
	// Click at row 7 is below thumb
	result := ScrollbarHit(7, 3, 2, 10)

	// Then
	if result != ScrollbarBelowThumb {
		t.Errorf("got %d, want ScrollbarBelowThumb", result)
	}
}

func TestScrollbarHitOnThumb(t *testing.T) {
	// Given — thumb at position 3, size 2 (rows 3-4)
	result := ScrollbarHit(3, 3, 2, 10)

	// Then
	if result != ScrollbarOnThumb {
		t.Errorf("got %d, want ScrollbarOnThumb", result)
	}
}

func TestScrollbarHitOnThumbEnd(t *testing.T) {
	// Given — thumb at position 3, size 2 (rows 3-4)
	result := ScrollbarHit(4, 3, 2, 10)

	// Then
	if result != ScrollbarOnThumb {
		t.Errorf("got %d, want ScrollbarOnThumb", result)
	}
}

func TestScrollYFromDragPosition(t *testing.T) {
	// Given — drag at row 0 of 10, content=20, viewport=10
	scrollY := ScrollYFromDrag(0, 20, 10)

	// Then
	if scrollY != 0 {
		t.Errorf("ScrollYFromDrag(0, 20, 10) = %d, want 0", scrollY)
	}
}

func TestScrollYFromDragPositionAtBottom(t *testing.T) {
	// Given — drag at last possible position
	// trackSpace = viewport - thumbSize = 10 - 5 = 5
	// dragging to row 5 should give maxScroll = 10
	scrollY := ScrollYFromDrag(5, 20, 10)

	// Then
	if scrollY != 10 {
		t.Errorf("ScrollYFromDrag(5, 20, 10) = %d, want 10", scrollY)
	}
}

func TestScrollYFromDragPositionMiddle(t *testing.T) {
	// Given — drag at middle
	// thumbSize = 10*10/20 = 5, trackSpace = 5
	// maxScroll = 10, dragRow=2 → 2*10/5 = 4
	scrollY := ScrollYFromDrag(2, 20, 10)

	// Then
	if scrollY != 4 {
		t.Errorf("ScrollYFromDrag(2, 20, 10) = %d, want 4", scrollY)
	}
}
