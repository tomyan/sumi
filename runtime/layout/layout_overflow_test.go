package layout

import "testing"

func TestOverflowHiddenSetsClipOnBox(t *testing.T) {
	// Given — a box with overflow=hidden and fixed height
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 3,
		Overflow:    "hidden",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
			{Kind: KindText, Content: "line 3"},
			{Kind: KindText, Content: "line 4"},
			{Kind: KindText, Content: "line 5"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — clip should be set to the content area
	if box.Clip == nil {
		t.Fatal("expected Clip to be set on overflow:hidden box")
	}
	if box.Clip.Top != 0 {
		t.Errorf("Clip.Top = %d, want 0", box.Clip.Top)
	}
	if box.Clip.Left != 0 {
		t.Errorf("Clip.Left = %d, want 0", box.Clip.Left)
	}
	if box.Clip.Bottom != 2 {
		t.Errorf("Clip.Bottom = %d, want 2 (FixedHeight-1)", box.Clip.Bottom)
	}
	if box.Clip.Right != 19 {
		t.Errorf("Clip.Right = %d, want 19 (FixedWidth-1)", box.Clip.Right)
	}
}

func TestOverflowHiddenWithBorderSetsClipInsideBorder(t *testing.T) {
	// Given — a bordered box with overflow=hidden
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Overflow:    "hidden",
		Border:      "single",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
			{Kind: KindText, Content: "line 3"},
			{Kind: KindText, Content: "line 4"},
			{Kind: KindText, Content: "line 5"},
			{Kind: KindText, Content: "line 6"},
			{Kind: KindText, Content: "line 7"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — clip should be inside the border
	if box.Clip == nil {
		t.Fatal("expected Clip to be set")
	}
	// Border takes 1 cell on each side
	if box.Clip.Top != 1 {
		t.Errorf("Clip.Top = %d, want 1 (inside border)", box.Clip.Top)
	}
	if box.Clip.Left != 1 {
		t.Errorf("Clip.Left = %d, want 1 (inside border)", box.Clip.Left)
	}
	if box.Clip.Bottom != 3 {
		t.Errorf("Clip.Bottom = %d, want 3 (height-1 - border)", box.Clip.Bottom)
	}
	if box.Clip.Right != 18 {
		t.Errorf("Clip.Right = %d, want 18 (width-1 - border)", box.Clip.Right)
	}
}

func TestOverflowHiddenWithPaddingSetsClipInsidePadding(t *testing.T) {
	// Given — a box with padding and overflow=hidden
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 6,
		Overflow:    "hidden",
		Padding:     Padding{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
			{Kind: KindText, Content: "line 3"},
			{Kind: KindText, Content: "line 4"},
			{Kind: KindText, Content: "line 5"},
			{Kind: KindText, Content: "line 6"},
			{Kind: KindText, Content: "line 7"},
			{Kind: KindText, Content: "line 8"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — clip should be inside padding
	if box.Clip == nil {
		t.Fatal("expected Clip to be set")
	}
	if box.Clip.Top != 1 {
		t.Errorf("Clip.Top = %d, want 1 (padding top)", box.Clip.Top)
	}
	if box.Clip.Left != 2 {
		t.Errorf("Clip.Left = %d, want 2 (padding left)", box.Clip.Left)
	}
	if box.Clip.Bottom != 4 {
		t.Errorf("Clip.Bottom = %d, want 4 (height-1 - padding bottom)", box.Clip.Bottom)
	}
	if box.Clip.Right != 17 {
		t.Errorf("Clip.Right = %d, want 17 (width-1 - padding right)", box.Clip.Right)
	}
}

func TestNoOverflowDoesNotSetClip(t *testing.T) {
	// Given — a normal box without overflow
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Clip != nil {
		t.Errorf("expected Clip to be nil for box without overflow, got %+v", box.Clip)
	}
}

func TestOverflowHiddenChildrenStillLaidOut(t *testing.T) {
	// Given — children extend beyond the box, but all are still laid out
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 3,
		Overflow:    "hidden",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
			{Kind: KindText, Content: "line 3"},
			{Kind: KindText, Content: "line 4"},
			{Kind: KindText, Content: "line 5"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — all 5 children should be present in the tree
	if len(box.Children) != 5 {
		t.Errorf("len(Children) = %d, want 5 (all children laid out)", len(box.Children))
	}
	// Children beyond the clip are still in the tree (rendering handles clipping)
	if box.Children[4].Y != 4 {
		t.Errorf("child[4].Y = %d, want 4", box.Children[4].Y)
	}
}
