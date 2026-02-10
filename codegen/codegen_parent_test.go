package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func counterComponents() map[string]*ComponentInfo {
	return map[string]*ComponentInfo{
		"counter": {
			Name:         "counter",
			ExportedName: "Counter",
			Props:        []string{"label"},
			HasState:     true,
		},
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

	// When generating code with component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  counterComponents(),
	})

	// Then the output is valid Go containing NewCounterComponent
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "NewCounterComponent") {
		t.Errorf("expected NewCounterComponent call in output:\n%s", src)
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

	// When generating code with component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  counterComponents(),
	})

	// Then both counter0 and counter1 appear with unique variable names
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0") {
		t.Errorf("expected counter0 variable in output:\n%s", src)
	}
	if !strings.Contains(src, "counter1") {
		t.Errorf("expected counter1 variable in output:\n%s", src)
	}
}

func TestGenerateParentCallsChildLayout(t *testing.T) {
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

	// When generating code with component info
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  counterComponents(),
	})

	// Then the layout tree references counter0.Layout()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0.Layout()") {
		t.Errorf("expected counter0.Layout() in layout tree:\n%s", src)
	}
}

func TestGenerateParentDispatchesHandleKey(t *testing.T) {
	// Given a parent document with a stateful <counter /> child
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

	// When generating code with HasState=true
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  counterComponents(),
	})

	// Then the event loop dispatches HandleKey to the child
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0.HandleKey(evt.Rune)") {
		t.Errorf("expected counter0.HandleKey(evt.Rune) in event loop:\n%s", src)
	}
}

func TestGenerateParentChecksDirty(t *testing.T) {
	// Given a parent document with a stateful <counter /> child
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

	// When generating code with HasState=true
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  counterComponents(),
	})

	// Then the dirty check includes counter0.Dirty()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "counter0.Dirty()") {
		t.Errorf("expected counter0.Dirty() in dirty check:\n%s", src)
	}
}

func TestGenerateParentPassesProps(t *testing.T) {
	// Given a parent document with <counter label="Clicks" />
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "counter",
				Attributes: map[string]string{"label": "Clicks"},
			},
		},
	}

	// When generating code with component info specifying props
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  counterComponents(),
	})

	// Then the constructor call includes the prop value
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `NewCounterComponent("Clicks")`) {
		t.Errorf("expected NewCounterComponent(\"Clicks\") in output:\n%s", src)
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
		Components:  counterComponents(),
	})

	// Then reactive code (event loop) is generated, not static
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "input.ReadEvent") {
		t.Errorf("expected input.ReadEvent (reactive mode) in output:\n%s", src)
	}
	if strings.Contains(src, "bufio.NewScanner") {
		t.Errorf("expected no bufio.NewScanner (static mode) in output:\n%s", src)
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
		Components:  counterComponents(),
	})

	// Then both TextElement and ComponentElement are in the output
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
	if !strings.Contains(src, "counter0.Layout()") {
		t.Errorf("expected counter0.Layout() in output:\n%s", src)
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
		Components:  counterComponents(),
	})

	// Then both parent state and child component code is present
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
	if !strings.Contains(src, "NewCounterComponent") {
		t.Errorf("expected NewCounterComponent call:\n%s", src)
	}
	if !strings.Contains(src, "counter0.HandleKey(evt.Rune)") {
		t.Errorf("expected counter0.HandleKey(evt.Rune):\n%s", src)
	}
	if !strings.Contains(src, "counter0.Dirty()") {
		t.Errorf("expected counter0.Dirty() in dirty check:\n%s", src)
	}
}
