package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateDeclaresScrollState(t *testing.T) {
	// Given — a box with overflow=scroll
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"overflow": "scroll", "height": "10"},
				Children:   []template.Node{textNode("Content")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "var scroll0 layout.ScrollState") {
		t.Errorf("expected scroll state declaration in output:\n%s", src)
	}
}

func TestGenerateWiresScrollYIntoTree(t *testing.T) {
	// Given — a box with overflow=scroll
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"overflow": "scroll", "height": "10"},
				Children:   []template.Node{textNode("Content")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then — scroll state should be wired into tree before rendering
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "scroll0.ScrollY") {
		t.Errorf("expected scroll0.ScrollY wiring in output:\n%s", src)
	}
}

func TestGenerateArrowKeyDispatch(t *testing.T) {
	// Given — a box with overflow=scroll
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"overflow": "scroll", "height": "10"},
				Children:   []template.Node{textNode("Content")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then — should dispatch arrow keys for scrolling
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "input.EventSpecial") {
		t.Errorf("expected EventSpecial handling in output:\n%s", src)
	}
	if !strings.Contains(src, "input.KeyDown") {
		t.Errorf("expected KeyDown dispatch in output:\n%s", src)
	}
	if !strings.Contains(src, "input.KeyUp") {
		t.Errorf("expected KeyUp dispatch in output:\n%s", src)
	}
}

func TestGenerateScrollIsValidGo(t *testing.T) {
	// Given — a box with overflow=scroll
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"overflow": "scroll", "height": "10"},
				Children:   []template.Node{textNode("Content")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestGenerateNoScrollStateWithoutOverflow(t *testing.T) {
	// Given — no overflow boxes
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then — no scroll state declaration
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, "ScrollState") {
		t.Errorf("expected no ScrollState in output without overflow:\n%s", src)
	}
}

func TestGenerateRootSelectorAppliesOverflow(t *testing.T) {
	// Given — a root selector with overflow and min-width
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "root", Properties: map[string]string{
				"overflow":  "auto",
				"min-width": "48",
			}},
		},
	}

	// When
	out, err := Generate(doc, sc, ss, Options{PackageName: "main"})

	// Then — root container should have overflow and min-width
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Overflow:  "auto"`) {
		t.Errorf("expected Overflow on root container:\n%s", src)
	}
	if !strings.Contains(src, "MinWidth:  48") {
		t.Errorf("expected MinWidth: 48 on root container:\n%s", src)
	}
}

func TestGenerateRootSelectorIsValidGo(t *testing.T) {
	// Given — root selector with overflow
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "root", Properties: map[string]string{
				"overflow":  "auto",
				"min-width": "48",
			}},
		},
	}

	// When
	out, err := Generate(doc, sc, ss, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestGenerateRootSelectorCreatesScrollState(t *testing.T) {
	// Given — root selector with overflow=auto triggers scroll state for root
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "root", Properties: map[string]string{
				"overflow":  "auto",
				"min-width": "48",
			}},
		},
	}

	// When
	out, err := Generate(doc, sc, ss, Options{PackageName: "main"})

	// Then — scroll state should exist for the root container
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "var scroll0 layout.ScrollState") {
		t.Errorf("expected scroll state for root container:\n%s", src)
	}
}
