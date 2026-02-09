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

// textNode is a test helper that creates a TextElement with a single StringPart.
func textNode(s string) *template.TextElement {
	return &template.TextElement{Parts: []template.Part{&template.StringPart{Value: s}}}
}

// --- Existing tests updated for Parts and new Generate signature ---

func TestGenerateSingleTextElementIsValidGo(t *testing.T) {
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

func TestGenerateTextElementUsesLayout(t *testing.T) {
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
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello"), textNode("World")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "main"})

	// Then
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
	// Given
	doc := &template.Document{
		Children: []template.Node{textNode("Hello")},
	}

	// When
	out, err := Generate(doc, nil, nil, Options{PackageName: "myapp"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	if !strings.Contains(src, "package myapp") {
		t.Errorf("expected 'package myapp' in output:\n%s", src)
	}
}

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

// --- New tests for reactive codegen ---

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

// --- Stylesheet tests ---

func TestGenerateWithNilStylesheetBackwardCompat(t *testing.T) {
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
	src := string(out)
	// Should use WriteStyledText and DrawStyledBorder (always use styled versions)
	if !strings.Contains(src, "WriteStyledText(") {
		t.Errorf("expected WriteStyledText in output:\n%s", src)
	}
}

func TestGenerateWithStylesheetAndClassOnText(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Attributes: map[string]string{"class": "title"},
				Parts:      []template.Part{&template.StringPart{Value: "Hello"}},
			},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".title", Properties: map[string]string{"color": "red", "bold": "true"}},
		},
	}

	// When
	out, err := Generate(doc, nil, ss, Options{PackageName: "main"})

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
	if !strings.Contains(src, `render.Style{`) {
		t.Errorf("expected render.Style literal in output:\n%s", src)
	}
	if !strings.Contains(src, `FG:`) {
		t.Errorf("expected FG field in Style literal:\n%s", src)
	}
	if !strings.Contains(src, `Bold: true`) {
		t.Errorf("expected Bold: true in Style literal:\n%s", src)
	}
}

func TestGenerateStylesheetLayoutProperties(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "container"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".container", Properties: map[string]string{
				"border":  "single",
				"padding": "1 2",
			}},
		},
	}

	// When
	out, err := Generate(doc, nil, ss, Options{PackageName: "main"})

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
	if !strings.Contains(src, `"single"`) || !strings.Contains(src, "Border:") {
		t.Errorf("expected Border with single from stylesheet in output:\n%s", src)
	}
	if !strings.Contains(src, `layout.ParsePadding("1 2")`) {
		t.Errorf("expected ParsePadding from stylesheet in output:\n%s", src)
	}
}

func TestGenerateInlineAttributeOverridesStylesheet(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"class": "container", "border": "double"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: ".container", Properties: map[string]string{
				"border":  "single",
				"padding": "1",
			}},
		},
	}

	// When
	out, err := Generate(doc, nil, ss, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)
	// Inline "double" should override stylesheet "single"
	if !strings.Contains(src, `"double"`) || !strings.Contains(src, "Border:") {
		t.Errorf("expected Border with double (inline override) in output:\n%s", src)
	}
	// Should NOT contain "single" since inline overrides it
	if strings.Contains(src, `"single"`) {
		t.Errorf("expected inline border to override stylesheet, but found single in output:\n%s", src)
	}
	// Stylesheet padding should still apply
	if !strings.Contains(src, `layout.ParsePadding("1")`) {
		t.Errorf("expected ParsePadding from stylesheet in output:\n%s", src)
	}
}

func TestGenerateElementSelectorStylesheet(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Attributes: map[string]string{},
				Parts:      []template.Part{&template.StringPart{Value: "Hello"}},
			},
		},
	}
	ss := &style.Stylesheet{
		Rules: []style.Rule{
			{Selector: "text", Properties: map[string]string{"color": "green"}},
		},
	}

	// When
	out, err := Generate(doc, nil, ss, Options{PackageName: "main"})

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
	if !strings.Contains(src, `render.Style{`) {
		t.Errorf("expected render.Style literal for element selector:\n%s", src)
	}
	if !strings.Contains(src, `"green"`) {
		t.Errorf("expected green color in Style literal:\n%s", src)
	}
}

func TestGenerateRenderTreeUsesStyledMethods(t *testing.T) {
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
	if !strings.Contains(src, "WriteStyledText(") {
		t.Errorf("expected WriteStyledText in renderTree:\n%s", src)
	}
	if !strings.Contains(src, "DrawStyledBorder(") {
		t.Errorf("expected DrawStyledBorder in renderTree:\n%s", src)
	}
}
