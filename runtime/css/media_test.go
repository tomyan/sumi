package css

import "testing"

// A10a: compile-time media evaluation — display-mode only; conditions that
// depend on runtime state are not yet supported and skip the block.

func TestDisplayModeTerminalRulesApply(t *testing.T) {
	ss := mustParse(t, `
.card { color: red; }
@media (display-mode: terminal) { .card { border: single; } }
@media (display-mode: browser) { .card { color: blue; } }
`)
	props := Resolve(ss, []Element{el("box", "card")})
	if props["border"] != "single" {
		t.Errorf("terminal media rules must apply: %+v", props)
	}
	if props["color"] != "red" {
		t.Errorf("browser media rules must not apply: %+v", props)
	}
}

func TestUnsupportedMediaConditionsSkip(t *testing.T) {
	ss := mustParse(t, `@media (min-width: 40) { .card { color: red; } }`)
	if got := Resolve(ss, []Element{el("box", "card")})["color"]; got != "" {
		t.Errorf("runtime-dependent media must skip for now, got %q", got)
	}
}

func TestMediaRulesKeepCascadeOrder(t *testing.T) {
	// A media rule later in the sheet overrides an equal-specificity rule.
	ss := mustParse(t, `
.card { color: red; }
@media (display-mode: terminal) { .card { color: green; } }
`)
	if got := Resolve(ss, []Element{el("box", "card")})["color"]; got != "green" {
		t.Errorf("color = %q, want green (later media rule wins)", got)
	}
}
