package template

import (
	"strings"
	"testing"
)

const slotPointer = "slots were removed"

func TestParseSlotPlaceholderRejected(t *testing.T) {
	// Given — a component template using the removed <slot:name /> placeholder
	input := `<div><slot:header /></div>`

	// When
	_, err := Parse(input)

	// Then
	if err == nil {
		t.Fatalf("expected error for <slot:> placeholder, got nil")
	}
	if !strings.Contains(err.Error(), slotPointer) {
		t.Errorf("error = %q, want it to mention %q", err.Error(), slotPointer)
	}
	if !strings.Contains(err.Error(), "snippet") || !strings.Contains(err.Error(), "render") {
		t.Errorf("error = %q, want it to point at snippet/render", err.Error())
	}
}

func TestParseSlotDefinitionRejected(t *testing.T) {
	// Given — a consumer using the removed {slot name}...{/slot} block
	input := `<div>{slot header}<span>Title</span>{/slot}</div>`

	// When
	_, err := Parse(input)

	// Then
	if err == nil {
		t.Fatalf("expected error for {slot} block, got nil")
	}
	if !strings.Contains(err.Error(), slotPointer) {
		t.Errorf("error = %q, want it to mention %q", err.Error(), slotPointer)
	}
}
