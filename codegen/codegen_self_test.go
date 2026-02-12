package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestSelfWidthDeclarationEmitted(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		SelfDecls: []script.SelfDecl{{Name: "selfW", Key: "width"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "selfW := 0") {
		t.Errorf("expected self variable declaration 'selfW := 0':\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfWidthPointerWiredOnRoot(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		SelfDecls: []script.SelfDecl{{Name: "selfW", Key: "width"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "root.SelfW = &selfW") {
		t.Errorf("expected SelfW pointer wiring on root:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfHeightPointerWiredOnRoot(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		SelfDecls: []script.SelfDecl{{Name: "selfH", Key: "height"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "root.SelfH = &selfH") {
		t.Errorf("expected SelfH pointer wiring on root:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfTriggersReactivePath(t *testing.T) {
	// Given — only self decl, no state
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		SelfDecls: []script.SelfDecl{{Name: "selfW", Key: "width"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "tui.App") {
		t.Errorf("expected reactive path with tui.App for self-only script:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfInInlinedComponent(t *testing.T) {
	// Given — parent with a child component that has $self(width)
	childInfo := &ComponentInfo{
		Name:         "widget",
		ExportedName: "Widget",
		HasState:     true,
		Doc: &template.Document{
			Children: []template.Node{
				&template.TextElement{
					Parts: []template.Part{
						&template.ExprPart{Expr: `fmt.Sprintf("w=%d", selfW)`},
					},
				},
			},
		},
		Script: &script.Script{
			SelfDecls: []script.SelfDecl{{Name: "selfW", Key: "width"}},
		},
	}

	doc := &template.Document{
		Children: []template.Node{
			&template.ComponentElement{
				Name:       "widget",
				Attributes: map[string]string{},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{
		PackageName: "main",
		Components:  map[string]*ComponentInfo{"widget": childInfo},
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should have namespaced self declaration
	if !strings.Contains(src, "widget0_selfW := 0") {
		t.Errorf("expected namespaced self declaration 'widget0_selfW := 0':\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfChangeDetectionEmitted(t *testing.T) {
	// Given — self decl present
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		SelfDecls: []script.SelfDecl{{Name: "selfW", Key: "width"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should have previous self tracking variable
	if !strings.Contains(src, "prevSelfW") {
		t.Errorf("expected prevSelfW tracking variable:\n%s", src)
	}

	// Should have change detection with app.Dirty
	if !strings.Contains(src, "selfW != prevSelfW") {
		t.Errorf("expected self-width change detection:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfChangeDetectionNotEmittedWhenAbsent(t *testing.T) {
	// Given — no self decls
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if strings.Contains(src, "prevSelf") {
		t.Errorf("expected no self change detection when self decls absent:\n%s", src)
	}

	assertValidGo(t, out)
}

func TestSelfNotEmittedWhenAbsent(t *testing.T) {
	// Given — no self decls
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "hello"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if strings.Contains(src, "SelfW") || strings.Contains(src, "SelfH") {
		t.Errorf("expected no SelfW/SelfH when self decls absent:\n%s", src)
	}

	assertValidGo(t, out)
}
