package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateBoxWithBorderCollapse(t *testing.T) {
	// Given — border-collapse in CSS
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "layout"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"border": "single"},
						Children:   []template.Node{textNode("A")},
					},
					&template.BoxElement{
						Attributes: map[string]string{"border": "single"},
						Children:   []template.Node{textNode("B")},
					},
				},
			},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{
				Selector:   ".layout",
				Properties: map[string]string{"border-collapse": "collapse"},
			},
		},
	}

	// When
	out, err := Generate(doc, nil, ss, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "BorderCollapse: true") {
		t.Errorf("expected BorderCollapse: true in output:\n%s", src)
	}
}

func TestGenerateWithBorderCollapseIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border-collapse": "collapse"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"border": "single"},
						Children:   []template.Node{textNode("A")},
					},
				},
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
