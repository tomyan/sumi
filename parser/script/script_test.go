package script

import (
	"testing"
)

func TestEmptyScript(t *testing.T) {
	s, err := Parse("")
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
	s, err := Parse("   \n\n\t  \n")
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
	s, err := Parse(`count := $state(0)`)
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
	s, err := Parse(`name := $state("world")`)
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
	input := `count := $state(0)
name := $state("hello")`
	s, err := Parse(input)
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
	s, err := Parse(`items := $state([]string{"a", "b"})`)
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
	s, err := Parse(`val := $state(max(1, min(2, 3)))`)
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

func TestSimpleFunction(t *testing.T) {
	input := `func increment() {
	count = count + 1
}`
	s, err := Parse(input)
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
	input := `func handleKey(key string) {
	name = key
}`
	s, err := Parse(input)
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
	input := `count := $state(0)

func increment() {
	count = count + 1
}`
	s, err := Parse(input)
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
	input := `count := $state(0)
name := $state("world")

func reset() {
	count = 0
	name = "world"
}`
	s, err := Parse(input)
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
	input := `count := $state(0)

func doSomething() {
	x = 42
	count = count + 1
}`
	s, err := Parse(input)
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
	input := `func increment() {
	count = count + 1
}

func decrement() {
	count = count - 1
}`
	s, err := Parse(input)
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
	input := `func doThings() {
	if true {
		count = 1
	}
}`
	s, err := Parse(input)
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
	input := `count := $state(0)
name := $state("world")

func increment() {
	count = count + 1
}

func reset() {
	count = 0
	name = "world"
}`
	s, err := Parse(input)
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

func TestStateWithBacktickString(t *testing.T) {
	s, err := Parse("msg := $state(`hello world`)")
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
	_, err := Parse("count := $state(0")
	if err == nil {
		t.Fatal("expected error for unterminated $state, got nil")
	}
}

func TestUnterminatedFuncBody(t *testing.T) {
	_, err := Parse("func foo() {")
	if err == nil {
		t.Fatal("expected error for unterminated function body, got nil")
	}
}

func TestStateAssignmentWithCompoundExpr(t *testing.T) {
	input := `count := $state(0)

func update() {
	count = append(items, "new")
}`
	s, err := Parse(input)
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
	input := `count := $state(0)

func doThing() {
	counter = 5
	count = 1
}`
	s, err := Parse(input)
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
	input := `count := $state(0)

func doThing() {
	count := 5
}`
	s, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// := is a short variable declaration, not an assignment to state
	if len(s.FuncDecls[0].StateAssignments) != 0 {
		t.Errorf("expected 0 state assignments for :=, got %d", len(s.FuncDecls[0].StateAssignments))
	}
}
