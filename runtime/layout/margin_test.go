package layout

import "testing"

// B1: margins in flex flow.

func TestParseMarginForms(t *testing.T) {
	cases := []struct {
		in   string
		want Margin
	}{
		{"1", Margin{Top: 1, Right: 1, Bottom: 1, Left: 1}},
		{"1 2", Margin{Top: 1, Bottom: 1, Right: 2, Left: 2}},
		{"1 2 3 4", Margin{Top: 1, Right: 2, Bottom: 3, Left: 4}},
		{"0 auto", Margin{AutoLeft: true, AutoRight: true}},
		{"1cell 2ch", Margin{Top: 1, Bottom: 1, Right: 2, Left: 2}},
	}
	for _, c := range cases {
		if got := ParseMargin(c.in); got != c.want {
			t.Errorf("ParseMargin(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestColumnFlowMarginsSpaceChildren(t *testing.T) {
	// Given: two 1-high texts, the second with margin 2 0.
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindText, Content: "a"},
		{Kind: KindText, Content: "b", Margin: Margin{Top: 2, Bottom: 1}},
	}}

	// When
	box := Layout(tree, 20, 20)

	// Then
	if got := box.Children[1].Y; got != 3 {
		t.Errorf("second child Y = %d, want 3 (1 content + 2 margin-top)", got)
	}
}

func TestRowFlowMarginsSpaceChildren(t *testing.T) {
	tree := &Input{Kind: KindBox, Direction: "row", Children: []*Input{
		{Kind: KindText, Content: "aa"},
		{Kind: KindText, Content: "bb", Margin: Margin{Left: 3}},
	}}
	box := Layout(tree, 20, 5)
	if got := box.Children[1].X; got != 5 {
		t.Errorf("second child X = %d, want 5 (2 content + 3 margin-left)", got)
	}
}

func TestMarginAutoCentresHorizontally(t *testing.T) {
	// Given: a fixed-width child with margin: 0 auto in a column parent.
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindBox, FixedWidth: 10, FixedHeight: 1,
			Margin: Margin{AutoLeft: true, AutoRight: true}},
	}}

	// When
	box := Layout(tree, 40, 10)

	// Then
	if got := box.Children[0].X; got != 15 {
		t.Errorf("X = %d, want 15 ((40-10)/2)", got)
	}
}

func TestFlexRowMarginsReduceDistributedSpace(t *testing.T) {
	// Given: one flex child plus a fixed child with horizontal margins.
	tree := &Input{Kind: KindBox, Direction: "row", FixedWidth: 20, FixedHeight: 1, Children: []*Input{
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1, Margin: Margin{Left: 2, Right: 2}},
		{Kind: KindBox, FlexGrow: 1, FixedHeight: 1},
	}}

	// When
	box := Layout(tree, 40, 5)

	// Then: flex child gets 20 - 4 - 2 - 2 = 12.
	if got := box.Children[1].Width; got != 12 {
		t.Errorf("flex width = %d, want 12", got)
	}
	if got := box.Children[1].X; got != 8 {
		t.Errorf("flex X = %d, want 8 (2 + 4 + 2)", got)
	}
}

// B2: min/max sizing and box-sizing.

func TestMaxWidthClampsContent(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindBox, MaxWidth: 10, Children: []*Input{
			{Kind: KindText, Content: "a very long line of text"},
		}},
	}}
	box := Layout(tree, 60, 10)
	if got := box.Children[0].Width; got != 10 {
		t.Errorf("width = %d, want clamped to 10", got)
	}
}

func TestMinHeightExpands(t *testing.T) {
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindBox, MinHeight: 5, Children: []*Input{
			{Kind: KindText, Content: "one"},
		}},
	}}
	box := Layout(tree, 60, 20)
	if got := box.Children[0].Height; got != 5 {
		t.Errorf("height = %d, want 5", got)
	}
}

func TestMaxHeightLimitsStretch(t *testing.T) {
	// Given: row parent with fixed height, stretch default, child capped.
	tree := &Input{Kind: KindBox, Direction: "row", FixedHeight: 10, Children: []*Input{
		{Kind: KindBox, MaxHeight: 4},
	}}
	box := Layout(tree, 40, 20)
	if got := box.Children[0].Height; got != 4 {
		t.Errorf("height = %d, want 4 (stretch clamped)", got)
	}
}

func TestContentBoxSizingAddsInsets(t *testing.T) {
	// Given: content-box with border and padding: outer = 10 + 2 border + 2 padding.
	tree := &Input{Kind: KindBox, Children: []*Input{
		{Kind: KindBox, FixedWidth: 10, FixedHeight: 3, ContentBox: true,
			Border: "single", Padding: Padding{Left: 1, Right: 1}},
	}}
	box := Layout(tree, 60, 20)
	if got := box.Children[0].Width; got != 14 {
		t.Errorf("width = %d, want 14 (10 content + 2 border + 2 padding)", got)
	}
}
