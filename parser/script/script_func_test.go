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

func TestStateAssignmentDetection(t *testing.T) {
	// Given
	input := `count := $state(0)

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
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if len(s.FuncDecls[0].StateAssignments) != 1 {
		t.Fatalf("expected 1 state assignment, got %d", len(s.FuncDecls[0].StateAssignments))
	}
	sa := s.FuncDecls[0].StateAssignments[0]
	if sa.VarName != "count" {
		t.Errorf("var name: got %q, want %q", sa.VarName, "count")
	}
	if sa.Line != "count = count + 1" {
		t.Errorf("line: got %q, want %q", sa.Line, "count = count + 1")
	}
}

func TestMultipleStateAssignments(t *testing.T) {
	// Given
	input := `count := $state(0)
name := $state("world")

func reset() {
	count = 0
	name = "world"
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls[0].StateAssignments) != 2 {
		t.Fatalf("expected 2 state assignments, got %d", len(s.FuncDecls[0].StateAssignments))
	}
	if s.FuncDecls[0].StateAssignments[0].VarName != "count" {
		t.Errorf("first var: got %q, want %q", s.FuncDecls[0].StateAssignments[0].VarName, "count")
	}
	if s.FuncDecls[0].StateAssignments[1].VarName != "name" {
		t.Errorf("second var: got %q, want %q", s.FuncDecls[0].StateAssignments[1].VarName, "name")
	}
}

func TestNonStateAssignmentIgnored(t *testing.T) {
	// Given
	input := `count := $state(0)

func doSomething() {
	x = 42
	count = count + 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls[0].StateAssignments) != 1 {
		t.Fatalf("expected 1 state assignment (only count), got %d", len(s.FuncDecls[0].StateAssignments))
	}
	if s.FuncDecls[0].StateAssignments[0].VarName != "count" {
		t.Errorf("var name: got %q, want %q", s.FuncDecls[0].StateAssignments[0].VarName, "count")
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

func TestMixedStateAndFunctions(t *testing.T) {
	// Given
	input := `count := $state(0)
name := $state("world")

func increment() {
	count = count + 1
}

func reset() {
	count = 0
	name = "world"
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.StateDecls) != 2 {
		t.Fatalf("expected 2 state decls, got %d", len(s.StateDecls))
	}
	if len(s.FuncDecls) != 2 {
		t.Fatalf("expected 2 func decls, got %d", len(s.FuncDecls))
	}
	// increment has 1 state assignment (count)
	if len(s.FuncDecls[0].StateAssignments) != 1 {
		t.Fatalf("increment: expected 1 state assignment, got %d", len(s.FuncDecls[0].StateAssignments))
	}
	// reset has 2 state assignments (count and name)
	if len(s.FuncDecls[1].StateAssignments) != 2 {
		t.Fatalf("reset: expected 2 state assignments, got %d", len(s.FuncDecls[1].StateAssignments))
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

func TestStateAssignmentWithCompoundExpr(t *testing.T) {
	// Given
	input := `count := $state(0)

func update() {
	count = append(items, "new")
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.FuncDecls[0].StateAssignments) != 1 {
		t.Fatalf("expected 1 state assignment, got %d", len(s.FuncDecls[0].StateAssignments))
	}
	if s.FuncDecls[0].StateAssignments[0].Line != `count = append(items, "new")` {
		t.Errorf("line: got %q, want %q", s.FuncDecls[0].StateAssignments[0].Line, `count = append(items, "new")`)
	}
}

func TestStateNamePrefixNotMatched(t *testing.T) {
	// Given
	input := `count := $state(0)

func doThing() {
	counter = 5
	count = 1
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "counter = 5" should NOT match state var "count"
	if len(s.FuncDecls[0].StateAssignments) != 1 {
		t.Fatalf("expected 1 state assignment (only count), got %d", len(s.FuncDecls[0].StateAssignments))
	}
	if s.FuncDecls[0].StateAssignments[0].VarName != "count" {
		t.Errorf("var name: got %q, want %q", s.FuncDecls[0].StateAssignments[0].VarName, "count")
	}
	if s.FuncDecls[0].StateAssignments[0].Line != "count = 1" {
		t.Errorf("line: got %q, want %q", s.FuncDecls[0].StateAssignments[0].Line, "count = 1")
	}
}

func TestShortVarDeclNotStateAssignment(t *testing.T) {
	// Given
	input := `count := $state(0)

func doThing() {
	count := 5
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// := is a short variable declaration, not an assignment to state
	if len(s.FuncDecls[0].StateAssignments) != 0 {
		t.Errorf("expected 0 state assignments for :=, got %d", len(s.FuncDecls[0].StateAssignments))
	}
}
