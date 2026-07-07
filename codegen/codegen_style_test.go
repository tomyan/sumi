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
	src := string(out)
	// The component wires the layout tree; rendering is the runtime's job.
	if !strings.Contains(src, "Tree: root,") {
		t.Errorf("expected the layout tree wired into the component:\n%s", src)
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
	out, err := generateStatic(doc, nil, ss, "main")

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
	if strings.Contains(src, `sumi.Style{`) {
		t.Errorf("styles must not be baked into literals (runtime resolution):\n%s", src)
	}
	if !strings.Contains(src, "MustParseStylesheet") || !strings.Contains(src, "font-weight: bold") {
		t.Errorf("expected embedded stylesheet with the rule:\n%s", src)
	}
	if !containsField(src, "Classes", `[]string{"title"}`) {
		t.Errorf("expected Classes identity for runtime matching:\n%s", src)
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
	out, err := generateStatic(doc, nil, ss, "main")

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
	if containsField(src, "Border", `"single"`) {
		t.Errorf("CSS layout props must not be baked into literals:\n%s", src)
	}
	if !strings.Contains(src, "MustParseStylesheet") || !strings.Contains(src, "border: single") {
		t.Errorf("expected embedded stylesheet with layout rules:\n%s", src)
	}
	if !strings.Contains(src, "Stylesheet: sumi.MustParseStylesheet(") {
		t.Errorf("component must carry its stylesheet for runtime resolution:\n%s", src)
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
	out, err := generateStatic(doc, nil, ss, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// Inline "double" is emitted in the literal; runtime resolution must not
	// override it (covered by layout.ResolveStyles inline-precedence tests).
	if !containsField(src, "Border", `"double"`) {
		t.Errorf("expected inline Border double in literal:\n%s", src)
	}
	if !containsField(src, "Attrs", "map[string]string{") {
		t.Errorf("expected Attrs identity so the resolver sees the inline override:\n%s", src)
	}
	if !strings.Contains(src, "border: single") {
		t.Errorf("stylesheet rule must still be embedded:\n%s", src)
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
	out, err := generateStatic(doc, nil, ss, "main")

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
	if strings.Contains(src, `sumi.Style{`) {
		t.Errorf("styles must not be baked into literals:\n%s", src)
	}
	if !strings.Contains(src, "color: green") {
		t.Errorf("expected element-selector rule in embedded stylesheet:\n%s", src)
	}
}

func TestGenerateDelegatesStylingToRuntime(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Attributes: map[string]string{"class": "title"},
				Parts:      []template.Part{&template.StringPart{Value: "Hello"}},
			},
		},
	}
	ss := mustParseStylesheet(t, `.title { color: red; }`)

	// When
	out, err := generateStatic(doc, nil, ss, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// Styling is delegated to the runtime via the embedded stylesheet, not
	// baked into cell literals.
	if strings.Contains(src, "sumi.Style{") {
		t.Errorf("styles must not be baked into literals:\n%s", src)
	}
	if !strings.Contains(src, "Stylesheet: sumi.MustParseStylesheet(") {
		t.Errorf("component must carry its stylesheet for runtime resolution:\n%s", src)
	}
}
