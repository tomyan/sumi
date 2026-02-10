package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateWithNilScriptIsBackwardsCompatible(t *testing.T) {
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
	// Static mode: uses bufio.Scanner to wait for Enter
	if !strings.Contains(src, "bufio.NewScanner") {
		t.Errorf("expected bufio.NewScanner in static mode output:\n%s", src)
	}
	// Should NOT contain event loop or input package
	if strings.Contains(src, "input.ReadKey") {
		t.Errorf("unexpected input.ReadKey in static mode output:\n%s", src)
	}
}

func TestGenerateWithStateDeclaration(t *testing.T) {
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
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
	src := string(out)
	if !strings.Contains(src, "count := 0") {
		t.Errorf("expected state variable declaration in output:\n%s", src)
	}
}

func TestGenerateWithExpressionUsesFmtSprintf(t *testing.T) {
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
	if !strings.Contains(src, "fmt.Sprintf") {
		t.Errorf("expected fmt.Sprintf for expression in output:\n%s", src)
	}
	if !strings.Contains(src, `"fmt"`) {
		t.Errorf("expected fmt import in output:\n%s", src)
	}
}

func TestGenerateWithStateContainsEventLoop(t *testing.T) {
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

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "input.ReadKey") {
		t.Errorf("expected input.ReadKey in reactive mode output:\n%s", src)
	}
	if !strings.Contains(src, "input.EnableRawMode") {
		t.Errorf("expected input.EnableRawMode in reactive mode output:\n%s", src)
	}
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/input"`) {
		t.Errorf("expected runtime/input import in output:\n%s", src)
	}
}

func TestGenerateWithOnkeyHandler(t *testing.T) {
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
	src := string(out)
	if !strings.Contains(src, "increment()") {
		t.Errorf("expected increment() call in event loop:\n%s", src)
	}
}

func TestGenerateWithFunctionSetsDirecty(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
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
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "dirty = true") {
		t.Errorf("expected dirty = true in function body:\n%s", src)
	}
}

func TestGenerateUsesTermGetSize(t *testing.T) {
	// Given a reactive document
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

	// Then generated code should use term.GetSize instead of hardcoded 80, 24
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "term.GetSize") {
		t.Errorf("expected term.GetSize in output:\n%s", src)
	}
	if strings.Contains(src, "layout.Layout(root, 80, 24)") {
		t.Errorf("should not hardcode 80, 24 in layout call:\n%s", src)
	}
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/term"`) {
		t.Errorf("expected runtime/term import in output:\n%s", src)
	}
}

func TestGenerateStaticUsesTermGetSize(t *testing.T) {
	// Given a static document (no script)
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then generated code should use term.GetSize
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "term.GetSize") {
		t.Errorf("expected term.GetSize in output:\n%s", src)
	}
	if strings.Contains(src, "layout.Layout(root, 80, 24)") {
		t.Errorf("should not hardcode 80, 24:\n%s", src)
	}
}

func TestGenerateMultipleStateVarsAndFunctions(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "X: "},
					&template.ExprPart{Expr: "x"},
					&template.StringPart{Value: " Y: "},
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
		FuncDecls: []script.FuncDecl{
			{
				Name: "incX", Params: "", Body: "\n\tx = x + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "x", Line: "x = x + 1"}},
			},
			{
				Name: "incY", Params: "", Body: "\n\ty = y + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "y", Line: "y = y + 1"}},
			},
		},
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
	src := string(out)
	if !strings.Contains(src, "x := 0") {
		t.Errorf("expected x := 0 in output:\n%s", src)
	}
	if !strings.Contains(src, "y := 0") {
		t.Errorf("expected y := 0 in output:\n%s", src)
	}
}
