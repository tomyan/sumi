package script

import "testing"

func TestSingleDerived(t *testing.T) {
	// Given
	input := `doubled := $derived(count * 2)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.DerivedDecls) != 1 {
		t.Fatalf("expected 1 derived decl, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Name != "doubled" {
		t.Errorf("name: got %q, want %q", s.DerivedDecls[0].Name, "doubled")
	}
	if s.DerivedDecls[0].Expr != "count * 2" {
		t.Errorf("expr: got %q, want %q", s.DerivedDecls[0].Expr, "count * 2")
	}
}

func TestDerivedStringExpr(t *testing.T) {
	// Given — nested parens in fmt.Sprintf
	input := `label := $derived(fmt.Sprintf("x=%v", x))`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.DerivedDecls) != 1 {
		t.Fatalf("expected 1 derived decl, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Expr != `fmt.Sprintf("x=%v", x)` {
		t.Errorf("expr: got %q, want %q", s.DerivedDecls[0].Expr, `fmt.Sprintf("x=%v", x)`)
	}
}

func TestDerivedWithNestedParens(t *testing.T) {
	// Given — deeply nested parentheses
	input := `result := $derived(max(a, min(b, c)))`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.DerivedDecls) != 1 {
		t.Fatalf("expected 1 derived decl, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Expr != "max(a, min(b, c))" {
		t.Errorf("expr: got %q, want %q", s.DerivedDecls[0].Expr, "max(a, min(b, c))")
	}
}

func TestDerivedWithBacktickString(t *testing.T) {
	// Given — backtick string in expression
	input := "label := $derived(`prefix` + name)"

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.DerivedDecls) != 1 {
		t.Fatalf("expected 1 derived decl, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Expr != "`prefix` + name" {
		t.Errorf("expr: got %q, want %q", s.DerivedDecls[0].Expr, "`prefix` + name")
	}
}

func TestUnterminatedDerived(t *testing.T) {
	// When
	_, err := Parse(`doubled := $derived(count * 2`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated $derived, got nil")
	}
}

func TestMultipleDerived(t *testing.T) {
	// Given
	input := `doubled := $derived(count * 2)
tripled := $derived(count * 3)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.DerivedDecls) != 2 {
		t.Fatalf("expected 2 derived decls, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Name != "doubled" {
		t.Errorf("first name: got %q, want %q", s.DerivedDecls[0].Name, "doubled")
	}
	if s.DerivedDecls[1].Name != "tripled" {
		t.Errorf("second name: got %q, want %q", s.DerivedDecls[1].Name, "tripled")
	}
}

func TestMixedStateAndDerived(t *testing.T) {
	// Given
	input := `count := $state(0)
doubled := $derived(count * 2)

func increment() {
	count = count + 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 1 {
		t.Errorf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if len(s.DerivedDecls) != 1 {
		t.Errorf("expected 1 derived decl, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Name != "doubled" {
		t.Errorf("derived name: got %q, want %q", s.DerivedDecls[0].Name, "doubled")
	}
	if len(s.FuncDecls) != 1 {
		t.Errorf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
}

func TestDerivedOnly(t *testing.T) {
	// Given — derived without state
	input := `label := $derived("hello")`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 0 {
		t.Errorf("expected 0 state decls, got %d", len(s.StateDecls))
	}
	if len(s.DerivedDecls) != 1 {
		t.Fatalf("expected 1 derived decl, got %d", len(s.DerivedDecls))
	}
	if s.DerivedDecls[0].Expr != `"hello"` {
		t.Errorf("expr: got %q, want %q", s.DerivedDecls[0].Expr, `"hello"`)
	}
}
