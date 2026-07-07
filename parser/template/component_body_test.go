package template

import "testing"

func TestParseComponentWithTextBody(t *testing.T) {
	// Given — a component tag wrapping body content
	input := `<Card>Hello</Card>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp, ok := doc.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *ComponentElement", doc.Children[0])
	}
	if comp.Name != "Card" {
		t.Errorf("Name = %q, want %q", comp.Name, "Card")
	}
	if len(comp.Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(comp.Children))
	}
	if _, ok := comp.Children[0].(*TextElement); !ok {
		t.Errorf("Children[0] type = %T, want *TextElement", comp.Children[0])
	}
}

func TestParseComponentWithSnippetAndBody(t *testing.T) {
	// Given — a component tag with a snippet block and remaining body content
	input := `<Card>{snippet footer()}<span>F</span>{/snippet}<span>Body</span></Card>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := doc.Children[0].(*ComponentElement)
	if len(comp.Children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(comp.Children))
	}
	snip, ok := comp.Children[0].(*SnippetNode)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *SnippetNode", comp.Children[0])
	}
	if snip.Name != "footer" {
		t.Errorf("snippet Name = %q, want %q", snip.Name, "footer")
	}
	if _, ok := comp.Children[1].(*BoxElement); !ok {
		if _, ok2 := comp.Children[1].(*TextElement); !ok2 {
			t.Errorf("Children[1] type = %T, want *TextElement or *BoxElement", comp.Children[1])
		}
	}
}

func TestParseSelfClosingComponentHasNoChildren(t *testing.T) {
	// Given — a self-closing component
	input := `<Card />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := doc.Children[0].(*ComponentElement)
	if len(comp.Children) != 0 {
		t.Errorf("len(Children) = %d, want 0", len(comp.Children))
	}
}
