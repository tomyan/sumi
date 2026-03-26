package template

import "testing"

func TestParseSlotPlaceholder(t *testing.T) {
	// Given — component template with a slot placeholder
	input := `<box><slot:header /><slot:children /></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if len(box.Children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(box.Children))
	}

	slot0, ok := box.Children[0].(*SlotElement)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *SlotElement", box.Children[0])
	}
	if slot0.Name != "header" {
		t.Errorf("slot0.Name = %q, want %q", slot0.Name, "header")
	}

	slot1, ok := box.Children[1].(*SlotElement)
	if !ok {
		t.Fatalf("Children[1] type = %T, want *SlotElement", box.Children[1])
	}
	if slot1.Name != "children" {
		t.Errorf("slot1.Name = %q, want %q", slot1.Name, "children")
	}
}

func TestParseSlotWithDefault(t *testing.T) {
	// Given — slot with default content
	input := `<box><slot:header><text>Default Title</text></slot:header></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	slot, ok := box.Children[0].(*SlotElement)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *SlotElement", box.Children[0])
	}
	if slot.Name != "header" {
		t.Errorf("Name = %q, want %q", slot.Name, "header")
	}
	if len(slot.Default) != 1 {
		t.Fatalf("len(Default) = %d, want 1", len(slot.Default))
	}
}

func TestParseSlotDefinition(t *testing.T) {
	// Given — consumer provides slot content
	input := `<box>{slot header}<text>My Title</text>{/slot}<text>Body</text></box>`

	// When
	doc, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	box := doc.Children[0].(*BoxElement)
	if len(box.Children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(box.Children))
	}

	slotDef, ok := box.Children[0].(*SlotDefNode)
	if !ok {
		t.Fatalf("Children[0] type = %T, want *SlotDefNode", box.Children[0])
	}
	if slotDef.Name != "header" {
		t.Errorf("Name = %q, want %q", slotDef.Name, "header")
	}
	if len(slotDef.Children) != 1 {
		t.Fatalf("len(slotDef.Children) = %d, want 1", len(slotDef.Children))
	}
}
