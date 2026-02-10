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
