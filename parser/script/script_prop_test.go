package script

import (
	"testing"
)

func TestEmptyScriptNoProps(t *testing.T) {
	// When
	s, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.PropDecls) != 0 {
		t.Errorf("expected 0 prop decls, got %d", len(s.PropDecls))
	}
}

func TestSinglePropString(t *testing.T) {
	// Given
	input := `label := $prop("Count")`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.PropDecls) != 1 {
		t.Fatalf("expected 1 prop decl, got %d", len(s.PropDecls))
	}
	if s.PropDecls[0].Name != "label" {
		t.Errorf("name: got %q, want %q", s.PropDecls[0].Name, "label")
	}
	if s.PropDecls[0].DefaultExpr != `"Count"` {
		t.Errorf("default expr: got %q, want %q", s.PropDecls[0].DefaultExpr, `"Count"`)
	}
}

func TestSinglePropInt(t *testing.T) {
	// Given
	input := `count := $prop(0)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.PropDecls) != 1 {
		t.Fatalf("expected 1 prop decl, got %d", len(s.PropDecls))
	}
	if s.PropDecls[0].Name != "count" {
		t.Errorf("name: got %q, want %q", s.PropDecls[0].Name, "count")
	}
	if s.PropDecls[0].DefaultExpr != "0" {
		t.Errorf("default expr: got %q, want %q", s.PropDecls[0].DefaultExpr, "0")
	}
}

func TestMultipleProps(t *testing.T) {
	// Given
	input := `label := $prop("Count")
count := $prop(0)`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.PropDecls) != 2 {
		t.Fatalf("expected 2 prop decls, got %d", len(s.PropDecls))
	}
	if s.PropDecls[0].Name != "label" {
		t.Errorf("first name: got %q, want %q", s.PropDecls[0].Name, "label")
	}
	if s.PropDecls[0].DefaultExpr != `"Count"` {
		t.Errorf("first default expr: got %q, want %q", s.PropDecls[0].DefaultExpr, `"Count"`)
	}
	if s.PropDecls[1].Name != "count" {
		t.Errorf("second name: got %q, want %q", s.PropDecls[1].Name, "count")
	}
	if s.PropDecls[1].DefaultExpr != "0" {
		t.Errorf("second default expr: got %q, want %q", s.PropDecls[1].DefaultExpr, "0")
	}
}

func TestPropWithNestedParens(t *testing.T) {
	// Given
	input := `items := $prop([]string{"a","b"})`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.PropDecls) != 1 {
		t.Fatalf("expected 1 prop decl, got %d", len(s.PropDecls))
	}
	if s.PropDecls[0].Name != "items" {
		t.Errorf("name: got %q, want %q", s.PropDecls[0].Name, "items")
	}
	if s.PropDecls[0].DefaultExpr != `[]string{"a","b"}` {
		t.Errorf("default expr: got %q, want %q", s.PropDecls[0].DefaultExpr, `[]string{"a","b"}`)
	}
}

func TestMixedStateAndProps(t *testing.T) {
	// Given
	input := `count := $state(0)
label := $prop("Count")

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
	if len(s.PropDecls) != 1 {
		t.Fatalf("expected 1 prop decl, got %d", len(s.PropDecls))
	}
	if len(s.FuncDecls) != 1 {
		t.Fatalf("expected 1 func decl, got %d", len(s.FuncDecls))
	}
	if s.StateDecls[0].Name != "count" {
		t.Errorf("state name: got %q, want %q", s.StateDecls[0].Name, "count")
	}
	if s.PropDecls[0].Name != "label" {
		t.Errorf("prop name: got %q, want %q", s.PropDecls[0].Name, "label")
	}
	if s.FuncDecls[0].Name != "increment" {
		t.Errorf("func name: got %q, want %q", s.FuncDecls[0].Name, "increment")
	}
}

func TestPropAssignmentDetected(t *testing.T) {
	// Given
	input := `label := $prop("Count")

func updateLabel() {
	label = "New Label"
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
	if len(s.FuncDecls[0].StateAssignments) != 1 {
		t.Fatalf("expected 1 state assignment, got %d", len(s.FuncDecls[0].StateAssignments))
	}
	sa := s.FuncDecls[0].StateAssignments[0]
	if sa.VarName != "label" {
		t.Errorf("var name: got %q, want %q", sa.VarName, "label")
	}
	if sa.Line != `label = "New Label"` {
		t.Errorf("line: got %q, want %q", sa.Line, `label = "New Label"`)
	}
}

func TestUnterminatedProp(t *testing.T) {
	// When
	_, err := Parse(`label := $prop("Count"`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated $prop, got nil")
	}
}
