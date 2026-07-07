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
	out, err := generateStatic(doc, nil, nil, "main")

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
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "sumi.KindText") {
		t.Errorf("expected layout.KindText in output:\n%s", src)
	}
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected Content: \"Hello\" in output:\n%s", src)
	}
	if !strings.Contains(src, "Tree: root,") {
		t.Errorf("expected the layout tree wired into the component:\n%s", src)
	}
}

func TestGenerateMultipleTextElements(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello"), textNode("World")},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

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
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `sumi "github.com/tomyan/sumi/runtime/prelude"`) {
		t.Errorf("expected runtime/prelude import in output:\n%s", src)
	}
	// The component form runs under tui.Run, so the constructor does not
	// import os or wire its own render loop.
	if strings.Contains(src, `"os"`) {
		t.Errorf("component form should not import os:\n%s", src)
	}
}

func TestGenerateReturnsComponent(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "return &sumi.Component{") {
		t.Errorf("expected a *sumi.Component return in output:\n%s", src)
	}
}

func TestGeneratesComponentConstructor(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When — a script-free template still compiles to the component form
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "func NewApp(props AppProps) *sumi.Component") {
		t.Errorf("expected NewApp constructor in output:\n%s", src)
	}
	if !strings.Contains(src, "type AppProps struct") {
		t.Errorf("expected AppProps struct in output:\n%s", src)
	}
}

func TestStaticCodeNoInlineEventLoop(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// No inline event loop boilerplate
	if strings.Contains(src, "select {") {
		t.Errorf("should not have inline select in static output:\n%s", src)
	}
	if strings.Contains(src, "evt.Rune == 'q'") {
		t.Errorf("should not have hardcoded 'q' quit in output:\n%s", src)
	}
	if strings.Contains(src, "sumi.EnableRawMode") {
		t.Errorf("should not have EnableRawMode in generated code (moved to runtime):\n%s", src)
	}
}

func TestGenerateRespectsPackageName(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "myapp")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "package myapp") {
		t.Errorf("expected 'package myapp' in output:\n%s", src)
	}
}
