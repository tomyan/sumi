package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateWithTitleEmitsOSC(t *testing.T) {
	// Given — a document with a <title> element
	doc := &template.Document{
		Children: []template.Node{
			&template.TitleElement{Parts: []template.Part{
				&template.StringPart{Value: "My App"},
			}},
			textNode("Hello"),
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
	if !strings.Contains(src, `\033]2;`) {
		t.Errorf("expected OSC title escape sequence in output:\n%s", src)
	}
}

func TestGenerateWithTitleExpressionUsesFormatting(t *testing.T) {
	// Given — a title with an expression
	doc := &template.Document{
		Children: []template.Node{
			&template.TitleElement{Parts: []template.Part{
				&template.ExprPart{Expr: "count"},
				&template.StringPart{Value: " items"},
			}},
			textNode("Hello"),
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
	src := string(out)
	if !strings.Contains(src, "fmt.Fprintf") {
		t.Errorf("expected fmt.Fprintf for dynamic title in output:\n%s", src)
	}
}

func TestGenerateWithStaticTitleSetsAppTitle(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TitleElement{Parts: []template.Part{
				&template.StringPart{Value: "My App"},
			}},
			textNode("Hello"),
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
	// Title save/restore is now handled by the App runtime via the Title field
	if !strings.Contains(src, `Title:`) {
		t.Errorf("expected Title field on App struct:\n%s", src)
	}
	if !strings.Contains(src, `"My App"`) {
		t.Errorf("expected title string in output:\n%s", src)
	}
}

func TestGenerateWithoutTitleNoOSC(t *testing.T) {
	// Given — no title element
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

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, `\033]2;`) {
		t.Errorf("expected no OSC sequence without <title>:\n%s", src)
	}
}

func TestGenerateWithTitleIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TitleElement{Parts: []template.Part{
				&template.ExprPart{Expr: "count"},
				&template.StringPart{Value: " items"},
			}},
			textNode("Hello"),
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
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}
