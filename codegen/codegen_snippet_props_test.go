package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func TestSnippetPropNilDefaulted(t *testing.T) {
	// Given — a component with a snippet prop rendered in its template
	src := generateFromTemplate(t, `<div>{render footer()}</div>`, `var footer func() []*sumi.Input`)

	// Then — the props struct carries the func-typed field
	if !strings.Contains(src, "Footer func() []*sumi.Input") {
		t.Errorf("expected Footer prop field:\n%s", src)
	}
	// And an unpassed prop is nil-defaulted so {render} renders nothing
	if !strings.Contains(src, "if footer == nil {") {
		t.Errorf("expected nil-default guard for snippet prop:\n%s", src)
	}
	if !strings.Contains(src, "cs = append(cs, footer()...)") {
		t.Errorf("expected render append for prop:\n%s", src)
	}
}

func TestRenderUnknownNameIsCompileError(t *testing.T) {
	// Given — a {render} naming neither a local snippet nor a snippet prop
	doc, err := template.Parse(`<div>{render bogus()}</div>`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// When
	_, err = GenerateComponent(doc, ``, nil, ComponentOptions{
		PackageName:   "card",
		ComponentName: "Card",
	})

	// Then
	if err == nil {
		t.Fatalf("expected error for unknown render name, got nil")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error = %q, want it to name the unknown snippet", err.Error())
	}
}

func TestConsumerBodyBecomesChildrenProp(t *testing.T) {
	// Given — a parent mounting a child component with body content
	src := generateFromTemplate(t, `<Card>Body</Card>`, ``)

	// Then — the body is passed as the implicit Children snippet prop
	if !strings.Contains(src, "NewCard(CardProps{") {
		t.Errorf("expected NewCard construction:\n%s", src)
	}
	if !strings.Contains(src, "Children: func() []*sumi.Input {") {
		t.Errorf("expected Children snippet prop:\n%s", src)
	}
}

func TestConsumerSnippetBecomesNamedProp(t *testing.T) {
	// Given — a parent mounting a child with a named {snippet} in its body
	src := generateFromTemplate(t, `<Card>{snippet footer()}<span>F</span>{/snippet}</Card>`, ``)

	// Then — the snippet is passed as the matching named prop
	if !strings.Contains(src, "Footer: func() []*sumi.Input {") {
		t.Errorf("expected Footer snippet prop:\n%s", src)
	}
}
