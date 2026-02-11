package codegen

import (
	"testing"
)

func TestPrefixConditionSimple(t *testing.T) {
	// Given
	varNames := map[string]bool{"count": true}

	// When
	result := prefixConditionExpr("count > 0", varNames)

	// Then
	if result != "c.count > 0" {
		t.Errorf("got %q, want %q", result, "c.count > 0")
	}
}

func TestPrefixConditionMultipleVars(t *testing.T) {
	// Given
	varNames := map[string]bool{"count": true, "visible": true}

	// When
	result := prefixConditionExpr("count > 0 && visible", varNames)

	// Then
	expected := "c.count > 0 && c.visible"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestPrefixConditionIgnoresBuiltins(t *testing.T) {
	// Given
	varNames := map[string]bool{"items": true}

	// When
	result := prefixConditionExpr("len(items) > 0", varNames)

	// Then
	expected := "len(c.items) > 0"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestPrefixConditionNoMatch(t *testing.T) {
	// Given
	varNames := map[string]bool{"count": true}

	// When
	result := prefixConditionExpr("x > 0", varNames)

	// Then
	if result != "x > 0" {
		t.Errorf("got %q, want %q", result, "x > 0")
	}
}

func TestPrefixForClause(t *testing.T) {
	// Given
	varNames := map[string]bool{"items": true}

	// When
	result := prefixConditionExpr("i, item := range items", varNames)

	// Then
	expected := "i, item := range c.items"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}
