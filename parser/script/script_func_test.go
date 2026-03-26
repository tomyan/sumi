package script

import (
	"testing"
)

func TestSimpleFunction(t *testing.T) {
	// Given
	input := `func increment() {
	count = count + 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if s.FuncDecls[0].Name != "increment" {
		t.Errorf("name: got %q, want %q", s.FuncDecls[0].Name, "increment")
	}
	if s.FuncDecls[0].Params != "" {
		t.Errorf("params: got %q, want %q", s.FuncDecls[0].Params, "")
	}
	expected := "\n\tcount = count + 1\n"
	if s.FuncDecls[0].Body != expected {
		t.Errorf("body: got %q, want %q", s.FuncDecls[0].Body, expected)
	}
}

func TestFunctionWithParams(t *testing.T) {
	// Given
	input := `func handleKey(key string) {
	name = key
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if s.FuncDecls[0].Name != "handleKey" {
		t.Errorf("name: got %q, want %q", s.FuncDecls[0].Name, "handleKey")
	}
	if s.FuncDecls[0].Params != "key string" {
		t.Errorf("params: got %q, want %q", s.FuncDecls[0].Params, "key string")
	}
}

func TestMultipleFunctions(t *testing.T) {
	// Given
	input := `func increment() {
	count = count + 1
}

func decrement() {
	count = count - 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls) != 2 {
		t.Fatalf("expected 2 func decls, got %d", len(s.FuncDecls))
	}
	if s.FuncDecls[0].Name != "increment" {
		t.Errorf("first func: got %q, want %q", s.FuncDecls[0].Name, "increment")
	}
	if s.FuncDecls[1].Name != "decrement" {
		t.Errorf("second func: got %q, want %q", s.FuncDecls[1].Name, "decrement")
	}
}

func TestFunctionWithNestedBraces(t *testing.T) {
	// Given
	input := `func doThings() {
	if true {
		count = 1
	}
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	expected := "\n\tif true {\n\t\tcount = 1\n\t}\n"
	if s.FuncDecls[0].Body != expected {
		t.Errorf("body: got %q, want %q", s.FuncDecls[0].Body, expected)
	}
}

func TestUnterminatedFuncBody(t *testing.T) {
	// When
	_, err := Parse("func foo() {")

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated function body, got nil")
	}
}

func TestFunctionWithReturnType(t *testing.T) {
	// Given
	input := `func buildLine() string {
	return "[" + value + "]"
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if s.FuncDecls[0].Name != "buildLine" {
		t.Errorf("name: got %q, want %q", s.FuncDecls[0].Name, "buildLine")
	}
}
