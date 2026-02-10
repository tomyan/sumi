package layout

import "testing"

func TestOverflowScrollUnboundedChildLayout(t *testing.T) {
	// Given — a box with overflow=scroll, fixed height=5, with 20 children
	children := make([]*Input, 20)
	for i := range children {
		children[i] = &Input{Kind: KindText, Content: "line"}
	}
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Overflow:    "scroll",
		Children:    children,
	}

	// When
	box := Layout(input, 80, 24)

	// Then — viewport height is fixed at 5
	if box.Height != 5 {
		t.Errorf("Height = %d, want 5 (viewport)", box.Height)
	}
	// Content height should be 20 (one line per child)
	if box.ContentHeight != 20 {
		t.Errorf("ContentHeight = %d, want 20", box.ContentHeight)
	}
	// All 20 children should be laid out
	if len(box.Children) != 20 {
		t.Errorf("len(Children) = %d, want 20", len(box.Children))
	}
	// Last child should be at Y=19 (relative to content area)
	if box.Children[19].Y != 19 {
		t.Errorf("child[19].Y = %d, want 19", box.Children[19].Y)
	}
}

func TestOverflowScrollContentWidthTracked(t *testing.T) {
	// Given — scroll box with children of varying widths
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  10,
		FixedHeight: 3,
		Overflow:    "scroll",
		Children: []*Input{
			{Kind: KindText, Content: "short"},
			{Kind: KindText, Content: "a longer line here"},
			{Kind: KindText, Content: "medium text"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — content width tracks the widest child
	if box.ContentWidth < 10 {
		t.Errorf("ContentWidth = %d, want >= 10", box.ContentWidth)
	}
}

func TestOverflowScrollSetsClip(t *testing.T) {
	// Given
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Overflow:    "scroll",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
			{Kind: KindText, Content: "line 3"},
			{Kind: KindText, Content: "line 4"},
			{Kind: KindText, Content: "line 5"},
			{Kind: KindText, Content: "line 6"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — clip should be set
	if box.Clip == nil {
		t.Fatal("expected Clip to be set on overflow:scroll box")
	}
	if box.Clip.Bottom != 4 {
		t.Errorf("Clip.Bottom = %d, want 4", box.Clip.Bottom)
	}
}

func TestOverflowScrollWithBorder(t *testing.T) {
	// Given — bordered scroll box
	children := make([]*Input, 10)
	for i := range children {
		children[i] = &Input{Kind: KindText, Content: "line"}
	}
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Overflow:    "scroll",
		Border:      "single",
		Children:    children,
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if box.Height != 5 {
		t.Errorf("Height = %d, want 5", box.Height)
	}
	// Content height = 10 children, each 1 line
	if box.ContentHeight != 10 {
		t.Errorf("ContentHeight = %d, want 10", box.ContentHeight)
	}
	// Clip should be inside the border
	if box.Clip == nil {
		t.Fatal("expected Clip to be set")
	}
	if box.Clip.Top != 1 {
		t.Errorf("Clip.Top = %d, want 1 (inside border)", box.Clip.Top)
	}
}

func TestNoOverflowContentHeightIsZero(t *testing.T) {
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

	// Then — ContentHeight should be 0 (not tracked)
	if box.ContentHeight != 0 {
		t.Errorf("ContentHeight = %d, want 0 for non-overflow box", box.ContentHeight)
	}
}

func TestOverflowScrollContentFitsNoExtraHeight(t *testing.T) {
	// Given — content fits within viewport
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Overflow:    "scroll",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — ContentHeight is the actual content height (2), not viewport
	if box.ContentHeight != 2 {
		t.Errorf("ContentHeight = %d, want 2", box.ContentHeight)
	}
}
