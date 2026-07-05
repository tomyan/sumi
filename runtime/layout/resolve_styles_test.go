package layout

import (
	"testing"

	"github.com/tomyan/sumi/parser/style"
)

// RS2: runtime style resolution against the input tree.

func sheet(t *testing.T, src string) *style.Stylesheet {
	t.Helper()
	ss, err := style.Parse(src)
	if err != nil {
		t.Fatalf("stylesheet: %v", err)
	}
	return ss
}

func TestResolveStylesAppliesVisualAndLayout(t *testing.T) {
	// Given
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"panel"}, Kind: KindBox, Children: []*Input{
			{Tag: "text", Kind: KindText, Content: "hi"},
		}},
	}}
	ss := sheet(t, `.panel { border: single; padding: 1 2; } .panel text { color: red; }`)

	// When
	ResolveStyles(tree, ss, 80, 24)

	// Then
	panel := tree.Children[0]
	if panel.Border != "single" || panel.Padding.Right != 2 {
		t.Errorf("panel = border %q padding %+v", panel.Border, panel.Padding)
	}
	if got := panel.Children[0].Style.FG.Name; got != "red" {
		t.Errorf("descendant text FG = %q, want red", got)
	}
}

func TestResolveStylesInlineAttrWins(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"a"}, Kind: KindBox,
			Attrs: map[string]string{"border": "double"}, Border: "double"},
	}}
	ss := sheet(t, `.a { border: single; }`)
	ResolveStyles(tree, ss, 80, 24)
	if got := tree.Children[0].Border; got != "double" {
		t.Errorf("border = %q, inline attr must win", got)
	}
}

func TestResolveStylesRuntimeSiblingsInForContent(t *testing.T) {
	// Given: three runtime children, as a {for} loop would build them —
	// static analysis could not know these siblings.
	items := []*Input{
		{Tag: "text", Kind: KindText, Content: "a"},
		{Tag: "text", Kind: KindText, Content: "b"},
		{Tag: "text", Kind: KindText, Content: "c"},
	}
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"list"}, Kind: KindBox, Children: items},
	}}
	ss := sheet(t, `.list text:nth-child(odd) { color: red; }`)

	// When
	ResolveStyles(tree, ss, 80, 24)

	// Then: 1st and 3rd stripe, 2nd doesn't.
	if items[0].Style.FG.Name != "red" || items[2].Style.FG.Name != "red" {
		t.Errorf("odd items should stripe: %+v %+v", items[0].Style, items[2].Style)
	}
	if items[1].Style.FG.Name == "red" {
		t.Errorf("even item must not stripe: %+v", items[1].Style)
	}
}

func TestResolveStylesSkipsComponentSubtrees(t *testing.T) {
	// Given: a spliced child component (its root carries Tag "root")
	compRoot := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "text", Kind: KindText, Content: "inner"},
	}}
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{compRoot}}
	ss := sheet(t, `text { color: red; }`)

	// When
	ResolveStyles(tree, ss, 80, 24)

	// Then: the component's text is styled by ITS stylesheet, not the parent's.
	if got := compRoot.Children[0].Style.FG.Name; got == "red" {
		t.Error("parent stylesheet must not leak into component subtree")
	}
}

func TestResolveStylesHoverAndFocus(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"btn"}, Kind: KindBox},
	}}
	ss := sheet(t, `.btn:hover { background: cyan; } .btn:focus { border-color: green; }`)
	ResolveStyles(tree, ss, 80, 24)
	btn := tree.Children[0]
	if btn.HoverStyle.BG.Name != "cyan" {
		t.Errorf("HoverStyle = %+v", btn.HoverStyle)
	}
	if btn.FocusStyle.FG.Name != "green" {
		t.Errorf("FocusStyle = %+v", btn.FocusStyle)
	}
}

func TestResolveStylesNilStylesheetNoop(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox}
	ResolveStyles(tree, nil, 80, 24) // must not panic
}

// A11: var() custom properties inherit down the tree.

func TestResolveStylesVarInheritance(t *testing.T) {
	// Given: a theme variable on root, used two levels down.
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Kind: KindBox, Children: []*Input{
			{Tag: "text", Classes: []string{"accented"}, Kind: KindText},
		}},
	}}
	ss := sheet(t, `root { --accent: cyan; } .accented { color: var(--accent); }`)

	// When
	ResolveStyles(tree, ss, 80, 24)

	// Then
	if got := tree.Children[0].Children[0].Style.FG.Name; got != "cyan" {
		t.Errorf("FG = %q, want cyan via inherited var", got)
	}
}

func TestResolveStylesVarShadowing(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"override"}, Kind: KindBox, Children: []*Input{
			{Tag: "text", Classes: []string{"accented"}, Kind: KindText},
		}},
		{Tag: "text", Classes: []string{"accented"}, Kind: KindText},
	}}
	ss := sheet(t, `root { --accent: cyan; } .override { --accent: magenta; } .accented { color: var(--accent); }`)
	ResolveStyles(tree, ss, 80, 24)
	if got := tree.Children[0].Children[0].Style.FG.Name; got != "magenta" {
		t.Errorf("shadowed FG = %q, want magenta", got)
	}
	if got := tree.Children[1].Style.FG.Name; got != "cyan" {
		t.Errorf("root-scope FG = %q, want cyan", got)
	}
}

func TestResolveStylesVarInLayoutProp(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"sized"}, Kind: KindBox},
	}}
	ss := sheet(t, `root { --panel-width: 30; } .sized { width: var(--panel-width); }`)
	ResolveStyles(tree, ss, 80, 24)
	if got := tree.Children[0].FixedWidth; got != 30 {
		t.Errorf("FixedWidth = %d, want 30 via var", got)
	}
}

func TestResolveStylesUnresolvedVarDropsProperty(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "text", Classes: []string{"x"}, Kind: KindText},
	}}
	ss := sheet(t, `.x { color: var(--nope); }`)
	ResolveStyles(tree, ss, 80, 24)
	if got := tree.Children[0].Style.FG.Name; got != "" {
		t.Errorf("FG = %q, want unset", got)
	}
}

// A12: CSS math functions in layout properties.

func TestResolveStylesCalcWithoutPercent(t *testing.T) {
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"x"}, Kind: KindBox},
	}}
	ss := sheet(t, `.x { width: calc(10 + 5); gap: min(3, 2); }`)
	ResolveStyles(tree, ss, 80, 24)
	if got := tree.Children[0].FixedWidth; got != 15 {
		t.Errorf("FixedWidth = %d, want 15", got)
	}
	if got := tree.Children[0].Gap; got != 2 {
		t.Errorf("Gap = %d, want 2", got)
	}
}

func TestResolveStylesCalcWithPercentDefersToLayout(t *testing.T) {
	// Given: calc against the containing block.
	tree := &Input{Tag: "root", Kind: KindBox, Children: []*Input{
		{Tag: "box", Classes: []string{"x"}, Kind: KindBox, Children: []*Input{
			{Tag: "text", Kind: KindText, Content: "hi"},
		}},
	}}
	ss := sheet(t, `.x { width: calc(100% - 10); height: 1; }`)

	// When: resolve then layout at 80 cols.
	ResolveStyles(tree, ss, 80, 24)
	box := Layout(tree, 80, 24)

	// Then
	if got := tree.Children[0].WidthCalc; got != "calc(100% - 10)" {
		t.Errorf("WidthCalc = %q", got)
	}
	if got := box.Children[0].Width; got != 70 {
		t.Errorf("laid-out width = %d, want 70", got)
	}
}
