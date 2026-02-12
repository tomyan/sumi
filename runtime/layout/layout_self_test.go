package layout

import "testing"

func TestSelfWWritesComputedWidth(t *testing.T) {
	// Given
	selfW := 0
	input := &Input{
		Kind:       KindBox,
		FixedWidth: 30,
		SelfW:      &selfW,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	Layout(input, 80, 24)

	// Then
	if selfW != 30 {
		t.Errorf("SelfW: got %d, want 30", selfW)
	}
}

func TestSelfHWritesComputedHeight(t *testing.T) {
	// Given
	selfH := 0
	input := &Input{
		Kind:        KindBox,
		FixedHeight: 10,
		SelfH:       &selfH,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	Layout(input, 80, 24)

	// Then
	if selfH != 10 {
		t.Errorf("SelfH: got %d, want 10", selfH)
	}
}

func TestSelfWAutoSizing(t *testing.T) {
	// Given - no fixed width, auto-sizes to content
	selfW := 0
	input := &Input{
		Kind:  KindBox,
		SelfW: &selfW,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	Layout(input, 80, 24)

	// Then - auto-sizes to content width (5)
	if selfW != 5 {
		t.Errorf("SelfW: got %d, want 5", selfW)
	}
}

func TestSelfNilPointerDoesNotCrash(t *testing.T) {
	// Given - SelfW and SelfH are nil (default)
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When - should not panic
	box := Layout(input, 80, 24)

	// Then
	if box.Width != 5 {
		t.Errorf("width: got %d, want 5", box.Width)
	}
}

func TestSelfWCapturesStretchedWidth(t *testing.T) {
	// Given — a child box with SelfW inside a column parent (align: stretch is default)
	selfW := 0
	input := &Input{
		Kind:       KindBox,
		FixedWidth: 60,
		Children: []*Input{
			{
				Kind:  KindBox,
				SelfW: &selfW,
				Children: []*Input{
					{Kind: KindText, Content: "hi"},
				},
			},
		},
	}

	// When
	Layout(input, 80, 24)

	// Then — child should be stretched to parent's content width (60), not intrinsic width (2)
	if selfW != 60 {
		t.Errorf("SelfW: got %d, want 60 (stretched)", selfW)
	}
}

func TestSelfWAndSelfHTogether(t *testing.T) {
	// Given
	selfW := 0
	selfH := 0
	input := &Input{
		Kind:        KindBox,
		FixedWidth:  40,
		FixedHeight: 12,
		SelfW:       &selfW,
		SelfH:       &selfH,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	Layout(input, 80, 24)

	// Then
	if selfW != 40 {
		t.Errorf("SelfW: got %d, want 40", selfW)
	}
	if selfH != 12 {
		t.Errorf("SelfH: got %d, want 12", selfH)
	}
}
