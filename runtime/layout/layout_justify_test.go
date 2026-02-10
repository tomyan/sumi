package layout

import (
	"testing"
)

func TestLayoutJustifyStartIsDefault(t *testing.T) {
	// Given a fixed-width row with no justify (default = start)
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then child is at the start (X=0)
	if box.Children[0].X != 0 {
		t.Errorf("child X = %d, want 0 (start)", box.Children[0].X)
	}
}

func TestLayoutJustifyEndRow(t *testing.T) {
	// Given a fixed-width row with justify=end
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Justify:    "end",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then child is at the end: 40 - 5 = 35
	if box.Children[0].X != 35 {
		t.Errorf("child X = %d, want 35 (end)", box.Children[0].X)
	}
}

func TestLayoutJustifyCenterRow(t *testing.T) {
	// Given a fixed-width row with justify=center
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Justify:    "center",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are centered: (40 - 10) / 2 = 15
	if box.Children[0].X != 15 {
		t.Errorf("first child X = %d, want 15 (centered)", box.Children[0].X)
	}
	if box.Children[1].X != 20 {
		t.Errorf("second child X = %d, want 20 (centered)", box.Children[1].X)
	}
}

func TestLayoutJustifySpaceBetweenRow(t *testing.T) {
	// Given a fixed-width row with justify=space-between
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Justify:    "space-between",
		Children: []*Input{
			{Kind: KindText, Content: "aaa"},
			{Kind: KindText, Content: "bbb"},
			{Kind: KindText, Content: "ccc"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then: total content = 9, remaining = 31, 2 gaps of 15 and 16 (integer division)
	// First at 0, gaps = 31/2 = 15 each (with remainder distributed)
	// Actually: remaining=31, gaps=2, perGap=15, remainder=1
	// first=0, second=0+3+15=18, third=18+3+16=37
	if box.Children[0].X != 0 {
		t.Errorf("child[0].X = %d, want 0", box.Children[0].X)
	}
	// space-between: gaps = remaining / (n-1) = 31/2 = 15, remainder = 1
	// child[1].X = 3 + 15 = 18
	if box.Children[1].X != 18 {
		t.Errorf("child[1].X = %d, want 18", box.Children[1].X)
	}
	// child[2].X = 40 - 3 = 37
	if box.Children[2].X != 37 {
		t.Errorf("child[2].X = %d, want 37", box.Children[2].X)
	}
}

func TestLayoutJustifyEndColumn(t *testing.T) {
	// Given a fixed-height column with justify=end
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedHeight: 20,
		Justify:     "end",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are pushed to the end: offset = 20 - 2 = 18
	if box.Children[0].Y != 18 {
		t.Errorf("child[0].Y = %d, want 18 (end)", box.Children[0].Y)
	}
	if box.Children[1].Y != 19 {
		t.Errorf("child[1].Y = %d, want 19 (end)", box.Children[1].Y)
	}
}

func TestLayoutJustifyCenterColumn(t *testing.T) {
	// Given a fixed-height column with justify=center
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedHeight: 20,
		Justify:     "center",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then children are centered: (20 - 2) / 2 = 9
	if box.Children[0].Y != 9 {
		t.Errorf("child[0].Y = %d, want 9 (centered)", box.Children[0].Y)
	}
	if box.Children[1].Y != 10 {
		t.Errorf("child[1].Y = %d, want 10 (centered)", box.Children[1].Y)
	}
}

func TestLayoutJustifySpaceBetweenColumn(t *testing.T) {
	// Given a fixed-height column with justify=space-between
	input := &Input{
		Kind:        KindBox,
		Direction:   "column",
		FixedHeight: 10,
		Justify:     "space-between",
		Children: []*Input{
			{Kind: KindText, Content: "top"},
			{Kind: KindText, Content: "bottom"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then: first at 0, last at 10-1=9
	if box.Children[0].Y != 0 {
		t.Errorf("child[0].Y = %d, want 0", box.Children[0].Y)
	}
	if box.Children[1].Y != 9 {
		t.Errorf("child[1].Y = %d, want 9", box.Children[1].Y)
	}
}

func TestLayoutJustifyWithPaddingAndBorder(t *testing.T) {
	// Given justify=center with padding and border
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Padding:    Padding{0, 2, 0, 2},
		Border:     "single",
		Justify:    "center",
		Children: []*Input{
			{Kind: KindText, Content: "hi"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then: content area = 40 - 2*border - 2*pad = 34, center = (34-2)/2 = 16
	// offsetX = 1 + 2 = 3, so child X = 3 + 16 = 19
	if box.Children[0].X != 19 {
		t.Errorf("child X = %d, want 19 (centered in content area)", box.Children[0].X)
	}
}

func TestLayoutJustifySingleChildSpaceBetween(t *testing.T) {
	// Given space-between with a single child, behaves like start
	input := &Input{
		Kind:       KindBox,
		Direction:  "row",
		FixedWidth: 40,
		Justify:    "space-between",
		Children: []*Input{
			{Kind: KindText, Content: "only"},
		},
	}

	// When
	box := Layout(input, 80, 24)

	// Then single child at start
	if box.Children[0].X != 0 {
		t.Errorf("child X = %d, want 0 (single child, space-between = start)", box.Children[0].X)
	}
}

func TestLayoutJustifyCenterColumnAutoHeight(t *testing.T) {
	// Given a column with justify=center but NO fixed height,
	// the container auto-sizes to its content — no free space to distribute.
	// Justify should have no effect (CSS: auto main-size means zero free space).
	input := &Input{
		Kind:    KindBox,
		Justify: "center",
		Border:  "single",
		Padding: Padding{0, 2, 0, 2},
		Children: []*Input{
			{Kind: KindText, Content: "Centered Title"},
		},
	}

	// When laid out in 80x50 terminal
	box := Layout(input, 80, 50)

	// Then the text stays at the top of the content area (offsetY=1 for border)
	if box.Children[0].Y != 1 {
		t.Errorf("child Y = %d, want 1 (no centering for auto-height column)", box.Children[0].Y)
	}
	// The box auto-sizes to content: 1 text + 2 border = 3
	if box.Height != 3 {
		t.Errorf("box Height = %d, want 3 (auto-sized, not inflated by justify)", box.Height)
	}
}

func TestLayoutJustifyEndColumnAutoHeight(t *testing.T) {
	// Given a column with justify=end but NO fixed height,
	// justify should have no effect — auto-sized container has no free space.
	input := &Input{
		Kind:    KindBox,
		Justify: "end",
		Children: []*Input{
			{Kind: KindText, Content: "hello"},
			{Kind: KindText, Content: "world"},
		},
	}

	// When
	box := Layout(input, 80, 40)

	// Then children stay at their natural positions
	if box.Children[0].Y != 0 {
		t.Errorf("child[0].Y = %d, want 0", box.Children[0].Y)
	}
	if box.Children[1].Y != 1 {
		t.Errorf("child[1].Y = %d, want 1", box.Children[1].Y)
	}
	if box.Height != 2 {
		t.Errorf("box Height = %d, want 2 (auto-sized)", box.Height)
	}
}
