package codegen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/tomyan/sumi/parser/template"
)

// B4a: mixed content — loose text interleaved with inline elements
// compiles to interleaved KindText children.

func TestGenerateMixedContentChildren(t *testing.T) {
	// Given
	doc, err := template.Parse(`<p>hello <strong>bold</strong> tail</p>`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	src := string(out)
	for _, want := range []string{`Content: "hello "`, `Content: "bold"`, `Content: " tail"`} {
		if !strings.Contains(src, want) {
			t.Errorf("missing %s in output:\n%s", want, src)
		}
	}
	hello := strings.Index(src, `Content: "hello "`)
	bold := strings.Index(src, `Content: "bold"`)
	tail := strings.Index(src, `Content: " tail"`)
	if !(hello < bold && bold < tail) {
		t.Errorf("children out of source order: hello=%d bold=%d tail=%d", hello, bold, tail)
	}
	fset := token.NewFileSet()
	if _, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors); parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", src, parseErr)
	}
}

func TestGenerateMixedContentWithExpression(t *testing.T) {
	// Given: loose text with an expression next to an element.
	doc, err := template.Parse(`<p>count: {count} <em>items</em></p>`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// When
	out, err := generateStatic(doc, nil, nil, "main")

	// Then: expression text node extracted for sync, em emitted after it.
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	gen := string(out)
	if !strings.Contains(gen, `Content: "items"`) {
		t.Errorf("missing em content in output:\n%s", gen)
	}
	fset := token.NewFileSet()
	if _, parseErr := parser.ParseFile(fset, "generated.go", out, parser.AllErrors); parseErr != nil {
		t.Fatalf("generated code is not valid Go:\n%s\n\nerror: %v", gen, parseErr)
	}
}
