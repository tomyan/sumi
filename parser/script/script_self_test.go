package script

import (
	"testing"
)

func TestEmptyScriptNoSelfDecls(t *testing.T) {
	// When
	s, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SelfDecls) != 0 {
		t.Errorf("expected 0 self decls, got %d", len(s.SelfDecls))
	}
}

func TestSelfWidth(t *testing.T) {
	// Given
	input := `selfW := $self(width)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SelfDecls) != 1 {
		t.Fatalf("expected 1 self decl, got %d", len(s.SelfDecls))
	}
	if s.SelfDecls[0].Name != "selfW" {
		t.Errorf("name: got %q, want %q", s.SelfDecls[0].Name, "selfW")
	}
	if s.SelfDecls[0].Key != "width" {
		t.Errorf("key: got %q, want %q", s.SelfDecls[0].Key, "width")
	}
}

func TestSelfHeight(t *testing.T) {
	// Given
	input := `selfH := $self(height)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SelfDecls) != 1 {
		t.Fatalf("expected 1 self decl, got %d", len(s.SelfDecls))
	}
	if s.SelfDecls[0].Name != "selfH" {
		t.Errorf("name: got %q, want %q", s.SelfDecls[0].Name, "selfH")
	}
	if s.SelfDecls[0].Key != "height" {
		t.Errorf("key: got %q, want %q", s.SelfDecls[0].Key, "height")
	}
}

func TestMultipleSelfDecls(t *testing.T) {
	// Given
	input := `selfW := $self(width)
selfH := $self(height)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SelfDecls) != 2 {
		t.Fatalf("expected 2 self decls, got %d", len(s.SelfDecls))
	}
	if s.SelfDecls[0].Name != "selfW" {
		t.Errorf("first name: got %q, want %q", s.SelfDecls[0].Name, "selfW")
	}
	if s.SelfDecls[0].Key != "width" {
		t.Errorf("first key: got %q, want %q", s.SelfDecls[0].Key, "width")
	}
	if s.SelfDecls[1].Name != "selfH" {
		t.Errorf("second name: got %q, want %q", s.SelfDecls[1].Name, "selfH")
	}
	if s.SelfDecls[1].Key != "height" {
		t.Errorf("second key: got %q, want %q", s.SelfDecls[1].Key, "height")
	}
}

func TestMixedStateAndSelf(t *testing.T) {
	// Given
	input := `count := $state(0)
selfW := $self(width)

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
		t.Fatalf("expected 1 state decl, got %d", len(s.StateDecls))
	}
	if len(s.SelfDecls) != 1 {
		t.Fatalf("expected 1 self decl, got %d", len(s.SelfDecls))
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if s.StateDecls[0].Name != "count" {
		t.Errorf("state name: got %q, want %q", s.StateDecls[0].Name, "count")
	}
	if s.SelfDecls[0].Name != "selfW" {
		t.Errorf("self name: got %q, want %q", s.SelfDecls[0].Name, "selfW")
	}
}

func TestUnterminatedSelf(t *testing.T) {
	// When
	_, err := Parse(`selfW := $self(width`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated $self, got nil")
	}
}

func TestNonMatchingInputFallsThrough(t *testing.T) {
	// Given - $self not present, just a regular function
	input := `func doThing() {
	x := 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.SelfDecls) != 0 {
		t.Errorf("expected 0 self decls, got %d", len(s.SelfDecls))
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
}
