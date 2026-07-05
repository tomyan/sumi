package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

func mustParseStylesheet(t *testing.T, src string) *style.Stylesheet {
	t.Helper()
	ss, err := style.Parse(src)
	if err != nil {
		t.Fatalf("stylesheet parse error: %v", err)
	}
	return ss
}

func TestGenerateWithNilStylesheetBackwardCompat(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	// Should call layout.RenderTree which handles styled rendering internally
	if !strings.Contains(src, "sumi.RenderTree(") {
		t.Errorf("expected layout.RenderTree in output:\n%s", src)
	}
}

func TestGenerateWithStylesheetAndClassOnText(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Attributes: map[string]string{"class": "title"},
				Parts:      []template.Part{&template.StringPart{Value: "Hello"}},
			},
		},
	}
	ss := mustParseStylesheet(t, `.title { color: red; font-weight: bold; }`)

	// When
	out, err := Generate(doc, nil, ss, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, `sumi.Style{`) {
		t.Errorf("expected render.Style literal in output:\n%s", src)
	}
	if !strings.Contains(src, `FG:`) {
		t.Errorf("expected FG field in Style literal:\n%s", src)
	}
	if !strings.Contains(src, `Bold: true`) {
		t.Errorf("expected Bold: true in Style literal:\n%s", src)
	}
}

func TestGenerateStylesheetLayoutProperties(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "container"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	ss := mustParseStylesheet(t, `.container { border: single; padding: 1 2; }`)

	// When
	out, err := Generate(doc, nil, ss, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, `"single"`) || !strings.Contains(src, "Border:") {
		t.Errorf("expected Border with single from stylesheet in output:\n%s", src)
	}
	if !strings.Contains(src, `sumi.ParsePadding("1 2")`) {
		t.Errorf("expected ParsePadding from stylesheet in output:\n%s", src)
	}
}

func TestGenerateInlineAttributeOverridesStylesheet(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "container", "border": "double"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	ss := mustParseStylesheet(t, `.container { border: single; padding: 1; }`)

	// When
	out, err := Generate(doc, nil, ss, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// Inline "double" should override stylesheet "single"
	if !strings.Contains(src, `"double"`) || !strings.Contains(src, "Border:") {
		t.Errorf("expected Border with double (inline override) in output:\n%s", src)
	}
	// Should NOT contain "single" since inline overrides it
	if strings.Contains(src, `"single"`) {
		t.Errorf("expected inline border to override stylesheet, but found single in output:\n%s", src)
	}
	// Stylesheet padding should still apply
	if !strings.Contains(src, `sumi.ParsePadding("1")`) {
		t.Errorf("expected ParsePadding from stylesheet in output:\n%s", src)
	}
}

func TestGenerateElementSelectorStylesheet(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Attributes: map[string]string{},
				Parts:      []template.Part{&template.StringPart{Value: "Hello"}},
			},
		},
	}
	ss := mustParseStylesheet(t, `text { color: green; }`)

	// When
	out, err := Generate(doc, nil, ss, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, `sumi.Style{`) {
		t.Errorf("expected render.Style literal for element selector:\n%s", src)
	}
	if !strings.Contains(src, `"green"`) {
		t.Errorf("expected green color in Style literal:\n%s", src)
	}
}

func TestGenerateUsesLayoutRenderTreeForStyling(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// Rendering is now delegated to layout.RenderTree which handles styled methods
	if !strings.Contains(src, "sumi.RenderTree(") {
		t.Errorf("expected layout.RenderTree call in output:\n%s", src)
	}
}
