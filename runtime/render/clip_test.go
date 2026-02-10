package render

import "testing"

func TestClipContainsPointInside(t *testing.T) {
	// Given
	clip := &Clip{Top: 2, Left: 3, Bottom: 5, Right: 10}

	// When / Then
	tests := []struct {
		name     string
		row, col int
		want     bool
	}{
		{"top-left corner", 2, 3, true},
		{"bottom-right corner", 5, 10, true},
		{"center", 3, 6, true},
		{"above top", 1, 5, false},
		{"below bottom", 6, 5, false},
		{"left of left", 3, 2, false},
		{"right of right", 3, 11, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clip.Contains(tt.row, tt.col)
			if got != tt.want {
				t.Errorf("Contains(%d, %d) = %v, want %v", tt.row, tt.col, got, tt.want)
			}
		})
	}
}

func TestClipIntersectOverlapping(t *testing.T) {
	// Given
	a := &Clip{Top: 0, Left: 0, Bottom: 10, Right: 10}
	b := &Clip{Top: 5, Left: 5, Bottom: 15, Right: 15}

	// When
	result := a.Intersect(b)

	// Then
	if result == nil {
		t.Fatal("expected non-nil intersection")
	}
	if result.Top != 5 || result.Left != 5 || result.Bottom != 10 || result.Right != 10 {
		t.Errorf("got {%d,%d,%d,%d}, want {5,5,10,10}",
			result.Top, result.Left, result.Bottom, result.Right)
	}
}

func TestClipIntersectNoOverlap(t *testing.T) {
	// Given
	a := &Clip{Top: 0, Left: 0, Bottom: 3, Right: 3}
	b := &Clip{Top: 5, Left: 5, Bottom: 10, Right: 10}

	// When
	result := a.Intersect(b)

	// Then
	if result != nil {
		t.Errorf("expected nil intersection, got %+v", result)
	}
}

func TestClipIntersectContained(t *testing.T) {
	// Given
	outer := &Clip{Top: 0, Left: 0, Bottom: 20, Right: 20}
	inner := &Clip{Top: 5, Left: 5, Bottom: 10, Right: 10}

	// When
	result := outer.Intersect(inner)

	// Then
	if result == nil {
		t.Fatal("expected non-nil intersection")
	}
	if *result != *inner {
		t.Errorf("got %+v, want %+v", result, inner)
	}
}

func TestClipIntersectEdgeTouching(t *testing.T) {
	// Given — share a single edge row
	a := &Clip{Top: 0, Left: 0, Bottom: 5, Right: 10}
	b := &Clip{Top: 5, Left: 0, Bottom: 10, Right: 10}

	// When
	result := a.Intersect(b)

	// Then — they share row 5
	if result == nil {
		t.Fatal("expected non-nil intersection for edge-touching clips")
	}
	if result.Top != 5 || result.Bottom != 5 {
		t.Errorf("got top=%d, bottom=%d, want top=5, bottom=5", result.Top, result.Bottom)
	}
}
