package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// counterWithDerivedInfo builds a ComponentInfo for a counter with a derived value.
func counterWithDerivedInfo() *ComponentInfo {
	return &ComponentInfo{
		Name:         "counter",
		ExportedName: "Counter",
		Props:        []string{"label"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{"onkey": "increment"},
					Children: []template.Node{
						&template.TextElement{
							Parts: []template.Part{
								&template.ExprPart{Expr: "label"},
								&template.StringPart{Value: ": "},
								&template.ExprPart{Expr: "doubled"},
							},
						},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "label", DefaultExpr: `"Count"`},
			},
			StateDecls: []script.StateDecl{
				{Name: "count", InitExpr: "0"},
			},
			DerivedDecls: []script.DerivedDecl{
				{Name: "doubled", Expr: "count * 2"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "increment", Params: "", Body: "\n\tcount = count + 1\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "count", Line: "count = count + 1"},
					},
				},
			},
		},
	}
}

func TestDerivedDeclarationEmitted(t *testing.T) {
	// Given — state + derived
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Doubled: "},
					&template.ExprPart{Expr: "doubled"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls:   []script.StateDecl{{Name: "count", InitExpr: "0"}},
		DerivedDecls: []script.DerivedDecl{{Name: "doubled", Expr: "count * 2"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Initial derived declaration
	if !strings.Contains(src, "doubled := count * 2") {
		t.Errorf("expected derived declaration 'doubled := count * 2':\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedRecalculatedInSync(t *testing.T) {
	// Given — state + derived
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Doubled: "},
					&template.ExprPart{Expr: "doubled"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls:   []script.StateDecl{{Name: "count", InitExpr: "0"}},
		DerivedDecls: []script.DerivedDecl{{Name: "doubled", Expr: "count * 2"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Sync should reassign (not :=) the derived value
	if !strings.Contains(src, "doubled = count * 2") {
		t.Errorf("expected derived recalculation 'doubled = count * 2' in sync:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedInTemplateExpression(t *testing.T) {
	// Given — derived used in template
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Value: "},
					&template.ExprPart{Expr: "doubled"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls:   []script.StateDecl{{Name: "count", InitExpr: "0"}},
		DerivedDecls: []script.DerivedDecl{{Name: "doubled", Expr: "count * 2"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Extracted node should reference doubled
	if !strings.Contains(src, "doubled") {
		t.Errorf("expected 'doubled' in template expression:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedIsReactive(t *testing.T) {
	// Given — only derived decls, no state
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Hello"},
				},
			},
		},
	}
	sc := &script.Script{
		DerivedDecls: []script.DerivedDecl{{Name: "label", Expr: `"hello"`}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then — derived alone should trigger reactive path
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should have tui.App (reactive path)
	if !strings.Contains(src, "tui.App") {
		t.Errorf("expected reactive path with tui.App for derived-only script:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedValidGo(t *testing.T) {
	// Given — full round-trip: state + derived + func + template
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "increment"},
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Count: "},
							&template.ExprPart{Expr: "count"},
						},
					},
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Doubled: "},
							&template.ExprPart{Expr: "doubled"},
						},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls:   []script.StateDecl{{Name: "count", InitExpr: "0"}},
		DerivedDecls: []script.DerivedDecl{{Name: "doubled", Expr: "count * 2"}},
		FuncDecls: []script.FuncDecl{
			{
				Name: "increment", Params: "", Body: "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{
					{VarName: "count", Line: "count = count + 1"},
				},
			},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertValidGo(t, out)
}

func TestDerivedInInlinedComponent(t *testing.T) {
	// Given — parent with a child component that has a derived value
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
				Children: []template.Node{
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Clicks"},
					},
				},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterWithDerivedInfo()},
	})

	// Then — derived should be namespaced
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "counter0_doubled := counter0_count * 2") {
		t.Errorf("expected namespaced derived declaration:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedNamespacedInSync(t *testing.T) {
	// Given — parent with a child component that has a derived value
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterWithDerivedInfo()},
	})

	// Then — sync should have namespaced derived recalculation
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "counter0_doubled = counter0_count * 2") {
		t.Errorf("expected namespaced derived recalculation in sync:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedNamespacedInTemplate(t *testing.T) {
	// Given — parent with a child component that uses derived in template
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterWithDerivedInfo()},
	})

	// Then — extracted node should reference namespaced derived var
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "counter0_doubled") {
		t.Errorf("expected counter0_doubled in template expression:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestDerivedMultipleInstances(t *testing.T) {
	// Given — parent with two counter components, both having derived values
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
				Children: []template.Node{
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Clicks"},
					},
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Score"},
					},
				},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"counter": counterWithDerivedInfo()},
	})

	// Then — each instance should have uniquely namespaced derived values
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "counter0_doubled") {
		t.Errorf("expected counter0_doubled:\n%s", src)
	}
	if !strings.Contains(src, "counter1_doubled") {
		t.Errorf("expected counter1_doubled:\n%s", src)
	}

	assertValidGo(t, out)
}
