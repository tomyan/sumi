package template

import "testing"

func TestParseSnippetDefinition(t *testing.T) {
	// Given
	input := `<box>{snippet item(name string)}<text>{name}</text>{/snippet}</box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	snippet, ok := box.Children[0].(*SnippetNode)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *SnippetNode", box.Children[0])
	}
	if snippet.Name != "item" {
		t.Errorf("Name = %q, want %q", snippet.Name, "item")
	}
	if snippet.Params != "(name string)" {
		t.Errorf("Params = %q, want %q", snippet.Params, "(name string)")
	}
	if len(snippet.Children) != 1 {
		t.Fatalf("len(Children) = %d, want 1", len(snippet.Children))
	}
}

func TestParseRenderSnippet(t *testing.T) {
	// Given
	input := `<box>{render item("hello")}</box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	render, ok := box.Children[0].(*RenderNode)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *RenderNode", box.Children[0])
	}
	if render.Name != "item" {
		t.Errorf("Name = %q, want %q", render.Name, "item")
	}
	if render.Args != `"hello"` {
		t.Errorf("Args = %q, want %q", render.Args, `"hello"`)
	}
}
