package script

import (
	"testing"
)

func TestEmptyScript(t *testing.T) {
	// When
	s, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 0 {
		t.Errorf("expected 0 state decls, got %d", len(s.StateDecls))
	}
	if len(s.FuncDecls) != 0 {
		t.Errorf("expected 0 func decls, got %d", len(s.FuncDecls))
	}
}

func TestWhitespaceOnlyScript(t *testing.T) {
	// When
	s, err := Parse("   \n\n\t  \n")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 0 {
		t.Errorf("expected 0 state decls, got %d", len(s.StateDecls))
	}
	if len(s.FuncDecls) != 0 {
		t.Errorf("expected 0 func decls, got %d", len(s.FuncDecls))
	}
}

func TestSingleStateInt(t *testing.T) {
	// When
	s, err := Parse(`count := $state(0)`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Fatalf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].Name != "count" {
		t.Errorf("name: got %q, want %q", s.StateDecls[0].Name, "count")
	}
	if s.StateDecls[0].InitExpr != "0" {
		t.Errorf("init expr: got %q, want %q", s.StateDecls[0].InitExpr, "0")
	}
}

func TestSingleStateString(t *testing.T) {
	// When
	s, err := Parse(`name := $state("world")`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Fatalf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].Name != "name" {
		t.Errorf("name: got %q, want %q", s.StateDecls[0].Name, "name")
	}
	if s.StateDecls[0].InitExpr != `"world"` {
		t.Errorf("init expr: got %q, want %q", s.StateDecls[0].InitExpr, `"world"`)
	}
}

func TestMultipleStateDecls(t *testing.T) {
	// Given
	input := `count := $state(0)
name := $state("hello")`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 2 {
		t.Fatalf("expected 2 state decls, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].Name != "count" {
		t.Errorf("first name: got %q, want %q", s.StateDecls[0].Name, "count")
	}
	if s.StateDecls[0].InitExpr != "0" {
		t.Errorf("first init expr: got %q, want %q", s.StateDecls[0].InitExpr, "0")
	}
	if s.StateDecls[1].Name != "name" {
		t.Errorf("second name: got %q, want %q", s.StateDecls[1].Name, "name")
	}
	if s.StateDecls[1].InitExpr != `"hello"` {
		t.Errorf("second init expr: got %q, want %q", s.StateDecls[1].InitExpr, `"hello"`)
	}
}

func TestStateWithNestedParens(t *testing.T) {
	// When
	s, err := Parse(`items := $state([]string{"a", "b"})`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Fatalf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].Name != "items" {
		t.Errorf("name: got %q, want %q", s.StateDecls[0].Name, "items")
	}
	if s.StateDecls[0].InitExpr != `[]string{"a", "b"}` {
		t.Errorf("init expr: got %q, want %q", s.StateDecls[0].InitExpr, `[]string{"a", "b"}`)
	}
}

func TestStateWithNestedParensInExpr(t *testing.T) {
	// When
	s, err := Parse(`val := $state(max(1, min(2, 3)))`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Fatalf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].InitExpr != "max(1, min(2, 3))" {
		t.Errorf("init expr: got %q, want %q", s.StateDecls[0].InitExpr, "max(1, min(2, 3))")
	}
}

func TestStateWithBacktickString(t *testing.T) {
	// When
	s, err := Parse("msg := $state(`hello world`)")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Fatalf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].InitExpr != "`hello world`" {
		t.Errorf("init expr: got %q, want %q", s.StateDecls[0].InitExpr, "`hello world`")
	}
}

func TestUnterminatedState(t *testing.T) {
	// When
	_, err := Parse("count := $state(0")

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated $state, got nil")
	}
}
