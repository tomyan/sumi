package codegen

import "testing"

func TestReplaceIdentifierBasic(t *testing.T) {
	// Given — a simple expression with an identifier surrounded by spaces
	input := "count = count + 1"

	// When
	result := replaceIdentifier(input, "count", "counter0_count")

	// Then
	if result != "counter0_count = counter0_count + 1" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierArrayIndex(t *testing.T) {
	// Given — identifier followed by bracket (array indexing)
	input := "value[:cursor-1] + value[cursor:]"

	// When — replace "value"
	result := replaceIdentifier(input, "value", "name")

	// Then — both occurrences replaced
	if result != "name[:cursor-1] + name[cursor:]" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierInBrackets(t *testing.T) {
	// Given — identifier inside brackets (as index expression)
	input := "value[:cursor-1] + value[cursor:]"

	// When — replace "cursor"
	result := replaceIdentifier(input, "cursor", "textinput0_cursor")

	// Then — cursor replaced but not "cursor" inside other identifiers
	if result != "value[:textinput0_cursor-1] + value[textinput0_cursor:]" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierInCondition(t *testing.T) {
	// Given — identifier in a conditional expression
	input := "if cursor > 0 {"

	// When
	result := replaceIdentifier(input, "cursor", "textinput0_cursor")

	// Then
	if result != "if textinput0_cursor > 0 {" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierInFunctionCall(t *testing.T) {
	// Given — identifier inside a function call
	input := "len(value)"

	// When
	result := replaceIdentifier(input, "value", "name")

	// Then
	if result != "len(name)" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierNoPartialMatch(t *testing.T) {
	// Given — identifier that is a substring of another identifier
	input := "discount = counting + 1"

	// When — replace "count" should not affect "discount" or "counting"
	result := replaceIdentifier(input, "count", "x")

	// Then — unchanged
	if result != "discount = counting + 1" {
		t.Errorf("got %q, want unchanged", result)
	}
}

func TestReplaceIdentifierAtStart(t *testing.T) {
	// Given — identifier at the start of the string
	input := "cursor = 0"

	// When
	result := replaceIdentifier(input, "cursor", "textinput0_cursor")

	// Then
	if result != "textinput0_cursor = 0" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierAtEnd(t *testing.T) {
	// Given — identifier at the end of the string
	input := "x = cursor"

	// When
	result := replaceIdentifier(input, "cursor", "textinput0_cursor")

	// Then
	if result != "x = textinput0_cursor" {
		t.Errorf("got %q", result)
	}
}

func TestReplaceIdentifierMultipleOccurrences(t *testing.T) {
	// Given — multiple occurrences in different contexts
	input := "value = value[:cursor] + string(evt.Rune) + value[cursor:]"

	// When — replace "value"
	result := replaceIdentifier(input, "value", "name")

	// Then
	expected := "name = name[:cursor] + string(evt.Rune) + name[cursor:]"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestReplaceIdentifierDotAccess(t *testing.T) {
	// Given — identifier before a dot (field access)
	input := "evt.Kind == input.EventKey"

	// When — "evt" should not match "input" and vice versa
	result := replaceIdentifier(input, "input", "myinput")

	// Then — standalone "input" before dot is replaced
	if result != "evt.Kind == myinput.EventKey" {
		t.Errorf("got %q", result)
	}
}
