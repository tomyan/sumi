package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateEmitsCreateApp(t *testing.T) {
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
	if !strings.Contains(src, "func CreateApp(") {
		t.Errorf("expected CreateApp function in output:\n%s", src)
	}
}

func TestCreateAppReturnsTuiApp(t *testing.T) {
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
	if !strings.Contains(src, "*tui.App") {
		t.Errorf("expected *tui.App return type in CreateApp:\n%s", src)
	}
}

func TestCreateAppSetsTestDimensions(t *testing.T) {
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
	if !strings.Contains(src, "app.TestWidth = w") {
		t.Errorf("expected TestWidth assignment in CreateApp:\n%s", src)
	}
	if !strings.Contains(src, "app.TestHeight = h") {
		t.Errorf("expected TestHeight assignment in CreateApp:\n%s", src)
	}
}

func TestCreateAppIsValidGo(t *testing.T) {
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

func TestCreateAppReactiveIsValidGo(t *testing.T) {
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

func TestCreateAppHasBothFunctions(t *testing.T) {
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
	if !strings.Contains(src, "func Run()") {
		t.Errorf("expected Run() function in output:\n%s", src)
	}
	if !strings.Contains(src, "func CreateApp(") {
		t.Errorf("expected CreateApp() function in output:\n%s", src)
	}
}

func TestCreateAppCallsRender(t *testing.T) {
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
	if !strings.Contains(src, "app.Render()") {
		t.Errorf("expected app.Render() call in CreateApp:\n%s", src)
	}
}

func TestCreateAppInitializesTestBuffer(t *testing.T) {
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
	if !strings.Contains(src, "app.TestBuffer = render.NewBuffer(w, h)") {
		t.Errorf("expected TestBuffer initialization in CreateApp:\n%s", src)
	}
}
