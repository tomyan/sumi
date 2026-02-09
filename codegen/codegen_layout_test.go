package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateBoxContainingTextIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{},
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
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestGenerateBoxUsesLayoutKindBox(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{},
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
	if !strings.Contains(src, "layout.KindBox") {
		t.Errorf("expected layout.KindBox in output:\n%s", src)
	}
}

func TestGenerateBoxWithAttributesDirection(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "column"},
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
	if !strings.Contains(src, `Direction: "column"`) {
		t.Errorf("expected Direction: \"column\" in output:\n%s", src)
	}
}

func TestGenerateBoxWithBorder(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
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
	if !strings.Contains(src, `Border: "single"`) {
		t.Errorf("expected Border: \"single\" in output:\n%s", src)
	}
}

func TestGenerateBoxWithPadding(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"padding": "1 2"},
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
	if !strings.Contains(src, `layout.ParsePadding("1 2")`) {
		t.Errorf("expected layout.ParsePadding(\"1 2\") in output:\n%s", src)
	}
}

func TestGenerateBoxWithWidthAndHeight(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"width": "40", "height": "10"},
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
	if !strings.Contains(src, "FixedWidth:  40") {
		t.Errorf("expected FixedWidth: 40 in output:\n%s", src)
	}
	if !strings.Contains(src, "FixedHeight: 10") {
		t.Errorf("expected FixedHeight: 10 in output:\n%s", src)
	}
}

func TestGenerateNestedBoxesIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"padding": "1"},
						Children:   []template.Node{textNode("Nested")},
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

func TestGenerateContainsRenderTree(t *testing.T) {
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
	if !strings.Contains(src, "renderTree(") {
		t.Errorf("expected renderTree call in output:\n%s", src)
	}
	if !strings.Contains(src, "func renderTree(") {
		t.Errorf("expected renderTree function definition in output:\n%s", src)
	}
}

func TestGenerateRenderTreeDrawsBorders(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
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
	if !strings.Contains(src, "DrawStyledBorder(") {
		t.Errorf("expected DrawStyledBorder call in renderTree in output:\n%s", src)
	}
	if !strings.Contains(src, "WriteStyledText(") {
		t.Errorf("expected WriteStyledText call in renderTree in output:\n%s", src)
	}
}
