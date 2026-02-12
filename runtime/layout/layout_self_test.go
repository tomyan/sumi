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

func TestSelfXWritesAbsolutePosition(t *testing.T) {
	// Given — a nested box with border and padding
	var selfX int
	input := &Input{
		Kind:    KindBox,
		Border:  "single",
		Padding: ParsePadding("0 2"),
		Children: []*Input{
			{Kind: KindBox, SelfX: &selfX, Children: []*Input{
				{Kind: KindText, Content: "hello"},
			}},
		},
	}

	// When
	Layout(input, 40, 10)

	// Then — selfX = border(1) + padding.Left(2) = 3
	if selfX != 3 {
		t.Errorf("SelfX = %d, want 3", selfX)
	}
}

func TestSelfYWritesAbsolutePosition(t *testing.T) {
	// Given — second child in a column with border
	var selfY int
	input := &Input{
		Kind:   KindBox,
		Border: "single",
		Children: []*Input{
			{Kind: KindText, Content: "first"},
			{Kind: KindBox, SelfY: &selfY, Children: []*Input{
				{Kind: KindText, Content: "second"},
			}},
		},
	}

	// When
	Layout(input, 40, 10)

	// Then — selfY = border(1) + first_child_height(1) = 2
	if selfY != 2 {
		t.Errorf("SelfY = %d, want 2", selfY)
	}
}

func TestSelfXNilDoesNotCrash(t *testing.T) {
	// Given — no SelfX pointer
	input := &Input{
		Kind:     KindBox,
		Children: []*Input{{Kind: KindText, Content: "hello"}},
	}

	// When — should not panic
	box := Layout(input, 40, 10)

	// Then
	if box.X != 0 {
		t.Errorf("X = %d, want 0", box.X)
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
