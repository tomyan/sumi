package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateIfProducesValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "count > 0",
				Then:      []template.Node{textNode("Has items")},
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

func TestGenerateIfContainsCondition(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "count > 0",
				Then:      []template.Node{textNode("Has items")},
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
	if !strings.Contains(src, "if count > 0 {") {
		t.Errorf("expected 'if count > 0 {' in output:\n%s", src)
	}
}

func TestGenerateIfElseProducesValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "count > 0",
				Then:      []template.Node{textNode("Has items")},
				Else:      []template.Node{textNode("Empty")},
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

func TestGenerateIfElseContainsBranches(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "count > 0",
				Then:      []template.Node{textNode("Has items")},
				Else:      []template.Node{textNode("Empty")},
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
	if !strings.Contains(src, "} else {") {
		t.Errorf("expected '} else {' in output:\n%s", src)
	}
}

func TestGenerateForProducesValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause:   "i, item := range items",
				Children: []template.Node{textNode("item")},
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

func TestGenerateForContainsLoop(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause:   "i, item := range items",
				Children: []template.Node{textNode("item")},
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
	if !strings.Contains(src, "for i, item := range items {") {
		t.Errorf("expected 'for i, item := range items {' in output:\n%s", src)
	}
}

func TestGenerateForContainingIfProducesValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause: "i, item := range items",
				Children: []template.Node{
					&template.IfNode{
						Condition: "i > 0",
						Then:      []template.Node{textNode("sep")},
					},
					textNode("item"),
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

func TestGenerateNestedForProducesValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause: "_, row := range rows",
				Children: []template.Node{
					&template.ForNode{
						Clause:   "_, col := range row",
						Children: []template.Node{textNode("cell")},
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

func TestGenerateIfAtRootWithBoxChildren(t *testing.T) {
	// Given - {if} at root level alongside a box
	doc := &template.Document{
		Children: []template.Node{
			textNode("Header"),
			&template.IfNode{
				Condition: "showBody",
				Then: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"border": "single"},
						Children:   []template.Node{textNode("Body")},
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

func TestGenerateStaticChildrenUnchanged(t *testing.T) {
	// Given - no control flow, just static children
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{},
				Children:   []template.Node{textNode("Static")},
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
	// Static children should use inline literal, not IIFE
	if strings.Contains(src, "func() []*layout.Input") {
		t.Errorf("static children should not use IIFE pattern:\n%s", src)
	}
	if !strings.Contains(src, "Children: []*layout.Input{") {
		t.Errorf("expected inline Children literal in output:\n%s", src)
	}
}
