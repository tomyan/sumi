package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

// textNode is a test helper that creates a TextElement with a single StringPart.
func textNode(s string) *template.TextElement {
	return &template.TextElement{Parts: []template.Part{&template.StringPart{Value: s}}}
}

// --- Existing tests updated for Parts and new Generate signature ---

func TestGenerateSingleTextElementIsValidGo(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
		Children: []template.Node{textNode("Hello"), textNode("World")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "myapp"})
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
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
				Attributes: map[string]string{"direction": "column"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
				Attributes: map[string]string{"border": "single"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
				Attributes: map[string]string{"padding": "1 2"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
				Attributes: map[string]string{"width": "40", "height": "10"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
						Children:   []template.Node{textNode("Nested")},
					},
				},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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

// --- New tests for reactive codegen ---

func TestGenerateWithNilScriptIsBackwardsCompatible(t *testing.T) {
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	out, err := Generate(doc, nil, Options{PackageName: "main"})
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
	out, err := Generate(doc, sc, Options{PackageName: "main"})
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
	out, err := Generate(doc, sc, Options{PackageName: "main"})
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
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
	}
	out, err := Generate(doc, sc, Options{PackageName: "main"})
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
	out, err := Generate(doc, sc, Options{PackageName: "main"})
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
	out, err := Generate(doc, sc, Options{PackageName: "main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "dirty = true") {
		t.Errorf("expected dirty = true in function body:\n%s", src)
	}
}

func TestGenerateMultipleStateVarsAndFunctions(t *testing.T) {
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
	out, err := Generate(doc, sc, Options{PackageName: "main"})
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
