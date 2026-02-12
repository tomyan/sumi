package script

import "testing"

func TestTextInputHandlerParsing(t *testing.T) {
	// Given — a realistic TextInput script block with nested conditionals
	input := `value := $prop("")
cursor := $state(0)

func handleEvent(evt input.Event) {
	if evt.Special == input.KeyBackspace && cursor > 0 {
		value = value[:cursor-1] + value[cursor:]
		cursor = cursor - 1
	}
	if evt.Kind == input.EventKey {
		value = value[:cursor] + string(evt.Rune) + value[cursor:]
		cursor = cursor + 1
	}
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.PropDecls) != 1 {
		t.Fatalf("expected 1 prop, got %d", len(s.PropDecls))
	}
	if s.PropDecls[0].Name != "value" {
		t.Errorf("prop name: got %q, want %q", s.PropDecls[0].Name, "value")
	}

	if len(s.StateDecls) != 1 {
		t.Fatalf("expected 1 state, got %d", len(s.StateDecls))
	}
	if s.StateDecls[0].Name != "cursor" {
		t.Errorf("state name: got %q, want %q", s.StateDecls[0].Name, "cursor")
	}

	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func, got %d", len(s.FuncDecls))
	}
	fd := s.FuncDecls[0]
	if fd.Name != "handleEvent" {
		t.Errorf("func name: got %q, want %q", fd.Name, "handleEvent")
	}
	if fd.Params != "evt input.Event" {
		t.Errorf("params: got %q, want %q", fd.Params, "evt input.Event")
	}

	// State assignments should detect assignments to both value (prop) and cursor (state)
	if len(fd.StateAssignments) != 4 {
		t.Fatalf("expected 4 state assignments, got %d", len(fd.StateAssignments))
	}

	// Check each assignment
	expected := []struct {
		varName string
		line    string
	}{
		{"value", "value = value[:cursor-1] + value[cursor:]"},
		{"cursor", "cursor = cursor - 1"},
		{"value", "value = value[:cursor] + string(evt.Rune) + value[cursor:]"},
		{"cursor", "cursor = cursor + 1"},
	}
	for i, exp := range expected {
		if fd.StateAssignments[i].VarName != exp.varName {
			t.Errorf("assignment[%d] var: got %q, want %q", i, fd.StateAssignments[i].VarName, exp.varName)
		}
		if fd.StateAssignments[i].Line != exp.line {
			t.Errorf("assignment[%d] line: got %q, want %q", i, fd.StateAssignments[i].Line, exp.line)
		}
	}
}
