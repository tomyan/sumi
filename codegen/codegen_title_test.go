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

func TestGenerateWithTitleRestoresOnExit(t *testing.T) {
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
	// Should save and restore title using xterm stack sequences
	if !strings.Contains(src, `\033[22;2t`) {
		t.Errorf("expected title save sequence in output:\n%s", src)
	}
	if !strings.Contains(src, `\033[23;2t`) {
		t.Errorf("expected title restore sequence in output:\n%s", src)
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
