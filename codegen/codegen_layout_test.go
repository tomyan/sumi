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
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "sumi.KindBox") {
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
	out, err := generateStatic(doc, nil, nil, "main")

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
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Border:    "single"`) {
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
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `sumi.ParsePadding("1 2")`) {
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
	out, err := generateStatic(doc, nil, nil, "main")

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

func TestGenerateBoxWithGap(t *testing.T) {
	// Given a box with gap attribute
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"gap": "2"},
				Children:   []template.Node{textNode("Hello"), textNode("World")},
			},
		},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "Gap:       2") {
		t.Errorf("expected Gap: 2 in output:\n%s", src)
	}
}

func TestGenerateBoxWithFlexGrow(t *testing.T) {
	// Given a box with flex-grow attribute
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"direction": "row"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"flex-grow": "1"},
						Children:   []template.Node{textNode("Grow")},
					},
				},
			},
		},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "FlexGrow:") {
		t.Errorf("expected FlexGrow in output:\n%s", src)
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

func TestGenerateWiresLayoutTreeIntoComponent(t *testing.T) {
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
	if !strings.Contains(src, "Tree: root,") {
		t.Errorf("expected the layout tree wired into the component:\n%s", src)
	}
}

func TestGenerateBoxWithScrollAttribute(t *testing.T) {
	// Given a box with a scroll expression attribute
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"overflow": "auto",
					"scroll":   "{myScroll}",
				},
				Children: []template.Node{textNode("content")},
			},
		},
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "Scroll:") || !strings.Contains(src, "myScroll,") {
		t.Errorf("expected Scroll: myScroll in output:\n%s", src)
	}
}
