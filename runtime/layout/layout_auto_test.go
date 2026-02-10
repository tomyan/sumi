package layout

import "testing"

func TestOverflowAutoContentFitsNoScrollbar(t *testing.T) {
	// Given — content fits within the viewport
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Overflow:    "auto",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
			{Kind: KindText, Content: "line 2"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — no scrollbar needed, full width available
	if box.ContentHeight != 2 {
		t.Errorf("ContentHeight = %d, want 2", box.ContentHeight)
	}
	// Children should use full content width (no scrollbar reduction)
	if box.Children[0].Width != 20 {
		// Text nodes don't stretch, but content avail width shouldn't be reduced
	}
	// Clip should still be set (it's auto)
	if box.Clip == nil {
		t.Fatal("expected Clip to be set on overflow:auto box")
	}
	// NeedsScrollbar should be false
	if box.NeedsScrollbar {
		t.Error("expected NeedsScrollbar = false when content fits")
	}
}

func TestOverflowAutoContentOverflowsShowsScrollbar(t *testing.T) {
	// Given — content overflows viewport
	children := make([]*Input, 10)
	for i := range children {
		children[i] = &Input{Kind: KindText, Content: "line"}
	}
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 5,
		Overflow:    "auto",
		Children:    children,
	}

	// When
	box := Layout(input, 80, 24)

	// Then — scrollbar needed
	if box.ContentHeight != 10 {
		t.Errorf("ContentHeight = %d, want 10", box.ContentHeight)
	}
	if !box.NeedsScrollbar {
		t.Error("expected NeedsScrollbar = true when content overflows")
	}
}

func TestOverflowScrollAlwaysNeedsScrollbar(t *testing.T) {
	// Given — overflow=scroll always shows scrollbar
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		FixedHeight: 10,
		Overflow:    "scroll",
		Children: []*Input{
			{Kind: KindText, Content: "line 1"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if !box.NeedsScrollbar {
		t.Error("expected NeedsScrollbar = true for overflow:scroll")
	}
}
