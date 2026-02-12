package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// inlinableCounter builds a ComponentInfo with full AST for inlining.
func inlinableCounter() map[string]*ComponentInfo {
	return map[string]*ComponentInfo{
		"counter": counterComponentInfo(),
	}
}

func TestGenerateParentWithSingleChild(t *testing.T) {
	// Given a parent document with one <counter label="Clicks" />
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

	// When generating code with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then the output is valid Go with inlined child state
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "counter0_count") {
		t.Errorf("expected inlined counter0_count in output:\n%s", src)
	}
}

func TestGenerateParentWithMultipleChildren(t *testing.T) {
	// Given a parent document with two <counter /> elements
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

	// When generating code with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then both instances appear with unique namespaces
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0_count") {
		t.Errorf("expected counter0_count in output:\n%s", src)
	}
	if !strings.Contains(src, "counter1_count") {
		t.Errorf("expected counter1_count in output:\n%s", src)
	}
}

func TestGenerateParentInlinesChildTemplate(t *testing.T) {
	// Given a parent document with a <counter /> element
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

	// When generating code with inlinable component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then the child template is inlined (no Layout() call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, ".Layout()") {
		t.Errorf("should not have .Layout() call, child is inlined:\n%s", src)
	}
	// Prop should be resolved to literal within the Sprintf format string
	if !strings.Contains(src, `"Clicks: `) {
		t.Errorf("expected prop resolved to literal in inlined template:\n%s", src)
	}
}

func TestGenerateParentInlinesOnkeyHandler(t *testing.T) {
	// Given a parent document with a stateful counter that has onkey
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating code
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then the event loop calls the inlined handler directly
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0_increment()") {
		t.Errorf("expected inlined onkey call counter0_increment():\n%s", src)
	}
	if strings.Contains(src, ".HandleKey(") {
		t.Errorf("should not have .HandleKey() dispatch:\n%s", src)
	}
}

func TestGenerateParentSingleDirtyFlag(t *testing.T) {
	// Given a parent with a stateful child component
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating code
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then a single dirty flag via app.Dirty (no Dirty() polling)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, ".Dirty()") {
		t.Errorf("should not have .Dirty() polling:\n%s", src)
	}
	if !strings.Contains(src, "app.Dirty = true") {
		t.Errorf("expected app.Dirty = true in output:\n%s", src)
	}
}

func TestGenerateParentWithoutScriptUsesReactive(t *testing.T) {
	// Given a parent document with no script but with component refs
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating code with nil script but with components
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then reactive code (tui.App with OnEvent) is generated, not static
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "OnEvent:") {
		t.Errorf("expected OnEvent in reactive mode output:\n%s", src)
	}
}

func TestGenerateParentMixedComponentsAndText(t *testing.T) {
	// Given a parent document with both text and component children
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
				Children: []template.Node{
					textNode("Title"),
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Clicks"},
					},
				},
			},
		},
	}

	// When generating code with component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then both TextElement and inlined component are in the output
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, `"Title"`) {
		t.Errorf("expected Title text in output:\n%s", src)
	}
	if !strings.Contains(src, "counter0_count") {
		t.Errorf("expected inlined counter0_count in output:\n%s", src)
	}
}

func TestGenerateParentWithStatefulScriptAndChild(t *testing.T) {
	// Given a parent with its own state AND a child component
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column", "onkey": "toggle"},
				Children: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Active: "},
							&template.ExprPart{Expr: "active"},
						},
					},
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{"label": "Clicks"},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "active", InitExpr: "true"},
		},
		FuncDecls: []script.FuncDecl{
			{
				Name: "toggle", Params: "", Body: "\n\tactive = !active\n",
				StateAssignments: []script.StateAssignment{
					{VarName: "active", Line: "active = !active"},
				},
			},
		},
	}

	// When generating code
	out, err := Generate(doc, sc, nil, Options{
		PackageName: "main",
		Components:  inlinableCounter(),
	})

	// Then both parent state and inlined child are present
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "active := true") {
		t.Errorf("expected parent state declaration:\n%s", src)
	}
	if !strings.Contains(src, "counter0_count") {
		t.Errorf("expected inlined counter0_count:\n%s", src)
	}
	if !strings.Contains(src, "counter0_increment()") {
		t.Errorf("expected inlined onkey handler call:\n%s", src)
	}
}
