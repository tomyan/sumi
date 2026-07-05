package css

import (
	"testing"

	"github.com/tomyan/sumi/parser/style"
)

// mustParse builds a stylesheet from source; selector matching tests go
// through the real parser so Parsed selectors are populated.
func mustParse(t *testing.T, src string) *style.Stylesheet {
	t.Helper()
	ss, err := style.Parse(src)
	if err != nil {
		t.Fatalf("stylesheet parse error: %v", err)
	}
	return ss
}

func el(tag string, classes ...string) Element {
	return Element{Tag: tag, Classes: classes}
}

func TestResolveClassMatch(t *testing.T) {
	// Given
	ss := mustParse(t, `.title { color: red; }`)

	// When
	props := Resolve(ss, []Element{el("text", "title")})

	// Then
	if got := props["color"]; got != "red" {
		t.Errorf("color = %q, want %q", got, "red")
	}
}

func TestResolveElementMatch(t *testing.T) {
	ss := mustParse(t, `text { font-weight: bold; }`)
	props := Resolve(ss, []Element{el("text")})
	if got := props["font-weight"]; got != "bold" {
		t.Errorf("font-weight = %q, want %q", got, "bold")
	}
}

func TestResolveNoMatch(t *testing.T) {
	ss := mustParse(t, `.title { color: red; }`)
	props := Resolve(ss, []Element{el("text", "subtitle")})
	if got := props["color"]; got != "" {
		t.Errorf("color = %q, want empty", got)
	}
}

func TestResolveNoRulesEmptyProperties(t *testing.T) {
	ss := &style.Stylesheet{Rules: nil}
	props := Resolve(ss, []Element{el("text", "anything")})
	if len(props) != 0 {
		t.Errorf("got %d properties, want 0", len(props))
	}
}

func TestResolveMultipleClassesMatchMultipleRules(t *testing.T) {
	ss := mustParse(t, `.primary { color: blue; } .strong { font-weight: bold; }`)
	props := Resolve(ss, []Element{el("text", "primary", "strong")})
	if props["color"] != "blue" || props["font-weight"] != "bold" {
		t.Errorf("props = %+v", props)
	}
}

func TestResolveElementSelectorDoesNotMatchWrongTag(t *testing.T) {
	ss := mustParse(t, `box { color: red; }`)
	props := Resolve(ss, []Element{el("text")})
	if got := props["color"]; got != "" {
		t.Errorf("color = %q, want empty", got)
	}
}

// --- A4: cascade ---

func TestResolveSpecificityClassBeatsType(t *testing.T) {
	// Given: type rule appears LATER but class is more specific.
	ss := mustParse(t, `.title { color: red; } text { color: blue; }`)

	// When
	props := Resolve(ss, []Element{el("text", "title")})

	// Then
	if got := props["color"]; got != "red" {
		t.Errorf("color = %q, want red (class beats type)", got)
	}
}

func TestResolveSpecificityIDBeatsClass(t *testing.T) {
	ss := mustParse(t, `#main { color: green; } .title { color: red; }`)
	props := Resolve(ss, []Element{{Tag: "text", ID: "main", Classes: []string{"title"}}})
	if got := props["color"]; got != "green" {
		t.Errorf("color = %q, want green (id beats class)", got)
	}
}

func TestResolveEqualSpecificitySourceOrderWins(t *testing.T) {
	ss := mustParse(t, `.a { color: red; } .b { color: blue; }`)
	props := Resolve(ss, []Element{el("text", "a", "b")})
	if got := props["color"]; got != "blue" {
		t.Errorf("color = %q, want blue (later source wins)", got)
	}
}

func TestResolveLowerSpecificityStillContributesOtherProps(t *testing.T) {
	ss := mustParse(t, `text { opacity: dim; } .title { color: red; }`)
	props := Resolve(ss, []Element{el("text", "title")})
	if props["color"] != "red" || props["opacity"] != "dim" {
		t.Errorf("props = %+v", props)
	}
}

// --- A4: combinators ---

func TestResolveDescendantMatch(t *testing.T) {
	// Given
	ss := mustParse(t, `.panel text { color: red; }`)
	path := []Element{el("root"), el("box", "panel"), el("box"), el("text")}

	// When
	props := Resolve(ss, path)

	// Then
	if got := props["color"]; got != "red" {
		t.Errorf("color = %q, want red (descendant through nesting)", got)
	}
}

func TestResolveDescendantNoMatchOutside(t *testing.T) {
	ss := mustParse(t, `.panel text { color: red; }`)
	path := []Element{el("root"), el("box"), el("text")}
	if got := Resolve(ss, path)["color"]; got != "" {
		t.Errorf("color = %q, want empty (not inside .panel)", got)
	}
}

func TestResolveChildMatch(t *testing.T) {
	ss := mustParse(t, `.panel > text { color: red; }`)
	path := []Element{el("root"), el("box", "panel"), el("text")}
	if got := Resolve(ss, path)["color"]; got != "red" {
		t.Errorf("color = %q, want red (direct child)", got)
	}
}

func TestResolveChildNoMatchWhenNested(t *testing.T) {
	ss := mustParse(t, `.panel > text { color: red; }`)
	path := []Element{el("root"), el("box", "panel"), el("box"), el("text")}
	if got := Resolve(ss, path)["color"]; got != "" {
		t.Errorf("color = %q, want empty (> must be immediate)", got)
	}
}

func TestResolveUniversalMatchesEverything(t *testing.T) {
	ss := mustParse(t, `* { color: cyan; }`)
	if got := Resolve(ss, []Element{el("text")})["color"]; got != "cyan" {
		t.Errorf("color = %q, want cyan", got)
	}
}

func TestResolveIDSelector(t *testing.T) {
	ss := mustParse(t, `#sidebar { width: 20; }`)
	path := []Element{{Tag: "box", ID: "sidebar"}}
	if got := Resolve(ss, path)["width"]; got != "20" {
		t.Errorf("width = %q, want 20", got)
	}
}

func TestResolveHoverWithPath(t *testing.T) {
	// Given
	ss := mustParse(t, `.btn:hover { background: cyan; } .btn { background: blue; }`)
	path := []Element{el("root"), el("box", "btn")}

	// When / Then
	if got := Resolve(ss, path)["background"]; got != "blue" {
		t.Errorf("normal background = %q, want blue", got)
	}
	hover := ResolveHover(ss, path)
	if hover == nil || hover["background"] != "cyan" {
		t.Errorf("hover = %+v, want background cyan", hover)
	}
}

func TestResolveHoverNoMatchReturnsNil(t *testing.T) {
	ss := mustParse(t, `.btn:hover { background: cyan; }`)
	if got := ResolveHover(ss, []Element{el("text")}); got != nil {
		t.Errorf("ResolveHover = %+v, want nil", got)
	}
}
