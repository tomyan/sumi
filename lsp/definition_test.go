package lsp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefinitionOnHandler(t *testing.T) {
	// Given: a handler reference in an onclick attribute
	text := "<script>\nfunc go() {}\n</script>\n<button onclick={go}>x</button>"

	// When: cursor sits on the handler name in onclick={go}
	loc := Definition(text, Position{Line: 3, Character: 17}, "file:///c.sumi")

	// Then: it resolves to the func declaration in the same file
	if loc == nil {
		t.Fatal("expected a definition, got nil")
	}
	if loc.URI != "file:///c.sumi" {
		t.Errorf("URI = %q, want same file", loc.URI)
	}
	if loc.Range.Start.Line != 1 {
		t.Errorf("func line = %d, want 1", loc.Range.Start.Line)
	}
}

func TestDefinitionOnComponentTag(t *testing.T) {
	// Given: a component and its sibling file
	dir := t.TempDir()
	widget := filepath.Join(dir, "widget.sumi")
	if err := os.WriteFile(widget, []byte("<div></div>"), 0o644); err != nil {
		t.Fatal(err)
	}
	uri := "file://" + filepath.Join(dir, "main.sumi")
	text := "<Widget />"

	// When: cursor sits on the component tag name
	loc := Definition(text, Position{Line: 0, Character: 3}, uri)

	// Then: it resolves to the component's file at its start
	if loc == nil {
		t.Fatal("expected a definition, got nil")
	}
	if loc.URI != "file://"+widget {
		t.Errorf("URI = %q, want %q", loc.URI, "file://"+widget)
	}
	if loc.Range.Start.Line != 0 || loc.Range.Start.Character != 0 {
		t.Errorf("range start = %+v, want 0:0", loc.Range.Start)
	}
}

func TestDefinitionElsewhereIsNil(t *testing.T) {
	// Given: a cursor in a plain text body
	text := "<div>hello</div>"

	// When
	loc := Definition(text, Position{Line: 0, Character: 7}, "file:///c.sumi")

	// Then
	if loc != nil {
		t.Errorf("expected nil, got %+v", loc)
	}
}
