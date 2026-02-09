package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

// textNode is a test helper that creates a TextElement with a single StringPart.
func textNode(s string) *template.TextElement {
	return &template.TextElement{Parts: []template.Part{&template.StringPart{Value: s}}}
}

func TestGenerateSingleTextElementIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

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

func TestGenerateTextElementUsesLayout(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "layout.KindText") {
		t.Errorf("expected layout.KindText in output:\n%s", src)
	}
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected Content: \"Hello\" in output:\n%s", src)
	}
	if !strings.Contains(src, "layout.Layout(") {
		t.Errorf("expected layout.Layout call in output:\n%s", src)
	}
}

func TestGenerateMultipleTextElements(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello"), textNode("World")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected Content: \"Hello\" in output:\n%s", src)
	}
	if !strings.Contains(src, `Content: "World"`) {
		t.Errorf("expected Content: \"World\" in output:\n%s", src)
	}
}

func TestGenerateContainsCorrectImports(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/render"`) {
		t.Errorf("expected runtime/render import in output:\n%s", src)
	}
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/layout"`) {
		t.Errorf("expected runtime/layout import in output:\n%s", src)
	}
	if !strings.Contains(src, `"os"`) {
		t.Errorf("expected os import in output:\n%s", src)
	}
	if !strings.Contains(src, `"bufio"`) {
		t.Errorf("expected bufio import in output:\n%s", src)
	}
}

func TestGenerateReferencesRuntimeRender(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
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
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "myapp"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "package myapp") {
		t.Errorf("expected 'package myapp' in output:\n%s", src)
	}
}
