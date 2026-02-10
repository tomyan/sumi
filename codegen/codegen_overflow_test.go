package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateBoxWithOverflowHidden(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"overflow": "hidden", "height": "5"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Overflow:`) || !strings.Contains(src, `"hidden"`) {
		t.Errorf("expected Overflow: \"hidden\" in output:\n%s", src)
	}
}

func TestGenerateBoxWithOverflowIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"overflow": "hidden",
					"height":   "5",
					"width":    "20",
					"border":   "single",
				},
				Children: []template.Node{textNode("Hello")},
			},
		},
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

func TestGenerateCallsLayoutRenderTree(t *testing.T) {
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
	// Should call layout.RenderTree instead of defining a local renderTree function
	if !strings.Contains(src, "layout.RenderTree(") {
		t.Errorf("expected layout.RenderTree call in output:\n%s", src)
	}
	// Should NOT define a local renderTree function
	if strings.Contains(src, "func renderTree(") {
		t.Errorf("should not define local renderTree function:\n%s", src)
	}
}
