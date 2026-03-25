package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/template"
)

func TestSignalDeclGeneratesSignalNew(t *testing.T) {
	// Given — a signal declaration instead of $state
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
		SignalDecls: []script.SignalDecl{
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

	// Should declare signal
	if !strings.Contains(src, "signal.New(0)") {
		t.Errorf("expected signal.New(0) in output:\n%s", src)
	}

	// Template expression should auto-unwrap with .Get()
	if !strings.Contains(src, "count.Get()") {
		t.Errorf("expected count.Get() in template expression:\n%s", src)
	}

	// Should NOT have dirty flag
	if strings.Contains(src, "dirty :=") || strings.Contains(src, "dirty = true") {
		t.Errorf("should not have dirty flag with signals:\n%s", src)
	}

	// Should import signal package
	if !strings.Contains(src, `"github.com/tomyan/sumi/runtime/signal"`) {
		t.Errorf("expected signal import:\n%s", src)
	}
}

func TestSignalDeclTemplateAutoUnwrapsInExpression(t *testing.T) {
	// Given — expression using signal variable in arithmetic
	doc := &template.Document{
		Children: []template.Node{
			&template.TextElement{
				Parts: []template.Part{
					&template.StringPart{Value: "Double: "},
					&template.ExprPart{Expr: "count * 2"},
				},
			},
		},
	}
	sc := &script.Script{
		SignalDecls: []script.SignalDecl{
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

	// Should unwrap count to count.Get() in the expression
	// go/format may compact: count.Get()*2
	if !strings.Contains(src, "count.Get()") {
		t.Errorf("expected count.Get() in template expression:\n%s", src)
	}
	if !strings.Contains(src, "count.Get()*2") && !strings.Contains(src, "count.Get() * 2") {
		t.Errorf("expected count.Get()*2 in template expression:\n%s", src)
	}
}
