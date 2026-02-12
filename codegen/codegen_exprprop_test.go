package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestExpressionPropInText(t *testing.T) {
	// Given — a child component receives an expression prop used in text
	childInfo := &ComponentInfo{
		Name:         "label",
		ExportedName: "Label",
		Props:        []string{"value"},
		Doc: &template.Document{
			Children: []template.Node{
				&template.TextElement{
					Parts: []template.Part{
						&template.StringPart{Value: "Value: "},
						&template.ExprPart{Expr: "value"},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "value", DefaultExpr: `""`},
			},
		},
	}

	// Parent passes expression prop: value={count}
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "label",
				Attributes: map[string]string{"value": "{count}"},
			},
		},
	}
	parentSc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, parentSc, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"label": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Expression prop should resolve to fmt.Sprintf("Value: %v", count) or similar
	// The key indicator is that count appears as a Go expression argument, not inside a string literal
	if !strings.Contains(src, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for expression prop in output:\n%s", src)
	}
	// count should appear as a Sprintf argument, not inside a string
	if !strings.Contains(src, ", count)") {
		t.Errorf("expected count as Sprintf argument:\n%s", src)
	}
}

func TestLiteralPropInText(t *testing.T) {
	// Given — a child component receives a literal prop used in text
	childInfo := &ComponentInfo{
		Name:         "label",
		ExportedName: "Label",
		Props:        []string{"title"},
		Doc: &template.Document{
			Children: []template.Node{
				&template.TextElement{
					Parts: []template.Part{
						&template.ExprPart{Expr: "title"},
					},
				},
			},
		},
		Script: &script.Script{
			PropDecls: []script.PropDecl{
				{Name: "title", DefaultExpr: `""`},
			},
		},
	}

	// Parent passes literal prop: title="Hello"
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "label",
				Attributes: map[string]string{"title": "Hello"},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"label": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Literal prop should become a Go string literal "Hello"
	if !strings.Contains(src, `"Hello"`) {
		t.Errorf("expected literal \"Hello\" in output:\n%s", src)
	}
}

func TestExpressionCallbackProp(t *testing.T) {
	// Given — a child component with a callback prop passed as expression
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

	// Parent passes callback as expression: onPress={handleSubmit}
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "button",
				Attributes: map[string]string{"onPress": "{handleSubmit}"},
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

	// Child's onPress() call should resolve to handleSubmit()
	if !strings.Contains(src, "handleSubmit()") {
		t.Errorf("expected callback resolved to handleSubmit():\n%s", src)
	}
	// Should NOT contain unresolved onPress()
	if strings.Contains(src, "onPress()") {
		t.Errorf("should not contain unresolved onPress() call:\n%s", src)
	}
}
