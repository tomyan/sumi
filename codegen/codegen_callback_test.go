package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestCallbackPropResolution(t *testing.T) {
	// Given — a child component with a callback prop that it calls
	childInfo := &ComponentInfo{
		Name:         "button",
		ExportedName: "Button",
		Props:        []string{"onPress"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{"onkey": "handleKey"},
					Children: []template.Node{
						&template.TextElement{
							Parts: []template.Part{
								&template.StringPart{Value: "Click me"},
							},
						},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "onPress", DefaultExpr: `""`},
			},
			StateDecls: []script.StateDecl{
				{Name: "pressed", InitExpr: "false"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "handleKey", Params: "evt input.Event",
					Body: "\n\tonPress()\n\tpressed = true\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "pressed", Line: "pressed = true"},
					},
				},
			},
		},
	}

	// Parent passes its function as callback prop
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "button",
				Attributes: map[string]string{"onPress": "handleSubmit"},
			},
		},
	}
	parentSc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "submitted", InitExpr: "false"},
		},
		FuncDecls: []script.FuncDecl{
			{
				Name: "handleSubmit", Params: "", Body: "\n\tsubmitted = true\n",
				StateAssignments: []script.StateAssignment{
					{VarName: "submitted", Line: "submitted = true"},
				},
			},
		},
	}

	// When
	out, err := Generate(doc, parentSc, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"button": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Child's onPress() call should become handleSubmit()
	if !strings.Contains(src, "handleSubmit()") {
		t.Errorf("expected callback prop resolved to handleSubmit():\n%s", src)
	}
	// Should NOT contain the raw prop call
	if strings.Contains(src, "onPress()") {
		t.Errorf("should not contain unresolved onPress() call:\n%s", src)
	}
}

func TestBindValueStateMapping(t *testing.T) {
	// Given — a child component with bind:value
	childInfo := &ComponentInfo{
		Name:         "field",
		ExportedName: "Field",
		Props:        []string{"value"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{"onkey": "handleKey"},
					Children: []template.Node{
						&template.TextElement{
							Parts: []template.Part{
								&template.ExprPart{Expr: "value"},
							},
						},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "value", DefaultExpr: `""`},
			},
			StateDecls: []script.StateDecl{
				{Name: "value", InitExpr: `""`},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "handleKey", Params: "evt input.Event",
					Body: "\n\tvalue = value + string(evt.Rune)\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "value", Line: `value = value + string(evt.Rune)`},
					},
				},
			},
		},
	}

	// Parent uses bind:value to map child's value to parent's name
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
				Children: []template.Node{
					&template.ComponentElement{
						Name:       "field",
						Attributes: map[string]string{"bind:value": "name"},
					},
					&template.TextElement{
						Parts: []template.Part{
							&template.StringPart{Value: "Name: "},
							&template.ExprPart{Expr: "name"},
						},
					},
				},
			},
		},
	}
	parentSc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "name", InitExpr: `""`},
		},
	}

	// When
	out, err := Generate(doc, parentSc, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"field": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Child's value should map directly to parent's name (no namespace prefix)
	if strings.Contains(src, "field0_value") {
		t.Errorf("bound variable should not be namespaced:\n%s", src)
	}

	// Parent's name variable should exist
	if !strings.Contains(src, `name := ""`) {
		t.Errorf("expected parent state name declaration:\n%s", src)
	}

	// The handler body should reference the parent's "name" variable directly
	// (via bind resolution, the child's "value" → parent's "name")
	if !strings.Contains(src, "name = name + string(evt.Rune)") {
		t.Errorf("expected bound variable 'name' in handler body:\n%s", src)
	}
}
