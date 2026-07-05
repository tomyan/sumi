package layout

import "testing"

// B4g: fragment-aware hit testing — wrapped inline runs overlap as
// bounding rects; clicks resolve by actual line fragments.

func fragmentHitTree() (*Input, *Input, *Input) {
	plain := &Input{Kind: KindText, Content: "aaaa "}
	strong := &Input{Kind: KindText, Tag: "strong", Content: "bbbb cccc"}
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{plain, strong}}
	return p, plain, strong
}

func TestHitTestResolvesByFragment(t *testing.T) {
	// Given: width 10 → line0 "aaaa bbbb", line1 "cccc"; the strong
	// run's bounding rect overlaps the plain run's.
	p, plain, strong := fragmentHitTree()
	box := Layout(p, 10, 24)

	// Then: clicks land on the run whose fragment covers the cell.
	cases := []struct {
		x, y int
		want *Input
		name string
	}{
		{1, 0, plain, "plain text on line 0"},
		{6, 0, strong, "strong fragment on line 0"},
		{2, 1, strong, "strong continuation on line 1"},
		{6, 1, p, "empty cell after cccc falls to the paragraph"},
	}
	for _, c := range cases {
		path := HitTestPath(p, box, c.x, c.y)
		if len(path) == 0 {
			t.Errorf("%s: no hit at (%d,%d)", c.name, c.x, c.y)
			continue
		}
		if got := path[len(path)-1]; got != c.want {
			t.Errorf("%s: hit %v, want %v", c.name, got.Tag+got.Content, c.want.Tag+c.want.Content)
		}
	}
}

func TestHitTestInlineElementUnionIsNotSolid(t *testing.T) {
	// Given: an inline element wrapping across lines — the union rect
	// covers cells its fragments don't.
	inner := &Input{Kind: KindText, Content: "bbbb cccc"}
	em := &Input{Kind: KindBox, Tag: "em", Display: "inline", Children: []*Input{inner}}
	plain := &Input{Kind: KindText, Content: "aaaa "}
	p := &Input{Kind: KindBox, Display: "block", Children: []*Input{plain, em}}
	box := Layout(p, 10, 24)

	// Then: a cell inside em's union but on plain's fragment hits plain.
	path := HitTestPath(p, box, 1, 0)
	if got := path[len(path)-1]; got != plain {
		t.Errorf("hit %+v, want the plain run", got)
	}
	// A cell inside em's fragment hits the inner text (em on the path).
	path = HitTestPath(p, box, 6, 0)
	if got := path[len(path)-1]; got != inner {
		t.Errorf("hit %+v, want em's inner text", got)
	}
}
