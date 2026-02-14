package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

func TestExpressionHeightAttribute(t *testing.T) {
	// Given — a box with an expression-valued height from parent state
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"height": "{panelHeight}",
				},
				Children: []template.Node{textNode("content")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "panelHeight", InitExpr: "10"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "FixedHeight: panelHeight,") {
		t.Errorf("expected FixedHeight: panelHeight, in output:\n%s", src)
	}
}

func TestExpressionBorderTitle(t *testing.T) {
	// Given — a box with an expression-valued border-title from parent state
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"border":       "single",
					"border-title": "{myTitle}",
				},
				Children: []template.Node{textNode("content")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "myTitle", InitExpr: `"Hello"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "BorderTitle: myTitle,") {
		t.Errorf("expected BorderTitle: myTitle, in output:\n%s", src)
	}
}

func TestPropOnlyComponentInlining(t *testing.T) {
	// Given — a component with only $prop declarations (no $state)
	childInfo := &ComponentInfo{
		Name:         "panel",
		ExportedName: "Panel",
		Props:        []string{"title", "height"},
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{
						"border":       "single",
						"border-title": "{title}",
						"height":       "{height}",
					},
					Children: []template.Node{textNode("inside")},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "title", DefaultExpr: `""`},
				{Name: "height", DefaultExpr: "0"},
			},
		},
	}

	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "panel",
				Attributes: map[string]string{"title": "My Panel", "height": "{panelH}"},
			},
		},
	}
	parentSc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "panelH", InitExpr: "10"},
		},
	}

	// When
	out, err := Generate(doc, parentSc, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"panel": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Literal prop should create a namespaced variable
	if !strings.Contains(src, `panel0_title := "My Panel"`) {
		t.Errorf("expected panel0_title declaration:\n%s", src)
	}
	// Expression prop should resolve to parent variable
	if !strings.Contains(src, "BorderTitle: panel0_title,") {
		t.Errorf("expected BorderTitle: panel0_title, in output:\n%s", src)
	}
	// Expression height prop should resolve to parent expression
	if !strings.Contains(src, "FixedHeight: panelH,") {
		t.Errorf("expected FixedHeight: panelH, in output:\n%s", src)
	}
}

func TestPropOnlyComponentWithStylesheet(t *testing.T) {
	// Given — a prop-only component with style classes
	childInfo := &ComponentInfo{
		Name:         "card",
		ExportedName: "Card",
		Props:        []string{"label"},
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{"class": "card"},
					Children: []template.Node{
						&template.TextElement{
							Parts: []template.Part{
								&template.ExprPart{Expr: "label"},
							},
						},
					},
				},
			},
		},
		Stylesheet: &style.Stylesheet{
			Rules: []style.Rule{
				{Selector: ".card", Properties: map[string]string{"border": "single"}},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "label", DefaultExpr: `""`},
			},
		},
	}

	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "card",
				Attributes: map[string]string{"label": "Hello"},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"card": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, `"Hello"`) {
		t.Errorf("expected literal prop value in output:\n%s", src)
	}
}
