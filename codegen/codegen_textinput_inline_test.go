package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// textInputWithCursorInfo builds a component with cursor-x and cursor-y attributes.
func textInputWithCursorInfo() *ComponentInfo {
	return &ComponentInfo{
		Name:         "textinput",
		ExportedName: "Textinput",
		Props:        []string{"value"},
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.BoxElement{
					Attributes: map[string]string{
						"focusable": "true",
						"onkey":     "handleEvent",
						"cursor-x":  "{cursor}",
						"cursor-y":  "0",
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
			},
			StateDecls: []script.StateDecl{
				{Name: "cursor", InitExpr: "0"},
			},
			FuncDecls: []script.FuncDecl{
				{
					Name:   "handleEvent",
					Params: "evt input.Event",
					Body: `
	if evt.Kind == input.EventKey {
		value = value[:cursor] + string(evt.Rune) + value[cursor:]
		cursor = cursor + 1
	}
`,
					StateAssignments: []script.StateAssignment{
						{VarName: "value", Line: "value = value[:cursor] + string(evt.Rune) + value[cursor:]"},
						{VarName: "cursor", Line: "cursor = cursor + 1"},
					},
				},
			},
		},
	}
}

func TestInlinedCursorNamespacing(t *testing.T) {
	// Given — parent uses TextInput component with cursor-x={cursor}
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
		Components:  map[string]*ComponentInfo{"textinput": textInputWithCursorInfo()},
	})

	// Then — valid Go with namespaced cursor in extracted box
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// Cursor should be namespaced in the extracted box declaration
	if !strings.Contains(src, "CursorCol: textinput0_cursor") {
		t.Errorf("expected namespaced CursorCol in extracted box:\n%s", src)
	}

	// CursorRow should be 0 (static value from component template)
	if !strings.Contains(src, "CursorRow: 0") {
		t.Errorf("expected CursorRow: 0 in extracted box:\n%s", src)
	}

	// Sync should patch CursorCol with namespaced variable
	if !strings.Contains(src, ".CursorCol = textinput0_cursor") {
		t.Errorf("expected namespaced CursorCol sync:\n%s", src)
	}

	// FindCursor should be in doRender
	if !strings.Contains(src, "layout.FindCursor") {
		t.Errorf("expected FindCursor in doRender:\n%s", src)
	}
}

func TestInlinedCursorBoxExtraction(t *testing.T) {
	// Given — parent uses TextInput component that has a cursor box with expression text
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
		Components:  map[string]*ComponentInfo{"textinput": textInputWithCursorInfo()},
	})

	// Then — text node inside cursor box should be extracted separately
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// The text node should be extracted (references parent's "name" via bind)
	if !strings.Contains(src, "textinput0_node0") {
		t.Errorf("expected extracted text node textinput0_node0:\n%s", src)
	}

	// The cursor box should be extracted
	if !strings.Contains(src, "textinput0_box0") {
		t.Errorf("expected extracted cursor box textinput0_box0:\n%s", src)
	}

	// Text node sync should use bound variable "name"
	if !strings.Contains(src, `textinput0_node0.Content = fmt.Sprintf("%v", name)`) {
		t.Errorf("expected text sync using bound 'name':\n%s", src)
	}
}

func TestInlinedFocusableAttributePreserved(t *testing.T) {
	// Given — TextInput component with focusable="true"
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
		Components:  map[string]*ComponentInfo{"textinput": textInputWithCursorInfo()},
	})

	// Then — focusable should be preserved in extracted box
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	if !strings.Contains(src, "Focusable: true") {
		t.Errorf("expected Focusable: true in extracted cursor box:\n%s", src)
	}
}
