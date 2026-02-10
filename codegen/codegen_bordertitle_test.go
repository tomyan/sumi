package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateBoxWithBorderTitle(t *testing.T) {
	// Given — inline border-title attribute
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"border":       "single",
					"border-title": "Panel",
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
	src := string(out)
	if !strings.Contains(src, `BorderTitle: "Panel"`) {
		t.Errorf("expected BorderTitle: \"Panel\" in output:\n%s", src)
	}
}

func TestGenerateBoxWithBorderTitleFromStylesheet(t *testing.T) {
	// Given — border-title in CSS
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "panel"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{
				Selector:   ".panel",
				Properties: map[string]string{"border": "single", "border-title": "Stats"},
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
	if !strings.Contains(src, `BorderTitle: "Stats"`) {
		t.Errorf("expected BorderTitle: \"Stats\" in output:\n%s", src)
	}
}

func TestGenerateWithBorderTitleIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"border":       "single",
					"border-title": "My Panel",
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
