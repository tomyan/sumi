package template

import (
	"testing"
)

// --- Expression tests ---

func TestParseTextWithExpressionOnly(t *testing.T) {
	// When
	doc, err := Parse(`<text>{count}</text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	te := doc.Children[0].(*TextElement)
	assertParts(t, te, []Part{&ExprPart{Expr: "count"}})
}

func TestParseTextWithStringAndExpression(t *testing.T) {
	// When
	doc, err := Parse(`<text>Count: {count}</text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	te := doc.Children[0].(*TextElement)
	assertParts(t, te, []Part{
		&StringPart{Value: "Count: "},
		&ExprPart{Expr: "count"},
	})
}

func TestParseTextWithTwoExpressions(t *testing.T) {
	// When
	doc, err := Parse(`<text>{a} and {b}</text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	te := doc.Children[0].(*TextElement)
	assertParts(t, te, []Part{
		&ExprPart{Expr: "a"},
		&StringPart{Value: " and "},
		&ExprPart{Expr: "b"},
	})
}

func TestParseTextWithExpressionContainingSpaces(t *testing.T) {
	// When
	doc, err := Parse(`<text>Count: {count + 1}!</text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	te := doc.Children[0].(*TextElement)
	assertParts(t, te, []Part{
		&StringPart{Value: "Count: "},
		&ExprPart{Expr: "count + 1"},
		&StringPart{Value: "!"},
	})
}

// --- Text element attribute tests ---

func TestParseTextWithClassAttribute(t *testing.T) {
	// When
	doc, err := Parse(`<text class="title">Hello</text>`)

	// Then
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
	if te.Attributes == nil {
		t.Fatal("Attributes should not be nil")
	}
	if got := te.Attributes["class"]; got != "title" {
		t.Errorf("Attributes[\"class\"] = %q, want %q", got, "title")
	}
	assertParts(t, te, []Part{&StringPart{Value: "Hello"}})
}

func TestParseTextWithoutAttributesBackwardCompat(t *testing.T) {
	// When
	doc, err := Parse(`<text>Hello</text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	te := doc.Children[0].(*TextElement)
	// Attributes should be nil or empty when no attributes specified
	if len(te.Attributes) != 0 {
		t.Errorf("got %d attributes, want 0", len(te.Attributes))
	}
	assertParts(t, te, []Part{&StringPart{Value: "Hello"}})
}

func TestParseTextWithMultipleAttributes(t *testing.T) {
	// When
	doc, err := Parse(`<text class="title" id="heading">Content</text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	te := doc.Children[0].(*TextElement)
	if te.Attributes == nil {
		t.Fatal("Attributes should not be nil")
	}
	expected := map[string]string{
		"class": "title",
		"id":    "heading",
	}
	for k, want := range expected {
		if got := te.Attributes[k]; got != want {
			t.Errorf("Attributes[%q] = %q, want %q", k, got, want)
		}
	}
	assertParts(t, te, []Part{&StringPart{Value: "Content"}})
}

func TestParseTextWithClassInsideBox(t *testing.T) {
	// When
	doc, err := Parse(`<box><text class="label">Name</text></box>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	te := box.Children[0].(*TextElement)
	if got := te.Attributes["class"]; got != "label" {
		t.Errorf("Attributes[\"class\"] = %q, want %q", got, "label")
	}
	assertParts(t, te, []Part{&StringPart{Value: "Name"}})
}
