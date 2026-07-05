package layout

import "testing"

// A3: percentage sizing resolves against the containing block's available space.

func TestWidthPctResolvesAgainstAvailableWidth(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox, WidthPct: 50, FixedHeight: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if got := box.Children[0].Width; got != 40 {
		t.Errorf("50%% of 80 = %d, want 40", got)
	}
}

func TestHeightPctResolvesAgainstAvailableHeight(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox, HeightPct: 25},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if got := box.Children[0].Height; got != 6 {
		t.Errorf("25%% of 24 = %d, want 6", got)
	}
}

func TestWidthPctNestedResolvesAgainstParentContentBox(t *testing.T) {
	// Given: parent fixed at 40 wide; child claims 50% of it.
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{
				Kind:       KindBox,
				FixedWidth: 40,
				Children: []*Input{
					{Kind: KindBox, WidthPct: 50, FixedHeight: 1},
				},
			},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if got := box.Children[0].Children[0].Width; got != 20 {
		t.Errorf("50%% of 40 = %d, want 20", got)
	}
}

func TestWidthPctHundredFillsParent(t *testing.T) {
	// Given
	input := &Input{
		Kind: KindBox,
		Children: []*Input{
			{Kind: KindBox, WidthPct: 100, FixedHeight: 1},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then
	if got := box.Children[0].Width; got != 80 {
		t.Errorf("100%% of 80 = %d, want 80", got)
	}
}

func TestParsePaddingWithCellUnits(t *testing.T) {
	// Given / When
	p := ParsePadding("1cell 2cell")

	// Then
	want := Padding{Top: 1, Right: 2, Bottom: 1, Left: 2}
	if p != want {
		t.Errorf("ParsePadding(\"1cell 2cell\") = %+v, want %+v", p, want)
	}
}

func TestParsePaddingWithChUnits(t *testing.T) {
	p := ParsePadding("1ch")
	want := Padding{Top: 1, Right: 1, Bottom: 1, Left: 1}
	if p != want {
		t.Errorf("ParsePadding(\"1ch\") = %+v, want %+v", p, want)
	}
}
