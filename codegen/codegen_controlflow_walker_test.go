package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestDocHasExprsInsideIfNode(t *testing.T) {
	// Given - expression inside an IfNode
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "x",
				Then: []template.Node{
					&template.TextElement{
						Parts: []template.Part{
							&template.ExprPart{Expr: "count"},
						},
					},
				},
			},
		},
	}

	// When
	result := docHasExprs(doc)

	// Then
	if !result {
		t.Error("expected docHasExprs to return true for expression inside IfNode")
	}
}

func TestNodeHasStylesInsideForNode(t *testing.T) {
	// Given - a ForNode containing elements (styles come from stylesheet)
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause:   "i := range items",
				Children: []template.Node{textNode("item")},
			},
		},
	}

	// When - with nil stylesheet, no styles
	result := docHasStyles(doc, nil)

	// Then
	if result {
		t.Error("expected docHasStyles to return false with nil stylesheet")
	}
}

func TestWalkNodeFindsComponentInsideIf(t *testing.T) {
	// Given
	components := map[string]*ComponentInfo{
		"counter": {ExportedName: "Counter", Props: nil},
	}
	doc := &template.Document{
		Children: []template.Node{
			&template.IfNode{
				Condition: "visible",
				Then: []template.Node{
					&template.ComponentElement{
						Name:       "counter",
						Attributes: map[string]string{},
					},
				},
			},
		},
	}

	// When
	instances := collectComponentInstances(doc, components)

	// Then
	if len(instances) != 1 {
		t.Fatalf("got %d instances, want 1", len(instances))
	}
	if instances[0].VarName != "counter0" {
		t.Errorf("VarName = %q, want %q", instances[0].VarName, "counter0")
	}
}

func TestGenerateComponentWithIfProducesValidGo(t *testing.T) {
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
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := generateComponent(doc, sc, nil, Options{
		PackageName:   "main",
		ComponentName: "Test",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "if c.count > 0 {") {
		t.Errorf("expected 'if c.count > 0 {' in output:\n%s", src)
	}
}

func TestGenerateComponentWithForProducesValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause:   "i, item := range items",
				Children: []template.Node{textNode("item")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "items", InitExpr: `[]string{"a", "b"}`},
		},
	}

	// When
	out, err := generateComponent(doc, sc, nil, Options{
		PackageName:   "main",
		ComponentName: "Test",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "for i, item := range c.items {") {
		t.Errorf("expected 'for i, item := range c.items {' in output:\n%s", src)
	}
}

func TestGenerateComponentForWithKey(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.ForNode{
				Clause:   "i, item := range items",
				Key:      "item.ID",
				Children: []template.Node{textNode("item")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "items", InitExpr: `[]string{"a", "b"}`},
		},
	}

	// When
	out, err := generateComponent(doc, sc, nil, Options{
		PackageName:   "main",
		ComponentName: "Test",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "cs[len(cs)-1].Key = fmt.Sprint(") {
		t.Errorf("expected Key assignment in component output:\n%s", src)
	}
}
