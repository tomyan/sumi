package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestStaticCodeHasTestModeDimensions(t *testing.T) {
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
	if !strings.Contains(src, "app.TestWidth > 0") {
		t.Errorf("expected test-mode dimension check in static output:\n%s", src)
	}
}

func TestStaticCodeHasTestModeBuffer(t *testing.T) {
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
	if !strings.Contains(src, "app.TestBuffer") {
		t.Errorf("expected test-mode buffer check in static output:\n%s", src)
	}
}

func TestStaticCodeWithTestModeIsValidGo(t *testing.T) {
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
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestReactiveCodeHasTestModeDimensions(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children:   []template.Node{textNode("Count: {count}")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
		FuncDecls:  []script.FuncDecl{{Name: "handleKey", Body: "count++\n", StateAssignments: []script.StateAssignment{{Line: "count++"}}}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "app.TestWidth > 0") {
		t.Errorf("expected test-mode dimension check in reactive output:\n%s", src)
	}
}

func TestReactiveCodeHasTestModeBuffer(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children:   []template.Node{textNode("Count: {count}")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
		FuncDecls:  []script.FuncDecl{{Name: "handleKey", Body: "count++\n", StateAssignments: []script.StateAssignment{{Line: "count++"}}}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "app.TestBuffer") {
		t.Errorf("expected test-mode buffer check in reactive output:\n%s", src)
	}
}

func TestReactiveCodeWithTestModeIsValidGo(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children:   []template.Node{textNode("Count: {count}")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{{Name: "count", InitExpr: "0"}},
		FuncDecls:  []script.FuncDecl{{Name: "handleKey", Body: "count++\n", StateAssignments: []script.StateAssignment{{Line: "count++"}}}},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

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
