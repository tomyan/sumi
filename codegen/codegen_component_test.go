package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateComponentWithPropsIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
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
}

func TestGenerateComponentStructName(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "type CounterComponent struct") {
		t.Errorf("expected struct named CounterComponent in output:\n%s", src)
	}
}

func TestGenerateComponentConstructor(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "func NewCounterComponent(label string)") {
		t.Errorf("expected NewCounterComponent constructor in output:\n%s", src)
	}
	if !strings.Contains(src, "*CounterComponent") {
		t.Errorf("expected *CounterComponent return type in output:\n%s", src)
	}
}

func TestGenerateComponentLayoutMethod(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "func (c *CounterComponent) Layout() *layout.Input") {
		t.Errorf("expected Layout method in output:\n%s", src)
	}
	if !strings.Contains(src, "layout.KindBox") {
		t.Errorf("expected layout.KindBox in Layout method output:\n%s", src)
	}
}

func TestGenerateComponentWithStateAndProps(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Count"`},
		},
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
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
	// Both label and count should appear as struct fields
	if !strings.Contains(src, "label string") {
		t.Errorf("expected label field in struct:\n%s", src)
	}
	if !strings.Contains(src, "count int") {
		t.Errorf("expected count field in struct:\n%s", src)
	}
}

func TestGenerateComponentHandleKeyGenerated(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "increment"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Count"`},
		},
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
		FuncDecls: []script.FuncDecl{
			{
				Name:   "increment",
				Params: "",
				Body:   "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{
					{VarName: "count", Line: "count = count + 1"},
				},
			},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "func (c *CounterComponent) HandleKey(key rune)") {
		t.Errorf("expected HandleKey method in output:\n%s", src)
	}
}

func TestGenerateComponentNoHandleKeyWithoutOnkey(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if strings.Contains(src, "HandleKey") {
		t.Errorf("expected no HandleKey method without onkey handler:\n%s", src)
	}
}

func TestGenerateComponentDirtyMethod(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "func (c *CounterComponent) Dirty() bool") {
		t.Errorf("expected Dirty method in output:\n%s", src)
	}
	if !strings.Contains(src, "c.dirty = false") {
		t.Errorf("expected dirty flag reset in Dirty method:\n%s", src)
	}
}

func TestGenerateComponentMethodHasDirtyFlag(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Count"`},
		},
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
		FuncDecls: []script.FuncDecl{
			{
				Name:   "increment",
				Params: "",
				Body:   "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{
					{VarName: "count", Line: "count = count + 1"},
				},
			},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "func (c *CounterComponent) increment()") {
		t.Errorf("expected increment as method on component:\n%s", src)
	}
	if !strings.Contains(src, "c.dirty = true") {
		t.Errorf("expected c.dirty = true in method body:\n%s", src)
	}
}

func TestGenerateComponentExprUsesReceiver(t *testing.T) {
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
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Count"`},
		},
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
	})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "c.count") {
		t.Errorf("expected c.count (receiver prefix) in expression:\n%s", src)
	}
}

func TestGenerateComponentWithStylesheet(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Attributes: map[string]string{"class": "title"},
				Parts:      []template.Part{&template.StringPart{Value: "Hello"}},
			},
		},
	}
	sc := &script.Script{
		PropDecls: []script.PropDecl{
			{Name: "label", DefaultExpr: `"Click me"`},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red", "bold": "true"}},
		},
	}

	// When
	out, err := Generate(doc, sc, ss, Options{
		PackageName:   "counter",
		ComponentName: "Counter",
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
	if !strings.Contains(src, "render.Style{") {
		t.Errorf("expected render.Style literal in component output:\n%s", src)
	}
	if !strings.Contains(src, "Bold: true") {
		t.Errorf("expected Bold: true in component style output:\n%s", src)
	}
}
