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

func TestGenerateRenderTreeWithClipping(t *testing.T) {
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
	// renderTree should accept a clip parameter and pass it to clipped write methods
	if !strings.Contains(src, "func renderTree(buf *render.Buffer, box *layout.Box, clip *render.Clip)") {
		t.Errorf("expected renderTree to accept clip parameter:\n%s", src)
	}
	if !strings.Contains(src, "WriteStyledTextClipped(") {
		t.Errorf("expected WriteStyledTextClipped call in renderTree:\n%s", src)
	}
}
