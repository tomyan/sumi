package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestStaticSyncReturnsChangedInputs(t *testing.T) {
	// Given — static document (no control flow, no scroll)
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

	// Sync should return []*layout.Input
	if !strings.Contains(src, "sync := func() []*layout.Input {") {
		t.Errorf("expected sync to return []*layout.Input:\n%s", src)
	}

	// Should compare before assigning
	if !strings.Contains(src, "!= node0.Content") {
		t.Errorf("expected compare-before-assign in sync:\n%s", src)
	}

	// Should return changed slice
	if !strings.Contains(src, "return changed") {
		t.Errorf("expected return changed in sync:\n%s", src)
	}

	// Generated code should be valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestStaticDoRenderHasNoOpSkip(t *testing.T) {
	// Given — static document
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

	// doRender should have early return when nothing changed
	if !strings.Contains(src, "changed := sync()") {
		t.Errorf("expected changed := sync() in doRender:\n%s", src)
	}
	if !strings.Contains(src, "len(changed) == 0") {
		t.Errorf("expected no-op skip with len(changed) == 0:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestStaticDoRenderHasDirectWriteFastPath(t *testing.T) {
	// Given — static document
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

	// Should declare nodeBoxMap
	if !strings.Contains(src, "nodeBoxMap") {
		t.Errorf("expected nodeBoxMap variable:\n%s", src)
	}

	// Should have MapInputToBox call
	if !strings.Contains(src, "layout.MapInputToBox") {
		t.Errorf("expected layout.MapInputToBox call:\n%s", src)
	}

	// Should have DirectWriteText call
	if !strings.Contains(src, "layout.DirectWriteText") {
		t.Errorf("expected layout.DirectWriteText call:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestDynamicDoRenderHasNoDirectWrite(t *testing.T) {
	// Given — dynamic document
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children: []template.Node{
					textNode("Title"),
					&template.IfNode{
						Condition: "showModal",
						Then:      []template.Node{textNode("Modal content")},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "showModal", InitExpr: "false"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Should NOT have direct-write fast path
	if strings.Contains(src, "nodeBoxMap") {
		t.Errorf("dynamic document should not have nodeBoxMap:\n%s", src)
	}
	if strings.Contains(src, "DirectWriteText") {
		t.Errorf("dynamic document should not have DirectWriteText:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}

func TestDynamicSyncIsVoid(t *testing.T) {
	// Given — dynamic document (has {if} control flow)
	doc := &template.Document{
		Children: []template.Node{
			&template.BoxElement{
				Attributes: map[string]string{"border": "single"},
				Children: []template.Node{
					textNode("Title"),
					&template.IfNode{
						Condition: "showModal",
						Then:      []template.Node{textNode("Modal content")},
					},
				},
			},
		},
	}
	sc := &script.Script{
		StateDecls: []script.StateDecl{
			{Name: "showModal", InitExpr: "false"},
		},
	}

	// When
	out, err := Generate(doc, sc, nil, Options{PackageName: "main"})

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	src := string(out)

	// Sync should NOT return anything (void pattern)
	if strings.Contains(src, "sync := func() []*layout.Input {") {
		t.Errorf("dynamic document should not have returning sync:\n%s", src)
	}
	// Should use void sync
	if !strings.Contains(src, "sync := func() {") {
		t.Errorf("expected void sync for dynamic document:\n%s", src)
	}
	// Should NOT have no-op skip
	if strings.Contains(src, "len(changed) == 0") {
		t.Errorf("dynamic document should not have no-op skip:\n%s", src)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", string(out), parseErr)
	}
}
