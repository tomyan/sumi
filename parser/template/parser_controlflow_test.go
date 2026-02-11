package template

import (
	"testing"
)

func TestParseIfBasic(t *testing.T) {
	// Given
	input := `{if x}<text>Yes</text>{/if}`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	ifNode, ok := doc.Children[0].(*IfNode)
	if !ok {
		t.Fatalf("child is %T, want *IfNode", doc.Children[0])
	}
	if ifNode.Condition != "x" {
		t.Errorf("Condition = %q, want %q", ifNode.Condition, "x")
	}
	if len(ifNode.Then) != 1 {
		t.Fatalf("Then has %d children, want 1", len(ifNode.Then))
	}
	te, ok := ifNode.Then[0].(*TextElement)
	if !ok {
		t.Fatalf("Then[0] is %T, want *TextElement", ifNode.Then[0])
	}
	assertParts(t, te, []Part{&StringPart{Value: "Yes"}})
	if ifNode.Else != nil {
		t.Errorf("Else should be nil, got %d children", len(ifNode.Else))
	}
}

func TestParseIfElse(t *testing.T) {
	// Given
	input := `{if count > 0}<text>Has items</text>{else}<text>Empty</text>{/if}`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifNode := doc.Children[0].(*IfNode)
	if ifNode.Condition != "count > 0" {
		t.Errorf("Condition = %q, want %q", ifNode.Condition, "count > 0")
	}
	if len(ifNode.Then) != 1 {
		t.Fatalf("Then has %d children, want 1", len(ifNode.Then))
	}
	thenText := ifNode.Then[0].(*TextElement)
	assertParts(t, thenText, []Part{&StringPart{Value: "Has items"}})
	if len(ifNode.Else) != 1 {
		t.Fatalf("Else has %d children, want 1", len(ifNode.Else))
	}
	elseText := ifNode.Else[0].(*TextElement)
	assertParts(t, elseText, []Part{&StringPart{Value: "Empty"}})
}

func TestParseIfAtRoot(t *testing.T) {
	// Given
	input := `<text>Before</text>{if visible}<text>Shown</text>{/if}<text>After</text>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 3 {
		t.Fatalf("got %d children, want 3", len(doc.Children))
	}
	_, ok := doc.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("child[0] is %T, want *TextElement", doc.Children[0])
	}
	_, ok = doc.Children[1].(*IfNode)
	if !ok {
		t.Fatalf("child[1] is %T, want *IfNode", doc.Children[1])
	}
	_, ok = doc.Children[2].(*TextElement)
	if !ok {
		t.Fatalf("child[2] is %T, want *TextElement", doc.Children[2])
	}
}

func TestParseIfInBox(t *testing.T) {
	// Given
	input := `<box>{if x}<text>Inside</text>{/if}</box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if len(box.Children) != 1 {
		t.Fatalf("box has %d children, want 1", len(box.Children))
	}
	ifNode, ok := box.Children[0].(*IfNode)
	if !ok {
		t.Fatalf("box child is %T, want *IfNode", box.Children[0])
	}
	if ifNode.Condition != "x" {
		t.Errorf("Condition = %q, want %q", ifNode.Condition, "x")
	}
}

func TestParseIfMultipleChildren(t *testing.T) {
	// Given
	input := `{if x}<text>A</text><text>B</text><box><text>C</text></box>{/if}`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ifNode := doc.Children[0].(*IfNode)
	if len(ifNode.Then) != 3 {
		t.Fatalf("Then has %d children, want 3", len(ifNode.Then))
	}
}

func TestParseIfNestedInIf(t *testing.T) {
	// Given
	input := `{if x}{if y}<text>Both</text>{/if}{/if}`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outer := doc.Children[0].(*IfNode)
	if outer.Condition != "x" {
		t.Errorf("outer Condition = %q, want %q", outer.Condition, "x")
	}
	if len(outer.Then) != 1 {
		t.Fatalf("outer Then has %d children, want 1", len(outer.Then))
	}
	inner, ok := outer.Then[0].(*IfNode)
	if !ok {
		t.Fatalf("inner is %T, want *IfNode", outer.Then[0])
	}
	if inner.Condition != "y" {
		t.Errorf("inner Condition = %q, want %q", inner.Condition, "y")
	}
}

func TestParseIfMissingClose(t *testing.T) {
	// Given
	input := `{if x}<text>Hello</text>`

	// When
	_, err := Parse(input)

	// Then
	if err == nil {
		t.Fatal("expected error for missing {/if}, got nil")
	}
}
