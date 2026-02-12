package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateCursorColStatic(t *testing.T) {
	// Given — a box with a static cursor-x attribute
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"cursor-x": "5"},
				Children:   []template.Node{textNode("hello")},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "CursorCol: 5") {
		t.Errorf("expected CursorCol: 5 in output:\n%s", src)
	}
}

func TestGenerateCursorRowStatic(t *testing.T) {
	// Given — a box with cursor-y
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"cursor-y": "2"},
				Children:   []template.Node{textNode("hello")},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "CursorRow: 2") {
		t.Errorf("expected CursorRow: 2 in output:\n%s", src)
	}
}

func TestGenerateDefaultCursorNegativeOne(t *testing.T) {
	// Given — a box without cursor attributes
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Children: []template.Node{textNode("hello")},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	// Default is -1 (no cursor) — emitted on all boxes
	if !strings.Contains(src, "CursorCol: -1") {
		t.Errorf("expected CursorCol: -1 in output:\n%s", src)
	}
	if !strings.Contains(src, "CursorRow: -1") {
		t.Errorf("expected CursorRow: -1 in output:\n%s", src)
	}
}

func TestGenerateCursorDynamicExpression(t *testing.T) {
	// Given — cursor-x is a dynamic expression referencing state
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"cursor-x": "{cursor}"},
				Children:   []template.Node{textNode("hello")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "cursor", InitExpr: "0"},
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
	// The cursor expression should reference the state variable
	if !strings.Contains(src, "CursorCol: cursor") {
		t.Errorf("expected dynamic CursorCol: cursor in output:\n%s", src)
	}
}

func TestGenerateCursorPositioningInDoRender(t *testing.T) {
	// Given — a box with dynamic cursor-x
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"cursor-x": "{cursor}"},
				Children:   []template.Node{textNode("hello")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "cursor", InitExpr: "0"},
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
	// doRender should call FindCursor and ShowCursor/HideCursor
	if !strings.Contains(src, "layout.FindCursor") {
		t.Errorf("expected FindCursor call in doRender:\n%s", src)
	}
	if !strings.Contains(src, "render.ShowCursor") {
		t.Errorf("expected ShowCursor call in doRender:\n%s", src)
	}
	if !strings.Contains(src, "render.HideCursor") {
		t.Errorf("expected HideCursor call in doRender:\n%s", src)
	}
}

func TestGenerateNoCursorPositioningWithoutCursorAttr(t *testing.T) {
	// Given — no cursor attributes
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Children: []template.Node{textNode("hello")},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	// No cursor system should be emitted
	if strings.Contains(src, "FindCursor") {
		t.Errorf("should not contain FindCursor when no cursor attributes:\n%s", src)
	}
}

func TestGenerateCursorSyncPatching(t *testing.T) {
	// Given — cursor-x is dynamic, needs sync patching for build-once tree
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"cursor-x": "{cursor}"},
				Children:   []template.Node{textNode("hello")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "cursor", InitExpr: "0"},
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
	// Sync function should patch CursorCol
	if !strings.Contains(src, ".CursorCol = cursor") {
		t.Errorf("expected CursorCol sync patching in output:\n%s", src)
	}
}
