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

func TestParseBoxWithSingleTextChild(t *testing.T) {
	doc, err := Parse(`<box><text>Hello</text></box>`)
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
	if te.Content != "Hello" {
		t.Errorf("Content = %q, want %q", te.Content, "Hello")
	}
}

func TestParseBoxWithDirectionAndTwoChildren(t *testing.T) {
	doc, err := Parse(`<box direction="column"><text>A</text><text>B</text></box>`)
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
	if te1.Content != "A" {
		t.Errorf("child 0 Content = %q, want %q", te1.Content, "A")
	}
	if te2.Content != "B" {
		t.Errorf("child 1 Content = %q, want %q", te2.Content, "B")
	}
}

func TestParseBoxWithMultipleAttributes(t *testing.T) {
	doc, err := Parse(`<box width="40" height="10" border="single" padding="1 2"><text>X</text></box>`)
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
	doc, err := Parse(`<box><box><text>Deep</text></box></box>`)
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
	if te.Content != "Deep" {
		t.Errorf("Content = %q, want %q", te.Content, "Deep")
	}
}

func TestParseEmptyBox(t *testing.T) {
	doc, err := Parse(`<box></box>`)
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
	_, err := Parse(`<box><text>Hello</text>`)
	if err == nil {
		t.Fatal("expected error for missing </box> closing tag, got nil")
	}
}

func TestParseBoxWithWhitespaceBetweenChildren(t *testing.T) {
	doc, err := Parse(`<box>
  <text>A</text>
  <text>B</text>
</box>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if len(box.Children) != 2 {
		t.Fatalf("box has %d children, want 2", len(box.Children))
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
