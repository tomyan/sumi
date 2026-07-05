package css

import "testing"

// A11: var() custom properties.

func TestExpandVarSimple(t *testing.T) {
	vars := map[string]string{"--accent": "cyan"}
	if got := ExpandVarRefs("var(--accent)", vars); got != "cyan" {
		t.Errorf("got %q, want cyan", got)
	}
}

func TestExpandVarInsideValue(t *testing.T) {
	vars := map[string]string{"--w": "20"}
	if got := ExpandVarRefs("var(--w)cell", vars); got != "20cell" {
		t.Errorf("got %q", got)
	}
}

func TestExpandVarFallback(t *testing.T) {
	if got := ExpandVarRefs("var(--missing, red)", nil); got != "red" {
		t.Errorf("got %q, want red", got)
	}
}

func TestExpandVarFallbackWithFunction(t *testing.T) {
	got := ExpandVarRefs("var(--missing, rgb(1, 2, 3))", nil)
	if got != "rgb(1, 2, 3)" {
		t.Errorf("got %q", got)
	}
}

func TestExpandVarNestedInFallback(t *testing.T) {
	vars := map[string]string{"--b": "blue"}
	if got := ExpandVarRefs("var(--a, var(--b))", vars); got != "blue" {
		t.Errorf("got %q, want blue", got)
	}
}

func TestExpandVarValueReferencesVar(t *testing.T) {
	vars := map[string]string{"--a": "var(--b)", "--b": "green"}
	if got := ExpandVarRefs("var(--a)", vars); got != "green" {
		t.Errorf("got %q, want green", got)
	}
}

func TestExpandVarCycleSafe(t *testing.T) {
	vars := map[string]string{"--a": "var(--b)", "--b": "var(--a)"}
	if got := ExpandVarRefs("var(--a)", vars); got != "" {
		t.Errorf("cycle should collapse to empty, got %q", got)
	}
}

func TestExpandVarUnresolvedNoFallbackEmpty(t *testing.T) {
	if got := ExpandVarRefs("var(--nope)", nil); got != "" {
		t.Errorf("got %q, want empty (property will drop)", got)
	}
}

func TestExpandVarPlainValueUntouched(t *testing.T) {
	if got := ExpandVarRefs("space-between", map[string]string{"--x": "1"}); got != "space-between" {
		t.Errorf("got %q", got)
	}
}
