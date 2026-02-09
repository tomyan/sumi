package template

import (
	"testing"
)

func TestParseSingleTextElement(t *testing.T) {
	doc, err := Parse(`<text>Hello</text>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	te, ok := doc.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("child is %T, want *TextElement", doc.Children[0])
	}
	if te.Content != "Hello" {
		t.Errorf("Content = %q, want %q", te.Content, "Hello")
	}
}

func TestParseTwoTextElements(t *testing.T) {
	doc, err := Parse(`<text>Hello</text><text>World</text>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(doc.Children))
	}
	te1, ok := doc.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("child 0 is %T, want *TextElement", doc.Children[0])
	}
	if te1.Content != "Hello" {
		t.Errorf("child 0 Content = %q, want %q", te1.Content, "Hello")
	}
	te2, ok := doc.Children[1].(*TextElement)
	if !ok {
		t.Fatalf("child 1 is %T, want *TextElement", doc.Children[1])
	}
	if te2.Content != "World" {
		t.Errorf("child 1 Content = %q, want %q", te2.Content, "World")
	}
}

func TestParseEmptyTextElement(t *testing.T) {
	doc, err := Parse(`<text></text>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	te, ok := doc.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("child is %T, want *TextElement", doc.Children[0])
	}
	if te.Content != "" {
		t.Errorf("Content = %q, want empty", te.Content)
	}
}

func TestParseWhitespaceBetweenElementsIgnored(t *testing.T) {
	doc, err := Parse("  <text>A</text>  \n  <text>B</text>  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(doc.Children))
	}
	te1 := doc.Children[0].(*TextElement)
	te2 := doc.Children[1].(*TextElement)
	if te1.Content != "A" {
		t.Errorf("child 0 Content = %q, want %q", te1.Content, "A")
	}
	if te2.Content != "B" {
		t.Errorf("child 1 Content = %q, want %q", te2.Content, "B")
	}
}

func TestParseMissingClosingTagReturnsError(t *testing.T) {
	_, err := Parse(`<text>Hello`)
	if err == nil {
		t.Fatal("expected error for missing closing tag, got nil")
	}
}

func TestParseWhitespaceInsideTextPreserved(t *testing.T) {
	doc, err := Parse(`<text>  Hello  </text>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	te := doc.Children[0].(*TextElement)
	if te.Content != "  Hello  " {
		t.Errorf("Content = %q, want %q", te.Content, "  Hello  ")
	}
}
