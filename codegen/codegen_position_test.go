package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestCodegenPositionRelative(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"position": "relative",
					"top":      "2",
					"left":     "3",
				},
				Children: []template.Node{textNode("Shifted")},
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
	if !strings.Contains(src, `Position: "relative"`) {
		t.Errorf("expected Position: \"relative\" in output:\n%s", src)
	}
	if !strings.Contains(src, "Top:") {
		t.Errorf("expected Top field in output:\n%s", src)
	}
	if !strings.Contains(src, "Left:") {
		t.Errorf("expected Left field in output:\n%s", src)
	}
	fset := token.NewFileSet()
	if _, parseErr := parser.ParseFile(fset, "gen.go", out, parser.AllErrors); parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", src, parseErr)
	}
}

func TestCodegenDisplayNone(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"display": "none"},
				Children:   []template.Node{textNode("Hidden")},
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
	if !strings.Contains(src, `Display: "none"`) {
		t.Errorf("expected Display: \"none\" in output:\n%s", src)
	}
	fset := token.NewFileSet()
	if _, parseErr := parser.ParseFile(fset, "gen.go", out, parser.AllErrors); parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", src, parseErr)
	}
}
