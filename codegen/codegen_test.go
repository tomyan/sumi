package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateSingleTextElementIsValidGo(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestGenerateMultipleTextElementsEachOnOwnRow(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
			&template.TextElement{Content: "World"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// First element at row 0, second at row 1
	if !strings.Contains(src, `WriteText(0, 0, "Hello")`) {
		t.Errorf("expected WriteText(0, 0, \"Hello\") in output:\n%s", src)
	}
	if !strings.Contains(src, `WriteText(1, 0, "World")`) {
		t.Errorf("expected WriteText(1, 0, \"World\") in output:\n%s", src)
	}
}

func TestGenerateContainsCorrectImports(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/render"`) {
		t.Errorf("expected runtime/render import in output:\n%s", src)
	}
	if !strings.Contains(src, `"os"`) {
		t.Errorf("expected os import in output:\n%s", src)
	}
	if !strings.Contains(src, `"bufio"`) {
		t.Errorf("expected bufio import in output:\n%s", src)
	}
}

func TestGenerateReferencesRuntimeRender(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "render.NewBuffer(") {
		t.Errorf("expected render.NewBuffer call in output:\n%s", src)
	}
	if !strings.Contains(src, "render.EnterAlternateScreen(") {
		t.Errorf("expected render.EnterAlternateScreen call in output:\n%s", src)
	}
	if !strings.Contains(src, "render.ExitAlternateScreen(") {
		t.Errorf("expected render.ExitAlternateScreen call in output:\n%s", src)
	}
}

func TestGenerateRespectsPackageName(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "package myapp") {
		t.Errorf("expected 'package myapp' in output:\n%s", src)
	}
}
