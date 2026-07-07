package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func mustSheet(t *testing.T, src string) *style.Stylesheet {
	t.Helper()
	ss, err := style.Parse(src)
	if err != nil {
		t.Fatalf("stylesheet: %v", err)
	}
	return ss
}

// childComponent builds a spliced child component: its root carries Tag
// "root" so the parent's cascade skips it, and it holds its own stylesheet.
func childComponent(content string, classes []string, css string, t *testing.T) *tui.Component {
	tree := &layout.Input{
		Tag: "root", Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Tag: "span", Kind: layout.KindText, Classes: classes, Content: content},
		},
	}
	return &tui.Component{Tree: tree, Stylesheet: mustSheet(t, css)}
}

func TestChildComponentStylesResolveThroughRender(t *testing.T) {
	// Given — a parent with a child; each colours its own span via a class.
	child := childComponent("X", []string{"inner"}, `.inner { color: red; }`, t)
	parent := &tui.Component{
		Tree: &layout.Input{
			Tag: "root", Kind: layout.KindBox, Direction: "column",
			CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{
				{Tag: "span", Kind: layout.KindText, Classes: []string{"outer"}, Content: "P"},
				child.Tree,
			},
		},
		Stylesheet: mustSheet(t, `.outer { color: blue; }`),
		Children:   []*tui.Component{child},
	}

	// When — render through the app.
	app := tui.TestApp(parent, 20, 3)
	buf := app.TestBuffer

	// Then — the child's class rule colours its cell through a full render.
	if got := buf.Cell(1, 0); got.Ch != 'X' || got.Style.FG.Name != "red" {
		t.Errorf("child cell = %c/%q, want X/red", got.Ch, got.Style.FG.Name)
	}
	// And the parent's own class rule colours its cell.
	if got := buf.Cell(0, 0); got.Ch != 'P' || got.Style.FG.Name != "blue" {
		t.Errorf("parent cell = %c/%q, want P/blue", got.Ch, got.Style.FG.Name)
	}
}

func TestChildComponentStylesDoNotLeakEitherDirection(t *testing.T) {
	// Given — parent and child style the same bare tag differently.
	child := childComponent("X", nil, `span { color: red; }`, t)
	parent := &tui.Component{
		Tree: &layout.Input{
			Tag: "root", Kind: layout.KindBox, Direction: "column",
			CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{
				{Tag: "span", Kind: layout.KindText, Content: "P"},
				child.Tree,
			},
		},
		Stylesheet: mustSheet(t, `span { color: blue; }`),
		Children:   []*tui.Component{child},
	}

	// When
	tui.TestApp(parent, 20, 3)

	// Then — parent's span rule stays blue on P; child's span rule stays red on X.
	// Neither direction leaks across the component boundary.
	if got := parent.Tree.Children[0].Style.FG.Name; got != "blue" {
		t.Errorf("parent span FG = %q, want blue (child rule must not leak out)", got)
	}
	if got := child.Tree.Children[0].Style.FG.Name; got != "red" {
		t.Errorf("child span FG = %q, want red (parent rule must not leak in)", got)
	}
}

func TestGrandchildComponentStylesResolve(t *testing.T) {
	// Given — a child that itself contains a grandchild component.
	grandchild := childComponent("G", []string{"deep"}, `.deep { color: green; }`, t)
	childTree := &layout.Input{
		Tag: "root", Kind: layout.KindBox, Direction: "column",
		CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{grandchild.Tree},
	}
	child := &tui.Component{
		Tree:       childTree,
		Stylesheet: mustSheet(t, `.mid { color: yellow; }`),
		Children:   []*tui.Component{grandchild},
	}
	parent := &tui.Component{
		Tree: &layout.Input{
			Tag: "root", Kind: layout.KindBox, Direction: "column",
			CursorCol: -1, CursorRow: -1,
			Children: []*layout.Input{child.Tree},
		},
		Stylesheet: mustSheet(t, `.top { color: blue; }`),
		Children:   []*tui.Component{child},
	}

	// When
	tui.TestApp(parent, 20, 3)

	// Then — the grandchild's own stylesheet resolves recursively.
	if got := grandchild.Tree.Children[0].Style.FG.Name; got != "green" {
		t.Errorf("grandchild FG = %q, want green", got)
	}
}
