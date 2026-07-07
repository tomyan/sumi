package codegen

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

func generateFromTemplate(t *testing.T, tmpl, scriptSrc string) string {
	t.Helper()
	doc, err := template.Parse(tmpl)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	out, err := GenerateComponent(doc, scriptSrc, nil, ComponentOptions{
		PackageName:   "card",
		ComponentName: "Card",
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	assertValidGo(t, out)
	return string(out)
}

func TestGenerateLocalSnippetClosure(t *testing.T) {
	// Given — a component declaring a local snippet and rendering it
	tmpl := `{snippet greeting(name string)}<span>Hi {name}</span>{/snippet}<div>{render greeting("Tom")}</div>`

	// When
	src := generateFromTemplate(t, tmpl, ``)

	// Then — the snippet becomes a local closure
	if !strings.Contains(src, "greeting := func(name string) []*sumi.Input {") {
		t.Errorf("expected greeting closure:\n%s", src)
	}
	// And the render call spreads its result into the parent's children
	if !strings.Contains(src, `greeting("Tom")...`) {
		t.Errorf("expected render append call:\n%s", src)
	}
}

func TestSnippetDeclarationDoesNotRenderInline(t *testing.T) {
	// Given — a snippet declared but never rendered
	tmpl := `{snippet unused()}<span>X</span>{/snippet}<div>Body</div>`

	// When
	src := generateFromTemplate(t, tmpl, ``)

	// Then — the closure exists but "X" is not emitted as tree content
	if !strings.Contains(src, "unused := func() []*sumi.Input {") {
		t.Errorf("expected unused closure:\n%s", src)
	}
}

func TestRenderMakesChildrenDynamic(t *testing.T) {
	// Given — a render call is the only dynamic-looking child
	tmpl := `{snippet body()}<span>content</span>{/snippet}<div>{render body()}</div>`

	// When
	src := generateFromTemplate(t, tmpl, ``)

	// Then — the parent div builds children via the append IIFE
	if !strings.Contains(src, "cs = append(cs, body()...)") {
		t.Errorf("expected dynamic append for render:\n%s", src)
	}
}
