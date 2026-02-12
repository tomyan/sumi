package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestGenerateReactiveUsesApp(t *testing.T) {
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
	if !strings.Contains(src, "tui.App") {
		t.Errorf("expected tui.App in reactive output:\n%s", src)
	}
	if !strings.Contains(src, "app.Run()") {
		t.Errorf("expected app.Run() in reactive output:\n%s", src)
	}
}

func TestGenerateReactiveNoInlineEventLoop(t *testing.T) {
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
	if strings.Contains(src, "select {") {
		t.Errorf("should not have inline select in output:\n%s", src)
	}
	if strings.Contains(src, "chan input.Event") {
		t.Errorf("should not have event channel in output:\n%s", src)
	}
	if strings.Contains(src, "input.EnableRawMode") {
		t.Errorf("should not have EnableRawMode in output:\n%s", src)
	}
	if strings.Contains(src, "evt.Rune == 'q'") {
		t.Errorf("should not have hardcoded 'q' quit:\n%s", src)
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
	assertValidGo(t, out)
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

func TestGenerateOnkeyStillFiresOnEventKey(t *testing.T) {
	// Given
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
		FuncDecls: []script.FuncDecl{
			{Name: "handleKey", Params: "", Body: "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "count", Line: "count = count + 1"}}},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "handleKey()") {
		t.Errorf("expected handleKey() call in event handler:\n%s", src)
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
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "increment()") {
		t.Errorf("expected increment() call in event loop:\n%s", src)
	}
}

func TestGenerateWithFunctionSetsAppDirty(t *testing.T) {
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
	if !strings.Contains(src, "app.Dirty = true") {
		t.Errorf("expected app.Dirty = true in function body:\n%s", src)
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

func TestGenerateWithEnvDeclsIsReactive(t *testing.T) {
	// Given: only env decls, no state — should still be reactive
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Terminal: "},
					&template.ExprPart{Expr: "width"},
					&template.StringPart{Value: "x"},
					&template.ExprPart{Expr: "height"},
				},
			},
		},
	}
	sc := &script.Script{
		EnvDecls: []script.EnvDecl{
			{Name: "width", Key: "width"},
			{Name: "height", Key: "height"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	// Should use tui.App
	if !strings.Contains(src, "tui.App") {
		t.Errorf("expected tui.App in output:\n%s", src)
	}
	// Env vars should be initialized
	if !strings.Contains(src, "width, height := term.GetSize") {
		t.Errorf("expected width, height := term.GetSize in output:\n%s", src)
	}
	// On resize: env vars should be updated via OnResize
	if !strings.Contains(src, "width, height = term.GetSize") {
		t.Errorf("expected width, height = term.GetSize on resize:\n%s", src)
	}
}

func TestGenerateWithEnvWidthOnly(t *testing.T) {
	// Given: only width env decl
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "W: "},
					&template.ExprPart{Expr: "w"},
				},
			},
		},
	}
	sc := &script.Script{
		EnvDecls: []script.EnvDecl{
			{Name: "w", Key: "width"},
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
	if !strings.Contains(src, "w, _ := term.GetSize") {
		t.Errorf("expected w, _ := term.GetSize in output:\n%s", src)
	}
}

func TestGenerateReactiveUsesSurgicalRendering(t *testing.T) {
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

	// Then — should use prevTree, DiffTrees, ApplyChanges
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "prevTree") {
		t.Errorf("expected prevTree variable in output:\n%s", src)
	}
	if !strings.Contains(src, "DiffTrees") {
		t.Errorf("expected layout.DiffTrees in output:\n%s", src)
	}
	if !strings.Contains(src, "ApplyChanges") {
		t.Errorf("expected layout.ApplyChanges in output:\n%s", src)
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
	assertValidGo(t, out)
	src := string(out)
	if !strings.Contains(src, "x := 0") {
		t.Errorf("expected x := 0 in output:\n%s", src)
	}
	if !strings.Contains(src, "y := 0") {
		t.Errorf("expected y := 0 in output:\n%s", src)
	}
}

func TestGenerateZeroArgHandlerGetsAutoQuit(t *testing.T) {
	// Given — a zero-arg handler should get auto-quit for backward compat
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
				Children:   []template.Node{textNode("Hello")},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "count", InitExpr: "0"},
		},
		FuncDecls: []script.FuncDecl{
			{Name: "handleKey", Params: "", Body: "\n\tcount = count + 1\n",
				StateAssignments: []script.StateAssignment{{VarName: "count", Line: "count = count + 1"}}},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertValidGo(t, out)
	src := string(out)
	// Should include auto-quit for Ctrl+C and signals
	if !strings.Contains(src, "app.Quit()") {
		t.Errorf("expected app.Quit() in auto-quit:\n%s", src)
	}
	if !strings.Contains(src, "evt.Rune == 3") {
		t.Errorf("expected Ctrl+C check in auto-quit:\n%s", src)
	}
	if !strings.Contains(src, "input.EventSignal") {
		t.Errorf("expected EventSignal check in auto-quit:\n%s", src)
	}
	// Zero-arg handler should be called without evt
	if !strings.Contains(src, "handleKey()") {
		t.Errorf("expected handleKey() call (not handleKey(evt)):\n%s", src)
	}
}

func TestGenerateEventAwareNoAutoQuit(t *testing.T) {
	// Given — event-aware handler should NOT get auto-quit
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
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
				Name:   "handleKey",
				Params: "evt input.Event",
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
	assertValidGo(t, out)
	src := string(out)
	// Should NOT have auto-quit
	if strings.Contains(src, "evt.Rune == 3") {
		t.Errorf("event-aware handler should not have auto Ctrl+C quit:\n%s", src)
	}
}

func TestGenerateEventAwareClosure(t *testing.T) {
	// Given — a function with params (event-aware handler)
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"onkey": "handleKey"},
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
				Name:   "handleKey",
				Params: "evt input.Event",
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
	assertValidGo(t, out)
	src := string(out)
	// Should emit parameterized closure
	if !strings.Contains(src, "func(evt input.Event)") {
		t.Errorf("expected parameterized closure in output:\n%s", src)
	}
	// Should call handler with evt
	if !strings.Contains(src, "handleKey(evt)") {
		t.Errorf("expected handleKey(evt) call in event handler:\n%s", src)
	}
}
