package script

import (
	"testing"
)

func TestEmptyScriptNoEnvDecls(t *testing.T) {
	// When
	s, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.EnvDecls) != 0 {
		t.Errorf("expected 0 env decls, got %d", len(s.EnvDecls))
	}
}

func TestSingleEnvWidth(t *testing.T) {
	// Given
	input := `w := $env(width)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.EnvDecls) != 1 {
		t.Fatalf("expected 1 env decl, got %d", len(s.EnvDecls))
	}
	if s.EnvDecls[0].Name != "w" {
		t.Errorf("name: got %q, want %q", s.EnvDecls[0].Name, "w")
	}
	if s.EnvDecls[0].Key != "width" {
		t.Errorf("key: got %q, want %q", s.EnvDecls[0].Key, "width")
	}
}

func TestSingleEnvHeight(t *testing.T) {
	// Given
	input := `h := $env(height)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.EnvDecls) != 1 {
		t.Fatalf("expected 1 env decl, got %d", len(s.EnvDecls))
	}
	if s.EnvDecls[0].Name != "h" {
		t.Errorf("name: got %q, want %q", s.EnvDecls[0].Name, "h")
	}
	if s.EnvDecls[0].Key != "height" {
		t.Errorf("key: got %q, want %q", s.EnvDecls[0].Key, "height")
	}
}

func TestMultipleEnvDecls(t *testing.T) {
	// Given
	input := `width := $env(width)
height := $env(height)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.EnvDecls) != 2 {
		t.Fatalf("expected 2 env decls, got %d", len(s.EnvDecls))
	}
	if s.EnvDecls[0].Name != "width" {
		t.Errorf("first name: got %q, want %q", s.EnvDecls[0].Name, "width")
	}
	if s.EnvDecls[0].Key != "width" {
		t.Errorf("first key: got %q, want %q", s.EnvDecls[0].Key, "width")
	}
	if s.EnvDecls[1].Name != "height" {
		t.Errorf("second name: got %q, want %q", s.EnvDecls[1].Name, "height")
	}
	if s.EnvDecls[1].Key != "height" {
		t.Errorf("second key: got %q, want %q", s.EnvDecls[1].Key, "height")
	}
}

func TestMixedStateAndEnv(t *testing.T) {
	// Given
	input := `count := $state(0)
width := $env(width)

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
	if len(s.EnvDecls) != 1 {
		t.Fatalf("expected 1 env decl, got %d", len(s.EnvDecls))
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if s.StateDecls[0].Name != "count" {
		t.Errorf("state name: got %q, want %q", s.StateDecls[0].Name, "count")
	}
	if s.EnvDecls[0].Name != "width" {
		t.Errorf("env name: got %q, want %q", s.EnvDecls[0].Name, "width")
	}
}

func TestUnterminatedEnv(t *testing.T) {
	// When
	_, err := Parse(`w := $env(width`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated $env, got nil")
	}
}
