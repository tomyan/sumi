package layout

import "testing"

func TestLayoutSetsHasOverlapForAbsoluteChild(t *testing.T) {
	// Given — tree with an absolute child
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindBox, Children: []*Input{
				{Kind: KindText, Content: "abs", Position: "absolute"},
			}},
		},
	}

	// When
	box := Layout(input, 20, 10)

	// Then
	if !box.HasOverlap {
		t.Error("expected HasOverlap=true for tree with absolute child")
	}
}

func TestLayoutSetsHasOverlapForFixedChild(t *testing.T) {
	// Given
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindText, Content: "fixed", Position: "fixed"},
		},
	}

	// When
	box := Layout(input, 20, 10)

	// Then
	if !box.HasOverlap {
		t.Error("expected HasOverlap=true for tree with fixed child")
	}
}

func TestLayoutSetsHasOverlapForZIndex(t *testing.T) {
	// Given
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Children: []*Input{
			{Kind: KindText, Content: "z", ZIndex: 1},
		},
	}

	// When
	box := Layout(input, 20, 10)

	// Then
	if !box.HasOverlap {
		t.Error("expected HasOverlap=true for tree with non-zero z-index")
	}
}

func TestLayoutHasOverlapFalseForPlainTree(t *testing.T) {
	// Given — no positioning or z-index
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.HasOverlap {
		t.Error("expected HasOverlap=false for plain tree")
	}
}

func TestLayoutHasOverlapPropagatesFromDeepChild(t *testing.T) {
	// Given — absolute is 3 levels deep
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  40,
		FixedHeight: 20,
		Children: []*Input{
			{Kind: KindBox, Children: []*Input{
				{Kind: KindBox, Children: []*Input{
					{Kind: KindText, Content: "deep", Position: "absolute"},
				}},
			}},
		},
	}

	// When
	box := Layout(input, 40, 20)

	// Then — HasOverlap should propagate up to root
	if !box.HasOverlap {
		t.Error("expected HasOverlap=true to propagate from deep child")
	}
}
