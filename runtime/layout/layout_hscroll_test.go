package layout

import (
	"strings"
	"testing"
)

func TestNeedsHorizontalScrollbarWhenContentExceedsViewport(t *testing.T) {
	// Given — overflow=auto, MinWidth > available width, content fills MinWidth
	input := &Input{
		Kind:     KindBox,
		Overflow: "auto",
		MinWidth: 60,
		Children: []*Input{
			{Kind: KindText, Content: strings.Repeat("x", 60)},
		},
	}

	// When — available width is 40 (less than content width of 60)
	box := Layout(input, 40, 20)

	// Then
	if !box.NeedsHorizontalScrollbar {
		t.Errorf("expected NeedsHorizontalScrollbar=true, contentW=%d viewportW=%d", box.ContentWidth, box.Width)
	}
}

func TestNoHorizontalScrollbarWhenContentFits(t *testing.T) {
	// Given — overflow=auto, content fits in viewport
	input := &Input{
		Kind:     KindBox,
		Overflow: "auto",
		Children: []*Input{
			{Kind: KindText, Content: "Hello"},
		},
	}

	// When — available width is 40 (content width "Hello" = 5, fits easily)
	box := Layout(input, 40, 20)

	// Then
	if box.NeedsHorizontalScrollbar {
		t.Error("expected NeedsHorizontalScrollbar=false when content fits")
	}
}

func TestScrollOverflowFillsAvailableHeight(t *testing.T) {
	// Given — overflow=auto without fixed height, content is shorter than viewport
	input := &Input{
		Kind:     KindBox,
		Overflow: "auto",
		MinWidth: 60,
		Children: []*Input{
			{Kind: KindText, Content: "Short"},
		},
	}

	// When — terminal is 40x20, content is only 1 row
	box := Layout(input, 40, 20)

	// Then — box should fill the full available height (viewport)
	if box.Height != 20 {
		t.Errorf("expected box height to fill available height 20, got %d", box.Height)
	}
	// Clip bottom should be at the last row
	if box.Clip == nil {
		t.Fatal("expected clip to be set")
	}
	if box.Clip.Bottom != 19 {
		t.Errorf("expected clip bottom at 19, got %d", box.Clip.Bottom)
	}
}

func TestScrollOverflowWithFixedHeightDoesNotExpand(t *testing.T) {
	// Given — overflow=auto WITH fixed height
	input := &Input{
		Kind:        KindBox,
		Overflow:    "auto",
		FixedHeight: 10,
		MinWidth:    60,
		Children: []*Input{
			{Kind: KindText, Content: "Short"},
		},
	}

	// When — available height is 20
	box := Layout(input, 40, 20)

	// Then — box should keep fixed height, not expand
	if box.Height != 10 {
		t.Errorf("expected box height 10 (fixed), got %d", box.Height)
	}
}

func TestHorizontalScrollbarScrollAlwaysShows(t *testing.T) {
	// Given — overflow=scroll, even when content fits
	input := &Input{
		Kind:     KindBox,
		Overflow: "scroll",
		Children: []*Input{
			{Kind: KindText, Content: "Hello"},
		},
	}

	// When
	box := Layout(input, 40, 20)

	// Then — scroll always shows both scrollbars
	if !box.NeedsHorizontalScrollbar {
		t.Error("expected NeedsHorizontalScrollbar=true for overflow=scroll")
	}
}
