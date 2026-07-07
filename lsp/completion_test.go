package lsp

import (
	"os"
	"path/filepath"
	"testing"
)

// labels extracts the label set from completion items.
func labels(items []CompletionItem) map[string]int {
	m := map[string]int{}
	for _, it := range items {
		m[it.Label] = it.Kind
	}
	return m
}

func TestCompletionCSSProperties(t *testing.T) {
	// Given: a cursor in the style section
	text := "<style>\n.a { c }\n</style>\n<div>x</div>"

	// When
	items := Completions(text, Position{Line: 1, Character: 6}, "file:///x.sumi")

	// Then: CSS property names appear as Property-kind items
	got := labels(items)
	if got["color"] != KindProperty {
		t.Errorf("color kind = %d, want %d", got["color"], KindProperty)
	}
	if _, ok := got["border-title"]; !ok {
		t.Error("expected border-title among CSS completions")
	}
}

func TestCompletionTagNames(t *testing.T) {
	// Given: a cursor right after `<`
	text := "<d"

	// When
	items := Completions(text, Position{Line: 0, Character: 2}, "file:///x.sumi")

	// Then: HTML tags appear as Keyword-kind items
	got := labels(items)
	if got["div"] != KindKeyword {
		t.Errorf("div kind = %d, want %d", got["div"], KindKeyword)
	}
}

func TestCompletionAttrNames(t *testing.T) {
	// Given: a cursor after the tag name and a space
	text := "<div ></div>"

	// When
	items := Completions(text, Position{Line: 0, Character: 5}, "file:///x.sumi")

	// Then: global attributes appear as Field-kind items
	got := labels(items)
	if got["class"] != KindField {
		t.Errorf("class kind = %d, want %d", got["class"], KindField)
	}
	if _, ok := got["bind:value"]; !ok {
		t.Error("expected bind:value among attr completions")
	}
}

func TestCompletionSiblingComponents(t *testing.T) {
	// Given: a directory with a sibling component file
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "my-widget.sumi"), []byte("<div></div>"), 0o644); err != nil {
		t.Fatal(err)
	}
	uri := "file://" + filepath.Join(dir, "main.sumi")
	text := "<M"

	// When
	items := Completions(text, Position{Line: 0, Character: 2}, uri)

	// Then: the sibling appears as a Class-kind PascalCase component
	got := labels(items)
	if got["Mywidget"] != KindClass {
		t.Errorf("component kind = %d, want %d (%v)", got["Mywidget"], KindClass, got)
	}
}
