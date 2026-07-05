package layout

import "testing"

// B6: CSS grid.

func TestParseTrackListForms(t *testing.T) {
	cases := []struct {
		spec  string
		avail int
		gap   int
		want  []int
	}{
		{"10 20", 40, 0, []int{10, 20}},
		{"1fr 1fr", 40, 0, []int{20, 20}},
		{"10 1fr", 40, 0, []int{10, 30}},
		{"1fr 2fr", 30, 0, []int{10, 20}},
		{"50% 1fr", 40, 0, []int{20, 20}},
		{"repeat(3, 1fr)", 30, 0, []int{10, 10, 10}},
		{"repeat(2, 5 10)", 40, 0, []int{5, 10, 5, 10}},
		{"1fr 1fr", 41, 1, []int{20, 20}},
		{"minmax(15, 1fr) 1fr", 20, 0, []int{15, 10}},
	}
	for _, c := range cases {
		got := parseTrackList(c.spec, c.avail, c.gap)
		if len(got) != len(c.want) {
			t.Errorf("%q → %v, want %v", c.spec, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q → %v, want %v", c.spec, got, c.want)
				break
			}
		}
	}
}

func gridBox(cols string, children ...*Input) *Input {
	return &Input{
		Kind: KindBox, Display: "grid", FixedWidth: 30,
		GridTemplateColumns: cols, Children: children,
	}
}

func TestGridAutoFlowPlacement(t *testing.T) {
	// Given: 3 columns of 10, four items auto-flowing.
	items := []*Input{
		{Kind: KindBox, FixedHeight: 1}, {Kind: KindBox, FixedHeight: 1},
		{Kind: KindBox, FixedHeight: 1}, {Kind: KindBox, FixedHeight: 1},
	}
	tree := &Input{Kind: KindBox, Children: []*Input{gridBox("10 10 10", items...)}}

	// When
	box := Layout(tree, 40, 10)

	// Then
	g := box.Children[0]
	wantX := []int{0, 10, 20, 0}
	wantY := []int{0, 0, 0, 1}
	for i, c := range g.Children {
		if c.X != wantX[i] || c.Y != wantY[i] {
			t.Errorf("item %d at (%d,%d), want (%d,%d)", i, c.X, c.Y, wantX[i], wantY[i])
		}
	}
}

func TestGridExplicitPlacementAndSpan(t *testing.T) {
	items := []*Input{
		{Kind: KindBox, FixedHeight: 1, GridColumn: "2", GridRow: "1"},
		{Kind: KindBox, FixedHeight: 1, GridColumn: "1 / 3", GridRow: "2"},
	}
	tree := &Input{Kind: KindBox, Children: []*Input{gridBox("10 10 10", items...)}}
	box := Layout(tree, 40, 10)
	g := box.Children[0]
	if g.Children[0].X != 10 || g.Children[0].Y != 0 {
		t.Errorf("pinned item at (%d,%d), want (10,0)", g.Children[0].X, g.Children[0].Y)
	}
	if g.Children[1].Width != 20 {
		t.Errorf("spanning item width = %d, want 20", g.Children[1].Width)
	}
	if g.Children[1].Y != 1 {
		t.Errorf("spanning item Y = %d, want 1", g.Children[1].Y)
	}
}

func TestGridSpanAutoFlow(t *testing.T) {
	items := []*Input{
		{Kind: KindBox, FixedHeight: 1, GridColumn: "span 2"},
		{Kind: KindBox, FixedHeight: 1},
	}
	tree := &Input{Kind: KindBox, Children: []*Input{gridBox("10 10 10", items...)}}
	box := Layout(tree, 40, 10)
	g := box.Children[0]
	if g.Children[0].Width != 20 {
		t.Errorf("span 2 width = %d, want 20", g.Children[0].Width)
	}
	if g.Children[1].X != 20 {
		t.Errorf("next item X = %d, want 20", g.Children[1].X)
	}
}

func TestGridTemplateAreas(t *testing.T) {
	items := []*Input{
		{Kind: KindBox, GridArea: "side"},
		{Kind: KindBox, GridArea: "main"},
	}
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "grid", FixedWidth: 30,
		GridTemplateColumns: "10 10 10",
		GridTemplateAreas:   `"side main main" "side main main"`,
		GridTemplateRows:    "2 2",
		Children:            items,
	}}}
	box := Layout(tree, 40, 10)
	g := box.Children[0]
	side, main := g.Children[0], g.Children[1]
	if side.X != 0 || side.Height != 4 {
		t.Errorf("side at X=%d H=%d, want X=0 H=4 (two 2-rows)", side.X, side.Height)
	}
	if main.X != 10 || main.Width != 20 {
		t.Errorf("main X=%d W=%d, want X=10 W=20", main.X, main.Width)
	}
}

func TestGridGapBetweenTracks(t *testing.T) {
	items := []*Input{
		{Kind: KindBox, FixedHeight: 1}, {Kind: KindBox, FixedHeight: 1},
	}
	tree := &Input{Kind: KindBox, Children: []*Input{{
		Kind: KindBox, Display: "grid", FixedWidth: 21, Gap: 1,
		GridTemplateColumns: "1fr 1fr", Children: items,
	}}}
	box := Layout(tree, 40, 10)
	g := box.Children[0]
	if g.Children[0].Width != 10 || g.Children[1].X != 11 {
		t.Errorf("w=%d x2=%d, want 10 and 11", g.Children[0].Width, g.Children[1].X)
	}
}

func TestGridImplicitRowsSizeToContent(t *testing.T) {
	items := []*Input{
		{Kind: KindText, Content: "one line"},
		{Kind: KindBox, FixedHeight: 3},
	}
	tree := &Input{Kind: KindBox, Children: []*Input{gridBox("15 15", items...)}}
	box := Layout(tree, 40, 10)
	g := box.Children[0]
	if got := g.Height; got != 3 {
		t.Errorf("grid height = %d, want 3 (tallest item in the row)", got)
	}
}
