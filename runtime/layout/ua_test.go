package layout

import "testing"

// C1b: UA stylesheet defaults for HTML elements.

func TestUAHeadingBold(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "h1", Kind: KindText, Content: "Title"},
	}}
	ResolveStyles(tree, nil, 80, 24)
	if !tree.Children[0].Style.Bold {
		t.Errorf("h1 should be bold by default: %+v", tree.Children[0].Style)
	}
}

func TestUAListMarkers(t *testing.T) {
	// Given: <ul><li>item</li></ul> as the parser now builds it.
	li := &Input{Tag: "li", Kind: KindBox, Children: []*Input{
		{Kind: KindText, Content: "item"},
	}}
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "ul", Kind: KindBox, Children: []*Input{li}},
	}}

	// When
	ResolveStyles(tree, nil, 80, 24)

	// Then: the li grew a "• " marker child.
	if len(li.Children) != 2 {
		t.Fatalf("li children = %d, want 2 (marker + text)", len(li.Children))
	}
	if got := li.Children[0].Content; got != "• " {
		t.Errorf("marker = %q", got)
	}
}

func TestUAAuthorOverridesDefaults(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "h1", Kind: KindText, Content: "Title"},
	}}
	ss := sheet(t, `h1 { font-weight: normal; opacity: dim; }`)
	ResolveStyles(tree, ss, 80, 24)
	st := tree.Children[0].Style
	if st.Bold {
		t.Errorf("author font-weight: normal must beat UA bold")
	}
	if !st.Dim {
		t.Errorf("author dim missing: %+v", st)
	}
}

func TestUAHrRendersRule(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "hr", Kind: KindBox},
	}}
	ResolveStyles(tree, nil, 80, 24)
	lines := renderToString(tree, 10, 3)
	// Row layout: hr occupies a margined row rendered as ─ across the width.
	found := false
	for _, l := range lines {
		if l == "──────────" {
			found = true
		}
	}
	if !found {
		t.Errorf("no rule line rendered: %q", lines)
	}
}

func TestUATextLevelElements(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "strong", Kind: KindText, Content: "b"},
		{Tag: "em", Kind: KindText, Content: "i"},
		{Tag: "a", Kind: KindText, Content: "link"},
		{Tag: "del", Kind: KindText, Content: "gone"},
	}}
	ResolveStyles(tree, nil, 80, 24)
	if !tree.Children[0].Style.Bold || !tree.Children[1].Style.Italic {
		t.Error("strong/em UA styles missing")
	}
	if !tree.Children[2].Style.Underline || !tree.Children[3].Style.Strikethrough {
		t.Error("a/del UA styles missing")
	}
}
