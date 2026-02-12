package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestFocusIndexStartsUnfocused(t *testing.T) {
	// Given — a focusable box with handler
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "handler"},
				Children:   []template.Node{textNode("Field")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "handler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}}},
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
	if !strings.Contains(src, "focusIndex := -1") {
		t.Errorf("expected focusIndex := -1 (start unfocused):\n%s", src)
	}
}

func TestCursorSyncConditionalOnFocus(t *testing.T) {
	// Given — a focusable box with dynamic cursor
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"focusable": "true",
					"onkey":     "handler",
					"cursor-x":  "{pos}",
					"cursor-y":  "0",
				},
				Children: []template.Node{textNode("Field")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
			{Name: "pos", InitExpr: "0"},
		},
		FuncDecls: []script.FuncDecl{
			{Name: "handler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}}},
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
	// Cursor sync should be conditional on focus
	if !strings.Contains(src, "if focusIndex == 0") {
		t.Errorf("expected cursor conditional on focusIndex:\n%s", src)
	}
	// Should hide cursor when unfocused
	if !strings.Contains(src, "CursorCol = -1") {
		t.Errorf("expected CursorCol = -1 in else branch:\n%s", src)
	}
}

func TestFocusedStateSyncedForInlinedComponent(t *testing.T) {
	// Given — an inlined component with focused state and focusable box
	childDoc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"focusable": "true",
					"onkey":     "handler",
					"cursor-x":  "{pos}",
					"cursor-y":  "0",
				},
				Children: []template.Node{textNode("Field")},
			},
		},
	}
	childScript := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "pos", InitExpr: "0"},
			{Name: "focused", InitExpr: "false"},
		},
		FuncDecls: []script.FuncDecl{
			{Name: "handler", Params: "evt input.Event", Body: "\n\tpos = pos + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "pos", Line: "pos = pos + 1"}}},
		},
	}
	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{Name: "widget"},
		},
	}
	sc := &script.Script{}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName: "main",
		Components: map[string]*ComponentInfo{
			"widget": {
				Name: "widget", ExportedName: "Widget",
				HasState: true,
				Doc: childDoc, Script: childScript,
				Props: []string{},
			},
		},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)

	// sync should write focused state based on focusIndex
	if !strings.Contains(src, "widget0_focused = focusIndex == 0") {
		t.Errorf("expected focused sync in output:\n%s", src)
	}
}

func TestCursorUnconditionalWhenNotFocusable(t *testing.T) {
	// Given — a box with dynamic cursor but NOT focusable
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"cursor-x": "{pos}",
					"cursor-y": "0",
				},
				Children: []template.Node{textNode("Field")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "pos", InitExpr: "0"},
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
	// Should NOT have focus-conditional cursor
	if strings.Contains(src, "if focusIndex") {
		t.Errorf("cursor should be unconditional when box is not focusable:\n%s", src)
	}
	// Should still have cursor sync
	if !strings.Contains(src, "CursorCol = pos") {
		t.Errorf("expected unconditional CursorCol = pos:\n%s", src)
	}
}
