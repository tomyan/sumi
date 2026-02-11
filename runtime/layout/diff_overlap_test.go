package layout

import "testing"

func TestHasOverlappingElementsDetectsPositioned(t *testing.T) {
	// Given — tree with an absolute child
	box := &Box{
		Children: []*Box{
			{Children: []*Box{
				{Position: "absolute"},
			}},
		},
	}

	// When/Then
	if !HasOverlappingElements(box) {
		t.Error("expected HasOverlappingElements to return true for absolute child")
	}
}

func TestHasOverlappingElementsDetectsFixed(t *testing.T) {
	// Given
	box := &Box{
		Children: []*Box{
			{Position: "fixed"},
		},
	}

	// When/Then
	if !HasOverlappingElements(box) {
		t.Error("expected HasOverlappingElements to return true for fixed child")
	}
}

func TestHasOverlappingElementsDetectsZIndex(t *testing.T) {
	// Given
	box := &Box{
		Children: []*Box{
			{ZIndex: 1},
		},
	}

	// When/Then
	if !HasOverlappingElements(box) {
		t.Error("expected HasOverlappingElements to return true for non-zero z-index")
	}
}

func TestHasOverlappingElementsFalseForPlainTree(t *testing.T) {
	// Given — no positioning or z-index
	box := &Box{
		Children: []*Box{
			{Content: "hello"},
			{Content: "world"},
		},
	}

	// When/Then
	if HasOverlappingElements(box) {
		t.Error("expected HasOverlappingElements to return false for plain tree")
	}
}
