package template

import (
	"strings"
	"testing"
)

func TestParseTitleElement(t *testing.T) {
	// Given
	input := `<title>My App</title>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(doc.Children))
	}
	title, ok := doc.Children[0].(*TitleElement)
	if !ok {
		t.Fatalf("expected TitleElement, got %T", doc.Children[0])
	}
	if len(title.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(title.Parts))
	}
	sp, ok := title.Parts[0].(*StringPart)
	if !ok {
		t.Fatalf("expected StringPart, got %T", title.Parts[0])
	}
	if sp.Value != "My App" {
		t.Errorf("expected %q, got %q", "My App", sp.Value)
	}
}

func TestParseTitleWithExpression(t *testing.T) {
	// Given
	input := `<title>{count} items</title>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	title, ok := doc.Children[0].(*TitleElement)
	if !ok {
		t.Fatalf("expected TitleElement, got %T", doc.Children[0])
	}
	if len(title.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(title.Parts))
	}
	ep, ok := title.Parts[0].(*ExprPart)
	if !ok {
		t.Fatalf("expected ExprPart, got %T", title.Parts[0])
	}
	if ep.Expr != "count" {
		t.Errorf("expected expr %q, got %q", "count", ep.Expr)
	}
	sp, ok := title.Parts[1].(*StringPart)
	if !ok {
		t.Fatalf("expected StringPart, got %T", title.Parts[1])
	}
	if sp.Value != " items" {
		t.Errorf("expected %q, got %q", " items", sp.Value)
	}
}

func TestParseTitleAndBox(t *testing.T) {
	// Given
	input := `<title>My App</title>
<box><text>Hello</text></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(doc.Children))
	}
	if _, ok := doc.Children[0].(*TitleElement); !ok {
		t.Errorf("expected TitleElement first, got %T", doc.Children[0])
	}
	if _, ok := doc.Children[1].(*BoxElement); !ok {
		t.Errorf("expected BoxElement second, got %T", doc.Children[1])
	}
}

func TestParseMultipleTitlesError(t *testing.T) {
	// Given
	input := `<title>First</title>
<title>Second</title>`

	// When
	_, err := Parse(input)

	// Then
	if err == nil {
		t.Fatal("expected error for multiple <title> elements")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("expected error mentioning 'title', got: %v", err)
	}
}
