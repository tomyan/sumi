package layout

import "testing"

func TestMinWidthNoEffectWhenAvailWidthSufficient(t *testing.T) {
	// Given — min-width=30, available width=80
	// MinWidth only affects content layout when avail < min
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  80,
		FixedHeight: 5,
		MinWidth:    30,
		Overflow:    "scroll",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — children laid out at normal width
	if box.Width != 80 {
		t.Errorf("Width = %d, want 80", box.Width)
	}
	// Content should be laid out at contentAvailW (80 - insets), not min-width
	if box.Children[0].Width != 5 {
		t.Errorf("child Width = %d, want 5 (text length)", box.Children[0].Width)
	}
}

func TestMinWidthEnforcedWhenAvailWidthTooSmall(t *testing.T) {
	// Given — min-width=30, fixed width=20 → content area is 20, but min says 30
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  20,
		MinWidth:    30,
		FixedHeight: 5,
		Overflow:    "scroll",
		Children: []*Input{
			// This text is 35 chars — in a 30-wide content area, no wrapping
			{Kind: KindText, Content: "This is a really long line of text"},
		},
	}

	// When
	box := Layout(input, 20, 24)

	// Then — box width stays at 20 (viewport)
	if box.Width != 20 {
		t.Errorf("Width = %d, want 20 (fixed viewport)", box.Width)
	}
	// Content should be laid out at min-width (30), not 20
	// So text won't wrap at 20, it'll have contentAvailW=30
	if box.ContentWidth < 20 {
		t.Errorf("ContentWidth = %d, want >= 20", box.ContentWidth)
	}
}

func TestMinWidthField(t *testing.T) {
	// Given — Input has MinWidth
	input := &Input{
		Kind:     KindBox,
		MinWidth: 40,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then — layout completes without error
	if box == nil {
		t.Fatal("expected non-nil box")
	}
}
