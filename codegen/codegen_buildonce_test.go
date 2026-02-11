package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestBuildOnceExtractsExpressionNode(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Count: "},
					&template.ExprPart{Expr: "count"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Expression text node should be extracted as a named variable
	if !strings.Contains(src, "node0 := &layout.Input{") {
		t.Errorf("expected extracted node0 variable:\n%s", src)
	}

	// Tree should be at function scope (root := before doRender :=)
	rootIdx := strings.Index(src, "root := &layout.Input{")
	doRenderIdx := strings.Index(src, "doRender := func()")
	if rootIdx < 0 || doRenderIdx < 0 || rootIdx >= doRenderIdx {
		t.Errorf("expected root tree before doRender closure:\n%s", src)
	}

	// Sync function should exist and patch Content
	if !strings.Contains(src, "sync := func() {") {
		t.Errorf("expected sync function:\n%s", src)
	}
	if !strings.Contains(src, `node0.Content = fmt.Sprintf("Count: %v", count)`) {
		t.Errorf("expected sync to patch node0.Content:\n%s", src)
	}

	// doRender should call sync
	if !strings.Contains(src, "sync()") {
		t.Errorf("expected doRender to call sync():\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceStaticTextNotExtracted(t *testing.T) {
	// Given — static text (no expressions) should NOT be extracted
	doc := &template.Document{
		Children: []template.Node{
			textNode("Hello"),
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Count: "},
					&template.ExprPart{Expr: "count"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Only one node should be extracted (the expression one)
	if strings.Contains(src, "node1 :=") {
		t.Errorf("should not extract static text nodes:\n%s", src)
	}
	// Static text should remain inline in the tree
	if !strings.Contains(src, `Content: "Hello"`) {
		t.Errorf("expected static text inline in tree:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceMultipleExpressionNodes(t *testing.T) {
	// Given — two expression text nodes should each get their own variable
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "X: "},
					&template.ExprPart{Expr: "x"},
				},
			},
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Y: "},
					&template.ExprPart{Expr: "y"},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "x", InitExpr: "0"},
			{Name: "y", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	if !strings.Contains(src, "node0 := &layout.Input{") {
		t.Errorf("expected node0:\n%s", src)
	}
	if !strings.Contains(src, "node1 := &layout.Input{") {
		t.Errorf("expected node1:\n%s", src)
	}

	// Both should have sync lines
	if !strings.Contains(src, `node0.Content = fmt.Sprintf("X: %v", x)`) {
		t.Errorf("expected sync for node0:\n%s", src)
	}
	if !strings.Contains(src, `node1.Content = fmt.Sprintf("Y: %v", y)`) {
		t.Errorf("expected sync for node1:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestBuildOnceTreeNotInsideDoRender(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then — tree should NOT be rebuilt inside doRender
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	doRenderIdx := strings.Index(src, "doRender := func()")
	if doRenderIdx < 0 {
		t.Fatalf("expected doRender closure:\n%s", src)
	}
	doRenderBody := src[doRenderIdx:]

	// The tree construction should NOT appear inside doRender
	if strings.Contains(doRenderBody, "root := &layout.Input{") {
		t.Errorf("tree should be built outside doRender, not inside:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}
