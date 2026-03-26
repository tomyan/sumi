package script

import "testing"

func TestParseSignalNew(t *testing.T) {
	// Given
	input := `count := sumi.New(0)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SignalDecls) != 1 {
		t.Fatalf("len(SignalDecls) = %d, want 1", len(s.SignalDecls))
	}
	if s.SignalDecls[0].Name != "count" {
		t.Errorf("Name = %q, want %q", s.SignalDecls[0].Name, "count")
	}
	if s.SignalDecls[0].InitExpr != "0" {
		t.Errorf("InitExpr = %q, want %q", s.SignalDecls[0].InitExpr, "0")
	}
}

func TestParseSignalNewString(t *testing.T) {
	// Given
	input := `name := sumi.New("hello")`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SignalDecls) != 1 {
		t.Fatalf("len(SignalDecls) = %d, want 1", len(s.SignalDecls))
	}
	if s.SignalDecls[0].Name != "name" {
		t.Errorf("Name = %q, want %q", s.SignalDecls[0].Name, "name")
	}
	if s.SignalDecls[0].InitExpr != `"hello"` {
		t.Errorf("InitExpr = %q, want %q", s.SignalDecls[0].InitExpr, `"hello"`)
	}
}

func TestParseSignalNewWithPackagePrefix(t *testing.T) {
	// Given — signal.New() instead of sumi.New()
	input := `count := signal.New(42)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SignalDecls) != 1 {
		t.Fatalf("len(SignalDecls) = %d, want 1", len(s.SignalDecls))
	}
	if s.SignalDecls[0].InitExpr != "42" {
		t.Errorf("InitExpr = %q, want %q", s.SignalDecls[0].InitExpr, "42")
	}
}

func TestParseSignalFrom(t *testing.T) {
	// Given
	input := `doubled := sumi.From(func() int { return count.Get() * 2 })`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.ComputedDecls) != 1 {
		t.Fatalf("len(ComputedDecls) = %d, want 1", len(s.ComputedDecls))
	}
	if s.ComputedDecls[0].Name != "doubled" {
		t.Errorf("Name = %q, want %q", s.ComputedDecls[0].Name, "doubled")
	}
	if s.ComputedDecls[0].Expr != "func() int { return count.Get() * 2 }" {
		t.Errorf("Expr = %q, want func literal", s.ComputedDecls[0].Expr)
	}
}

func TestParseMixedSignalAndFunc(t *testing.T) {
	// Given
	input := `count := sumi.New(0)

func increment() {
    count.Update(func(n int) int { return n + 1 })
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SignalDecls) != 1 {
		t.Errorf("len(SignalDecls) = %d, want 1", len(s.SignalDecls))
	}
	if len(s.FuncDecls) != 1 {
		t.Errorf("len(FuncDecls) = %d, want 1", len(s.FuncDecls))
	}
}
