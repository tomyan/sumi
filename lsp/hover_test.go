package lsp

import (
	"strings"
	"testing"
)

func TestHoverOnCSSProperty(t *testing.T) {
	// Given: a cursor on the "opacity" property in a style rule
	text := "<style>\n.a { opacity: 0.5 }\n</style>\n<div>x</div>"

	// When
	h := HoverAt(text, Position{Line: 1, Character: 8})

	// Then: markdown mentions the property and its sumi-specific note
	if h == nil {
		t.Fatal("expected hover, got nil")
	}
	if !strings.Contains(h.Contents.Value, "opacity") {
		t.Errorf("hover value = %q, want it to mention opacity", h.Contents.Value)
	}
	if !strings.Contains(strings.ToLower(h.Contents.Value), "dim") {
		t.Errorf("hover value = %q, want the dim note", h.Contents.Value)
	}
}

func TestHoverOnTagName(t *testing.T) {
	// Given: a cursor on a <button> tag name
	text := "<button>Go</button>"

	// When
	h := HoverAt(text, Position{Line: 0, Character: 3})

	// Then: a UA note mentioning activation
	if h == nil {
		t.Fatal("expected hover, got nil")
	}
	if !strings.Contains(strings.ToLower(h.Contents.Value), "activat") {
		t.Errorf("hover value = %q, want an activation note", h.Contents.Value)
	}
}

func TestHoverElsewhereIsNil(t *testing.T) {
	// Given: a cursor in a plain text body
	text := "<div>hello</div>"

	// When
	h := HoverAt(text, Position{Line: 0, Character: 7})

	// Then
	if h != nil {
		t.Errorf("expected nil hover in text body, got %+v", h)
	}
}
