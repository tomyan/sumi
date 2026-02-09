package template

import (
	"testing"
)

// assertParts is a test helper that checks a TextElement's Parts match expectations.
func assertParts(t *testing.T, te *TextElement, expected []Part) {
	t.Helper()
	if len(te.Parts) != len(expected) {
		t.Fatalf("got %d parts, want %d", len(te.Parts), len(expected))
	}
	for i, want := range expected {
		got := te.Parts[i]
		switch w := want.(type) {
		case *StringPart:
			g, ok := got.(*StringPart)
			if !ok {
				t.Fatalf("part[%d] is %T, want *StringPart", i, got)
			}
			if g.Value != w.Value {
				t.Errorf("part[%d] Value = %q, want %q", i, g.Value, w.Value)
			}
		case *ExprPart:
			g, ok := got.(*ExprPart)
			if !ok {
				t.Fatalf("part[%d] is %T, want *ExprPart", i, got)
			}
			if g.Expr != w.Expr {
				t.Errorf("part[%d] Expr = %q, want %q", i, g.Expr, w.Expr)
			}
		default:
			t.Fatalf("unexpected part type %T at index %d", want, i)
		}
	}
}

func TestParseSingleTextElement(t *testing.T) {
	// When
	doc, err := Parse(`<text>Hello</text>`)

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
	assertParts(t, te, []Part{&StringPart{Value: "Hello"}})
}

func TestParseTwoTextElements(t *testing.T) {
	// When
	doc, err := Parse(`<text>Hello</text><text>World</text>`)

	// Then
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
	assertParts(t, te1, []Part{&StringPart{Value: "Hello"}})
	te2, ok := doc.Children[1].(*TextElement)
	if !ok {
		t.Fatalf("child 1 is %T, want *TextElement", doc.Children[1])
	}
	assertParts(t, te2, []Part{&StringPart{Value: "World"}})
}

func TestParseEmptyTextElement(t *testing.T) {
	// When
	doc, err := Parse(`<text></text>`)

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
	if len(te.Parts) != 0 {
		t.Errorf("got %d parts, want 0", len(te.Parts))
	}
}

func TestParseWhitespaceBetweenElementsIgnored(t *testing.T) {
	// When
	doc, err := Parse("  <text>A</text>  \n  <text>B</text>  ")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(doc.Children))
	}
	te1 := doc.Children[0].(*TextElement)
	te2 := doc.Children[1].(*TextElement)
	assertParts(t, te1, []Part{&StringPart{Value: "A"}})
	assertParts(t, te2, []Part{&StringPart{Value: "B"}})
}

func TestParseMissingClosingTagReturnsError(t *testing.T) {
	// When
	_, err := Parse(`<text>Hello`)

	// Then
	if err == nil {
		t.Fatal("expected error for missing closing tag, got nil")
	}
}

func TestParseBoxWithSingleTextChild(t *testing.T) {
	// When
	doc, err := Parse(`<box><text>Hello</text></box>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	box, ok := doc.Children[0].(*BoxElement)
	if !ok {
		t.Fatalf("child is %T, want *BoxElement", doc.Children[0])
	}
	if box.nodeType() != "box" {
		t.Errorf("nodeType() = %q, want %q", box.nodeType(), "box")
	}
	if len(box.Children) != 1 {
		t.Fatalf("box has %d children, want 1", len(box.Children))
	}
	te, ok := box.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("box child is %T, want *TextElement", box.Children[0])
	}
	assertParts(t, te, []Part{&StringPart{Value: "Hello"}})
}

func TestParseBoxWithDirectionAndTwoChildren(t *testing.T) {
	// When
	doc, err := Parse(`<box direction="column"><text>A</text><text>B</text></box>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	box := doc.Children[0].(*BoxElement)
	if box.Attributes["direction"] != "column" {
		t.Errorf("direction = %q, want %q", box.Attributes["direction"], "column")
	}
	if len(box.Children) != 2 {
		t.Fatalf("box has %d children, want 2", len(box.Children))
	}
	te1 := box.Children[0].(*TextElement)
	te2 := box.Children[1].(*TextElement)
	assertParts(t, te1, []Part{&StringPart{Value: "A"}})
	assertParts(t, te2, []Part{&StringPart{Value: "B"}})
}

func TestParseBoxWithMultipleAttributes(t *testing.T) {
	// When
	doc, err := Parse(`<box width="40" height="10" border="single" padding="1 2"><text>X</text></box>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	expected := map[string]string{
		"width":   "40",
		"height":  "10",
		"border":  "single",
		"padding": "1 2",
	}
	for k, want := range expected {
		if got := box.Attributes[k]; got != want {
			t.Errorf("Attributes[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestParseNestedBoxes(t *testing.T) {
	// When
	doc, err := Parse(`<box><box><text>Deep</text></box></box>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outer := doc.Children[0].(*BoxElement)
	if len(outer.Children) != 1 {
		t.Fatalf("outer box has %d children, want 1", len(outer.Children))
	}
	inner, ok := outer.Children[0].(*BoxElement)
	if !ok {
		t.Fatalf("outer child is %T, want *BoxElement", outer.Children[0])
	}
	if len(inner.Children) != 1 {
		t.Fatalf("inner box has %d children, want 1", len(inner.Children))
	}
	te := inner.Children[0].(*TextElement)
	assertParts(t, te, []Part{&StringPart{Value: "Deep"}})
}

func TestParseEmptyBox(t *testing.T) {
	// When
	doc, err := Parse(`<box></box>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if len(box.Children) != 0 {
		t.Fatalf("box has %d children, want 0", len(box.Children))
	}
	if box.Attributes == nil {
		t.Fatal("Attributes should not be nil")
	}
}

func TestParseBoxMissingClosingTag(t *testing.T) {
	// When
	_, err := Parse(`<box><text>Hello</text>`)

	// Then
	if err == nil {
		t.Fatal("expected error for missing </box> closing tag, got nil")
	}
}

func TestParseBoxWithWhitespaceBetweenChildren(t *testing.T) {
	// Given
	input := `<box>
  <text>A</text>
  <text>B</text>
</box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if len(box.Children) != 2 {
		t.Fatalf("box has %d children, want 2", len(box.Children))
	}
}

func TestParseWhitespaceInsideTextPreserved(t *testing.T) {
	// When
	doc, err := Parse(`<text>  Hello  </text>`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	te := doc.Children[0].(*TextElement)
	assertParts(t, te, []Part{&StringPart{Value: "  Hello  "}})
}

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
