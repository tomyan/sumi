package layout

import "testing"

// B3a: justify space-around/evenly, align-self, order, reverse directions.

func rowOf(widths ...int) *Input {
	tree := &Input{Kind: KindBox, Direction: "row", FixedWidth: 20, FixedHeight: 1}
	for _, w := range widths {
		tree.Children = append(tree.Children, &Input{Kind: KindBox, FixedWidth: w, FixedHeight: 1})
	}
	return tree
}

func xs(box *Box) []int {
	var out []int
	for _, c := range box.Children {
		out = append(out, c.X)
	}
	return out
}

func TestJustifySpaceAround(t *testing.T) {
	// Given: two 4-wide items in 20: remaining 12, half-gaps → 3, 13.
	tree := rowOf(4, 4)
	tree.Justify = "space-around"
	box := Layout(tree, 40, 5)
	got := xs(box)
	if got[0] != 3 || got[1] != 13 {
		t.Errorf("xs = %v, want [3 13]", got)
	}
}

func TestJustifySpaceEvenly(t *testing.T) {
	// Given: two 4-wide in 20: remaining 12 in 3 gaps of 4 → 4, 12.
	tree := rowOf(4, 4)
	tree.Justify = "space-evenly"
	box := Layout(tree, 40, 5)
	got := xs(box)
	if got[0] != 4 || got[1] != 12 {
		t.Errorf("xs = %v, want [4 12]", got)
	}
}

func TestAlignSelfOverridesParentAlign(t *testing.T) {
	// Given: row with fixed height 5, parent align start, one child align-self end.
	tree := &Input{Kind: KindBox, Direction: "row", FixedHeight: 5, Align: "start", Children: []*Input{
		{Kind: KindBox, FixedWidth: 2, FixedHeight: 1},
		{Kind: KindBox, FixedWidth: 2, FixedHeight: 1, AlignSelf: "end"},
	}}
	box := Layout(tree, 40, 10)
	if got := box.Children[0].Y; got != 0 {
		t.Errorf("first Y = %d, want 0", got)
	}
	if got := box.Children[1].Y; got != 4 {
		t.Errorf("align-self end Y = %d, want 4", got)
	}
}

func TestOrderReordersPlacement(t *testing.T) {
	// Given: three items, the first pushed last via order.
	tree := &Input{Kind: KindBox, Direction: "row", Children: []*Input{
		{Kind: KindBox, FixedWidth: 2, FixedHeight: 1, Order: 1},
		{Kind: KindBox, FixedWidth: 3, FixedHeight: 1},
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1},
	}}
	box := Layout(tree, 40, 5)
	// box.Children keeps SOURCE order; placement moves the ordered child.
	if got := box.Children[0].X; got != 7 {
		t.Errorf("order:1 child X = %d, want 7 (placed after 3+4)", got)
	}
	if got := box.Children[1].X; got != 0 {
		t.Errorf("second child X = %d, want 0", got)
	}
}

func TestRowReversePacksRight(t *testing.T) {
	// Given: fixed 20-wide row-reverse with two items 4 and 6.
	tree := &Input{Kind: KindBox, Direction: "row-reverse", FixedWidth: 20, FixedHeight: 1, Children: []*Input{
		{Kind: KindBox, FixedWidth: 4, FixedHeight: 1},
		{Kind: KindBox, FixedWidth: 6, FixedHeight: 1},
	}}
	box := Layout(tree, 40, 5)
	// Source-first item hugs the right edge; source-second sits left of it.
	if got := box.Children[0].X; got != 16 {
		t.Errorf("first child X = %d, want 16", got)
	}
	if got := box.Children[1].X; got != 10 {
		t.Errorf("second child X = %d, want 10", got)
	}
}

func TestColumnReverse(t *testing.T) {
	tree := &Input{Kind: KindBox, Direction: "column-reverse", FixedHeight: 6, Children: []*Input{
		{Kind: KindText, Content: "a"},
		{Kind: KindText, Content: "b"},
	}}
	box := Layout(tree, 20, 10)
	if got := box.Children[0].Y; got != 5 {
		t.Errorf("first child Y = %d, want 5 (bottom)", got)
	}
	if got := box.Children[1].Y; got != 4 {
		t.Errorf("second child Y = %d, want 4", got)
	}
}

// B3b: flex-shrink, flex-basis, flex shorthand.

func TestFlexShrinkDefaultCompressesOverflow(t *testing.T) {
	// Given: two basis-10 children in a 12-wide row: deficit 8 shared evenly.
	tree := &Input{Kind: KindBox, Direction: "row", FixedWidth: 12, FixedHeight: 1, Children: []*Input{
		{Kind: KindBox, FlexBasis: "10", FixedHeight: 1},
		{Kind: KindBox, FlexBasis: "10", FixedHeight: 1},
	}}
	box := Layout(tree, 40, 5)
	if got := box.Children[0].Width + box.Children[1].Width; got != 12 {
		t.Errorf("total width = %d, want 12 (shrunk to fit)", got)
	}
}

func TestFlexShrinkZeroKeepsSize(t *testing.T) {
	// Given: same overflow, but the first child refuses to shrink.
	tree := &Input{Kind: KindBox, Direction: "row", FixedWidth: 12, FixedHeight: 1, Children: []*Input{
		{Kind: KindBox, FlexBasis: "10", FlexShrink: -1, FixedHeight: 1},
		{Kind: KindBox, FlexBasis: "10", FixedHeight: 1},
	}}
	box := Layout(tree, 40, 5)
	if got := box.Children[0].Width; got != 10 {
		t.Errorf("no-shrink child width = %d, want 10", got)
	}
	if got := box.Children[1].Width; got != 2 {
		t.Errorf("shrinking child width = %d, want 2", got)
	}
}

func TestFlexBasisSetsNaturalSize(t *testing.T) {
	tree := &Input{Kind: KindBox, Direction: "row", FixedWidth: 30, FixedHeight: 1, Children: []*Input{
		{Kind: KindBox, FlexBasis: "8", FixedHeight: 1},
		{Kind: KindBox, FlexBasis: "50%", FixedHeight: 1},
	}}
	box := Layout(tree, 40, 5)
	if got := box.Children[0].Width; got != 8 {
		t.Errorf("basis 8 width = %d, want 8", got)
	}
	if got := box.Children[1].Width; got != 15 {
		t.Errorf("basis 50%% width = %d, want 15 (of 30)", got)
	}
}

func TestFlexOneEqualSplitRegardlessOfContent(t *testing.T) {
	// Given: flex: 1 semantics — grow 1, basis 0 — content length irrelevant.
	tree := &Input{Kind: KindBox, Direction: "row", FixedWidth: 20, FixedHeight: 1, Children: []*Input{
		{Kind: KindBox, FlexGrow: 1, FlexBasis: "0", FixedHeight: 1,
			Children: []*Input{{Kind: KindText, Content: "long content here"}}},
		{Kind: KindBox, FlexGrow: 1, FlexBasis: "0", FixedHeight: 1,
			Children: []*Input{{Kind: KindText, Content: "x"}}},
	}}
	box := Layout(tree, 40, 5)
	if box.Children[0].Width != box.Children[1].Width {
		t.Errorf("flex:1 children unequal: %d vs %d", box.Children[0].Width, box.Children[1].Width)
	}
}
