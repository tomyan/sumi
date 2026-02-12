package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// textInputComponentInfo builds a ComponentInfo for a TextInput-like component.
func textInputComponentInfo() *ComponentInfo {
	return &ComponentInfo{
		Name:         "textinput",
		ExportedName: "Textinput",
		Props:        []string{"value", "onSubmit"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{
						"focusable": "true",
						"onkey":     "handleEvent",
					},
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
				{Name: "onSubmit", DefaultExpr: `""`},
			},
			StateDecls: []script.StateDecl{
				{Name: "cursor", InitExpr: "0"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name:   "handleEvent",
					Params: "evt input.Event",
					Body: `
	if evt.Special == input.KeyBackspace && cursor > 0 {
		value = value[:cursor-1] + value[cursor:]
		cursor = cursor - 1
	}
	if evt.Kind == input.EventKey {
		value = value[:cursor] + string(evt.Rune) + value[cursor:]
		cursor = cursor + 1
	}
`,
					StateAssignments: []script.StateAssignment{
						{VarName: "value", Line: "value = value[:cursor-1] + value[cursor:]"},
						{VarName: "cursor", Line: "cursor = cursor - 1"},
						{VarName: "value", Line: "value = value[:cursor] + string(evt.Rune) + value[cursor:]"},
						{VarName: "cursor", Line: "cursor = cursor + 1"},
					},
				},
			},
		},
	}
}

func TestTextInputComplexHandlerNamespacing(t *testing.T) {
	// Given — a parent using a TextInput-like component with bind:value
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "textinput",
				Attributes: map[string]string{"bind:value": "name", "onSubmit": "{handleSubmit}"},
			},
		},
	}
	parentSc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "name", InitExpr: `""`},
		},
		FuncDecls: []script.FuncDecl{
			{
				Name: "handleSubmit", Params: "", Body: "\n",
			},
		},
	}

	// When
	out, err := Generate(doc, parentSc, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"textinput": textInputComponentInfo()},
	})

	// Then — generated code should be valid Go
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// cursor should be namespaced everywhere (conditionals AND assignments)
	if !strings.Contains(src, "textinput0_cursor > 0") {
		t.Errorf("expected namespaced cursor in conditional:\n%s", src)
	}

	// Array indexing should use namespaced cursor
	if !strings.Contains(src, "name[:textinput0_cursor-1]") {
		t.Errorf("expected namespaced cursor in array index:\n%s", src)
	}

	// value should map to parent's "name" via bind
	if !strings.Contains(src, "name = name[:textinput0_cursor-1]") {
		t.Errorf("expected bound value in assignment:\n%s", src)
	}

	// cursor assignment should be namespaced
	if !strings.Contains(src, "textinput0_cursor = textinput0_cursor - 1") {
		t.Errorf("expected namespaced cursor assignment:\n%s", src)
	}

	// No un-namespaced cursor references (except in evt.Special patterns)
	// This checks that "cursor > 0" was properly renamed
	if strings.Contains(src, " cursor > 0") {
		t.Errorf("should not have un-namespaced cursor reference:\n%s", src)
	}
}

func TestTextInputBoundPropOnlyRenaming(t *testing.T) {
	// Given — a child where the bound variable is ONLY a $prop (not also $state)
	childInfo := &ComponentInfo{
		Name:         "textinput",
		ExportedName: "Textinput",
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
				{Name: "cursor", InitExpr: "0"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name: "handleKey", Params: "evt input.Event",
					Body: "\n\tvalue = value + string(evt.Rune)\n\tcursor = cursor + 1\n",
					StateAssignments: []script.StateAssignment{
						{VarName: "value", Line: "value = value + string(evt.Rune)"},
						{VarName: "cursor", Line: "cursor = cursor + 1"},
					},
				},
			},
		},
	}

	// Parent uses bind:value to map child's value prop to parent's name state
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "textinput",
				Attributes: map[string]string{"bind:value": "name"},
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
		Components:  map[string]*ComponentInfo{"textinput": childInfo},
	})

	// Then — value (a prop-only var) should be renamed to parent's "name"
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Handler body should use "name" instead of "value"
	if !strings.Contains(src, "name = name + string(evt.Rune)") {
		t.Errorf("expected bound prop 'value' renamed to 'name' in handler:\n%s", src)
	}

	// cursor should be namespaced normally
	if !strings.Contains(src, "textinput0_cursor = textinput0_cursor + 1") {
		t.Errorf("expected namespaced cursor in handler:\n%s", src)
	}
}

func TestTextInputCursorNotDeclaredForBoundVar(t *testing.T) {
	// Given — a parent using TextInput with bind:value
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "textinput",
				Attributes: map[string]string{"bind:value": "name"},
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
		Components:  map[string]*ComponentInfo{"textinput": textInputComponentInfo()},
	})

	// Then — bound var should NOT be re-declared; cursor SHOULD be declared
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// cursor is own state, should be declared
	if !strings.Contains(src, "textinput0_cursor := 0") {
		t.Errorf("expected textinput0_cursor declaration:\n%s", src)
	}

	// value is bound to parent's name, should NOT be declared separately
	if strings.Contains(src, "textinput0_value") {
		t.Errorf("bound variable should not be namespaced as textinput0_value:\n%s", src)
	}
}
