package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateFocusableAttribute(t *testing.T) {
	// Given — a box with focusable="true"
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true"},
				Children:   []template.Node{textNode("Field 1")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
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
	if !strings.Contains(src, "Focusable: true") {
		t.Errorf("expected Focusable: true in output:\n%s", src)
	}
}

func TestGenerateFocusStateVars(t *testing.T) {
	// Given — two focusable boxes with handlers
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "handler1"},
				Children:   []template.Node{textNode("Field 1")},
			},
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "handler2"},
				Children:   []template.Node{textNode("Field 2")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
		FuncDecls: []script.FuncDecl{
			{Name: "handler1", Params: "evt input.Event", Body: "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "count", Line: "count = count + 1"}}},
			{Name: "handler2", Params: "evt input.Event", Body: "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "count", Line: "count = count + 1"}}},
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
	// Focus state variables
	if !strings.Contains(src, "focusIndex := -1") {
		t.Errorf("expected focusIndex := -1:\n%s", src)
	}
	if !strings.Contains(src, "focusCount := 2") {
		t.Errorf("expected focusCount := 2:\n%s", src)
	}
}

func TestGenerateFocusTabCycling(t *testing.T) {
	// Given — two focusable boxes
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "h1"},
				Children:   []template.Node{textNode("A")},
			},
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "h2"},
				Children:   []template.Node{textNode("B")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "h1", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}}},
			{Name: "h2", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
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
	// Tab cycling
	if !strings.Contains(src, "input.KeyTab") {
		t.Errorf("expected Tab key handling:\n%s", src)
	}
	if !strings.Contains(src, "input.KeyShiftTab") {
		t.Errorf("expected Shift-Tab key handling:\n%s", src)
	}
}

func TestGenerateFocusDirectedDispatch(t *testing.T) {
	// Given — two focusable boxes with handlers + a root handler
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "globalHandler"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"focusable": "true", "onkey": "inputHandler"},
						Children:   []template.Node{textNode("Field 1")},
					},
					&template.BoxElement{
						Attributes: map[string]string{"focusable": "true", "onkey": "buttonHandler"},
						Children:   []template.Node{textNode("Submit")},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "globalHandler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}}},
			{Name: "inputHandler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}}},
			{Name: "buttonHandler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
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
	// Focus-directed dispatch switch
	if !strings.Contains(src, "switch focusIndex") {
		t.Errorf("expected switch focusIndex dispatch:\n%s", src)
	}
	// Each handler called based on focus
	if !strings.Contains(src, "inputHandler(evt)") {
		t.Errorf("expected inputHandler(evt) in switch:\n%s", src)
	}
	if !strings.Contains(src, "buttonHandler(evt)") {
		t.Errorf("expected buttonHandler(evt) in switch:\n%s", src)
	}
}

func TestGenerateStopPropagation(t *testing.T) {
	// Given — focusable boxes with handlers
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "globalHandler"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"focusable": "true", "onkey": "inputHandler"},
						Children:   []template.Node{textNode("Field 1")},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "globalHandler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}}},
			{Name: "inputHandler", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
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
	// stopPropagation function
	if !strings.Contains(src, "stopPropagation := func()") {
		t.Errorf("expected stopPropagation function:\n%s", src)
	}
	// Bubbling check
	if !strings.Contains(src, "propagationStopped") {
		t.Errorf("expected propagationStopped variable:\n%s", src)
	}
	// Global handler behind propagation check
	if !strings.Contains(src, "if !propagationStopped") {
		t.Errorf("expected propagation guard before global handler:\n%s", src)
	}
}

func TestGenerateHasMouseWhenFocusableBoxes(t *testing.T) {
	// Given — a focusable box
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "handleKey"},
				Children:   []template.Node{textNode("Input")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "handleKey", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
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
	if !strings.Contains(src, "HasMouse: true") {
		t.Errorf("expected HasMouse: true when focusable boxes present:\n%s", src)
	}
}

func TestGenerateTabCyclesToUnfocused(t *testing.T) {
	// Given — a single focusable box; Tab should cycle: -1 → 0 → -1 → 0
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"focusable": "true", "onkey": "h1"},
				Children:   []template.Node{textNode("A")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "h1", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
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

	// Tab should wrap around through -1 (unfocused), not stay stuck at last index
	// focusIndex goes: -1 → 0 → -1 → 0 ...
	// This requires cycling over focusCount+1 positions and subtracting 1
	if !strings.Contains(src, "focusIndex = (focusIndex+2)%(focusCount+1) - 1") {
		t.Errorf("expected Tab to cycle through unfocused state:\n%s", src)
	}
}

func TestGenerateNoFocusWhenNoFocusableBoxes(t *testing.T) {
	// Given — no focusable boxes
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "x", InitExpr: "0"}},
		FuncDecls: []script.FuncDecl{
			{Name: "handleKey", Params: "evt input.Event", Body: "\n\tx = x + 1\n",
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
	// No focus system
	if strings.Contains(src, "focusIndex") {
		t.Errorf("should not have focusIndex when no focusable boxes:\n%s", src)
	}
	if strings.Contains(src, "stopPropagation") {
		t.Errorf("should not have stopPropagation when no focusable boxes:\n%s", src)
	}
}
