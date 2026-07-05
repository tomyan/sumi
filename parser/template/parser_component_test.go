package template

import (
	"testing"
)

func TestParseComponentSelfClosing(t *testing.T) {
	// Given
	input := `<counter />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	comp, ok := doc.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("child is %T, want *ComponentElement", doc.Children[0])
	}
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
	if comp.nodeType() != "component" {
		t.Errorf("nodeType() = %q, want %q", comp.nodeType(), "component")
	}
}

func TestParseComponentWithAttributes(t *testing.T) {
	// Given
	input := `<counter label="Clicks" />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	comp := doc.Children[0].(*ComponentElement)
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
	if got := comp.Attributes["label"]; got != "Clicks" {
		t.Errorf("Attributes[\"label\"] = %q, want %q", got, "Clicks")
	}
}

func TestParseComponentMultipleAttributes(t *testing.T) {
	// Given
	input := `<counter label="Clicks" start="0" />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := doc.Children[0].(*ComponentElement)
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
	expected := map[string]string{
		"label": "Clicks",
		"start": "0",
	}
	for k, want := range expected {
		if got := comp.Attributes[k]; got != want {
			t.Errorf("Attributes[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestParseComponentInsideBox(t *testing.T) {
	// Given
	input := `<div><counter label="X" /></div>`

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
	comp, ok := box.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("box child is %T, want *ComponentElement", box.Children[0])
	}
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
	if got := comp.Attributes["label"]; got != "X" {
		t.Errorf("Attributes[\"label\"] = %q, want %q", got, "X")
	}
}

func TestParseMultipleComponentsInBox(t *testing.T) {
	// Given
	input := `<div><counter label="A" /><counter label="B" /></div>`

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
	comp1 := box.Children[0].(*ComponentElement)
	comp2 := box.Children[1].(*ComponentElement)
	if got := comp1.Attributes["label"]; got != "A" {
		t.Errorf("child 0 label = %q, want %q", got, "A")
	}
	if got := comp2.Attributes["label"]; got != "B" {
		t.Errorf("child 1 label = %q, want %q", got, "B")
	}
}

func TestParseComponentWithClosingTag(t *testing.T) {
	// Given
	input := `<counter></counter>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	comp, ok := doc.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("child is %T, want *ComponentElement", doc.Children[0])
	}
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
}

func TestParseComponentMixedWithText(t *testing.T) {
	// Given
	input := `<div><span>Title</span><counter /></div>`

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
	_, ok := box.Children[0].(*TextElement)
	if !ok {
		t.Fatalf("child 0 is %T, want *TextElement", box.Children[0])
	}
	comp, ok := box.Children[1].(*ComponentElement)
	if !ok {
		t.Fatalf("child 1 is %T, want *ComponentElement", box.Children[1])
	}
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
}

func TestParseNamespacedComponentSelfClosing(t *testing.T) {
	// Given
	input := `<sumi:TextInput />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	comp, ok := doc.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("child is %T, want *ComponentElement", doc.Children[0])
	}
	if comp.Name != "sumi:TextInput" {
		t.Errorf("Name = %q, want %q", comp.Name, "sumi:TextInput")
	}
}

func TestParseNamespacedComponentWithClosingTag(t *testing.T) {
	// Given
	input := `<sumi:TextInput></sumi:TextInput>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp, ok := doc.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("child is %T, want *ComponentElement", doc.Children[0])
	}
	if comp.Name != "sumi:TextInput" {
		t.Errorf("Name = %q, want %q", comp.Name, "sumi:TextInput")
	}
}

func TestParseNamespacedComponentWithAttributes(t *testing.T) {
	// Given
	input := `<sumi:TextInput bind:value={name} placeholder="Enter name" />`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comp := doc.Children[0].(*ComponentElement)
	if comp.Name != "sumi:TextInput" {
		t.Errorf("Name = %q, want %q", comp.Name, "sumi:TextInput")
	}
	if got := comp.Attributes["bind:value"]; got != "{name}" {
		t.Errorf("bind:value = %q, want %q", got, "{name}")
	}
	if got := comp.Attributes["placeholder"]; got != "Enter name" {
		t.Errorf("placeholder = %q, want %q", got, "Enter name")
	}
}

func TestParseNamespacedComponentInsideBox(t *testing.T) {
	// Given
	input := `<div><sumi:TextInput bind:value={x} /></div>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	comp, ok := box.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("box child is %T, want *ComponentElement", box.Children[0])
	}
	if comp.Name != "sumi:TextInput" {
		t.Errorf("Name = %q, want %q", comp.Name, "sumi:TextInput")
	}
}

func TestParseSelfClosingNoSpace(t *testing.T) {
	// Given
	input := `<counter/>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(doc.Children))
	}
	comp, ok := doc.Children[0].(*ComponentElement)
	if !ok {
		t.Fatalf("child is %T, want *ComponentElement", doc.Children[0])
	}
	if comp.Name != "counter" {
		t.Errorf("Name = %q, want %q", comp.Name, "counter")
	}
}
