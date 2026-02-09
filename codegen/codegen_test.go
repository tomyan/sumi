package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateSingleTextElementIsValidGo(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestGenerateTextElementUsesLayout(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "layout.KindText") {
		t.Errorf("expected layout.KindText in output:\n%s", src)
	}
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected Content: \"Hello\" in output:\n%s", src)
	}
	if !strings.Contains(src, "layout.Layout(") {
		t.Errorf("expected layout.Layout call in output:\n%s", src)
	}
}

func TestGenerateMultipleTextElements(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
			&template.TextElement{Content: "World"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected Content: \"Hello\" in output:\n%s", src)
	}
	if !strings.Contains(src, `Content: "World"`) {
		t.Errorf("expected Content: \"World\" in output:\n%s", src)
	}
}

func TestGenerateContainsCorrectImports(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/render"`) {
		t.Errorf("expected runtime/render import in output:\n%s", src)
	}
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/layout"`) {
		t.Errorf("expected runtime/layout import in output:\n%s", src)
	}
	if !strings.Contains(src, `"os"`) {
		t.Errorf("expected os import in output:\n%s", src)
	}
	if !strings.Contains(src, `"bufio"`) {
		t.Errorf("expected bufio import in output:\n%s", src)
	}
}

func TestGenerateReferencesRuntimeRender(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "render.NewBuffer(") {
		t.Errorf("expected render.NewBuffer call in output:\n%s", src)
	}
	if !strings.Contains(src, "render.EnterAlternateScreen(") {
		t.Errorf("expected render.EnterAlternateScreen call in output:\n%s", src)
	}
	if !strings.Contains(src, "render.ExitAlternateScreen(") {
		t.Errorf("expected render.ExitAlternateScreen call in output:\n%s", src)
	}
}

func TestGenerateRespectsPackageName(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "myapp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "package myapp") {
		t.Errorf("expected 'package myapp' in output:\n%s", src)
	}
}

func TestGenerateBoxContainingTextIsValidGo(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
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
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "layout.KindBox") {
		t.Errorf("expected layout.KindBox in output:\n%s", src)
	}
}

func TestGenerateBoxWithAttributesDirection(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"direction": "column",
				},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Direction: "column"`) {
		t.Errorf("expected Direction: \"column\" in output:\n%s", src)
	}
}

func TestGenerateBoxWithBorder(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"border": "single",
				},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `Border: "single"`) {
		t.Errorf("expected Border: \"single\" in output:\n%s", src)
	}
}

func TestGenerateBoxWithPadding(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"padding": "1 2",
				},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, `layout.ParsePadding("1 2")`) {
		t.Errorf("expected layout.ParsePadding(\"1 2\") in output:\n%s", src)
	}
}

func TestGenerateBoxWithWidthAndHeight(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{
					"width":  "40",
					"height": "10",
				},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
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
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children: []template.Node{
					&template.BoxElement{
						Attributes: map[string]string{"padding": "1"},
						Children: []template.Node{
							&template.TextElement{Content: "Nested"},
						},
					},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
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
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{Content: "Hello"},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "renderTree(") {
		t.Errorf("expected renderTree call in output:\n%s", src)
	}
	// Should contain the renderTree function definition
	if !strings.Contains(src, "func renderTree(") {
		t.Errorf("expected renderTree function definition in output:\n%s", src)
	}
}

func TestGenerateRenderTreeDrawsBorders(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children: []template.Node{
					&template.TextElement{Content: "Hello"},
				},
			},
		},
	}
	out, err := Generate(doc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "DrawBorder(") {
		t.Errorf("expected DrawBorder call in renderTree in output:\n%s", src)
	}
	if !strings.Contains(src, "WriteText(") {
		t.Errorf("expected WriteText call in renderTree in output:\n%s", src)
	}
}
