package css

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

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

// RS3: runtime media conditions.

func TestMinWidthMediaAgainstViewport(t *testing.T) {
	ss := mustParse(t, `@media (min-width: 100) { .wide { color: red; } }`)
	SetViewport(120, 40)
	defer SetViewport(0, 0)
	if got := Resolve(ss, []Element{el("box", "wide")})["color"]; got != "red" {
		t.Errorf("min-width should match at 120 cols, got %q", got)
	}
	SetViewport(80, 24)
	if got := Resolve(ss, []Element{el("box", "wide")})["color"]; got != "" {
		t.Errorf("min-width must not match at 80 cols, got %q", got)
	}
}

func TestMaxHeightMedia(t *testing.T) {
	ss := mustParse(t, `@media (max-height: 20) { .short { color: red; } }`)
	SetViewport(80, 15)
	defer SetViewport(0, 0)
	if got := Resolve(ss, []Element{el("box", "short")})["color"]; got != "red" {
		t.Errorf("max-height should match at 15 rows, got %q", got)
	}
}

func TestPrefersColorSchemeMedia(t *testing.T) {
	ss := mustParse(t, `@media (prefers-color-scheme: dark) { .x { color: white; } } @media (prefers-color-scheme: light) { .x { color: black; } }`)
	prev := render.GetColorScheme()
	defer render.SetColorScheme(prev)

	render.SetColorScheme(render.SchemeDark)
	if got := Resolve(ss, []Element{el("box", "x")})["color"]; got != "white" {
		t.Errorf("dark scheme = %q, want white", got)
	}
	render.SetColorScheme(render.SchemeLight)
	if got := Resolve(ss, []Element{el("box", "x")})["color"]; got != "black" {
		t.Errorf("light scheme = %q, want black", got)
	}
}

func TestCombinedMediaConditions(t *testing.T) {
	ss := mustParse(t, `@media (display-mode: terminal) and (min-width: 40) { .x { color: red; } }`)
	SetViewport(80, 24)
	defer SetViewport(0, 0)
	if got := Resolve(ss, []Element{el("box", "x")})["color"]; got != "red" {
		t.Errorf("combined conditions should match, got %q", got)
	}
}
