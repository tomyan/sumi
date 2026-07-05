package css

import "testing"

// A6: :focus resolution mirrors :hover.

func TestResolveFocusWithPath(t *testing.T) {
	ss := mustParse(t, `.field:focus { border-color: cyan; } .field { border-color: white; }`)
	path := []Element{el("root"), el("box", "field")}
	if got := Resolve(ss, path)["border-color"]; got != "white" {
		t.Errorf("base border-color = %q, want white", got)
	}
	focus := ResolveFocus(ss, path)
	if focus == nil || focus["border-color"] != "cyan" {
		t.Errorf("focus = %+v, want border-color cyan", focus)
	}
}

func TestResolveFocusNoMatchReturnsNil(t *testing.T) {
	ss := mustParse(t, `.field:focus { color: red; }`)
	if got := ResolveFocus(ss, []Element{el("text")}); got != nil {
		t.Errorf("ResolveFocus = %+v, want nil", got)
	}
}
