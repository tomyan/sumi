package css

import "testing"

// A5c: :not(), :is(), :where() logical pseudo-classes.

func TestNotExcludesMatching(t *testing.T) {
	ss := mustParse(t, `text:not(.hint) { color: red; }`)
	if got := Resolve(ss, []Element{el("text")})["color"]; got != "red" {
		t.Errorf("plain text = %q, want red", got)
	}
	if got := Resolve(ss, []Element{el("text", "hint")})["color"]; got != "" {
		t.Errorf(".hint text = %q, want empty", got)
	}
}

func TestIsMatchesAnyArgument(t *testing.T) {
	ss := mustParse(t, `:is(.a, .b) { color: red; }`)
	for _, c := range []string{"a", "b"} {
		if got := Resolve(ss, []Element{el("text", c)})["color"]; got != "red" {
			t.Errorf(".%s = %q, want red", c, got)
		}
	}
	if got := Resolve(ss, []Element{el("text", "c")})["color"]; got != "" {
		t.Errorf(".c = %q, want empty", got)
	}
}

func TestWhereMatchesLikeIs(t *testing.T) {
	ss := mustParse(t, `:where(.a) { color: red; }`)
	if got := Resolve(ss, []Element{el("text", "a")})["color"]; got != "red" {
		t.Errorf(":where(.a) = %q, want red", got)
	}
}

func TestWhereHasZeroSpecificity(t *testing.T) {
	// :where(.a) has specificity 0, so the LATER type rule outranks it...
	// actually type (0,0,1) > (0,0,0). The type rule must win despite
	// appearing first.
	ss := mustParse(t, `text { color: blue; } :where(.a) { color: red; }`)
	if got := Resolve(ss, []Element{el("text", "a")})["color"]; got != "blue" {
		t.Errorf("color = %q, want blue (:where adds no specificity)", got)
	}
}

func TestIsKeepsArgumentSpecificity(t *testing.T) {
	// :is(.a) counts the class, so it beats a type rule.
	ss := mustParse(t, `:is(.a) { color: red; } text { color: blue; }`)
	if got := Resolve(ss, []Element{el("text", "a")})["color"]; got != "red" {
		t.Errorf("color = %q, want red (:is keeps class specificity)", got)
	}
}

func TestNotWithCompoundArgument(t *testing.T) {
	ss := mustParse(t, `box:not(.panel.active) { color: red; }`)
	active := []Element{el("box", "panel", "active")}
	if got := Resolve(ss, active)["color"]; got != "" {
		t.Errorf("active panel = %q, want empty", got)
	}
	partial := []Element{el("box", "panel")}
	if got := Resolve(ss, partial)["color"]; got != "red" {
		t.Errorf("partial match = %q, want red", got)
	}
}
