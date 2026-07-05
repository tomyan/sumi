package css

import "testing"

// A5: attribute selector matching against element attributes.

func elWithAttrs(tag string, attrs map[string]string) Element {
	return Element{Tag: tag, Attrs: attrs}
}

func TestResolveAttributePresence(t *testing.T) {
	ss := mustParse(t, `[focusable] { color: cyan; }`)
	path := []Element{elWithAttrs("box", map[string]string{"focusable": "true"})}
	if got := Resolve(ss, path)["color"]; got != "cyan" {
		t.Errorf("color = %q, want cyan", got)
	}
	bare := []Element{elWithAttrs("box", nil)}
	if got := Resolve(ss, bare)["color"]; got != "" {
		t.Errorf("color = %q, want empty without attribute", got)
	}
}

func TestResolveAttributeOperators(t *testing.T) {
	cases := []struct {
		selector string
		value    string
		match    bool
	}{
		{`[a=hello]`, "hello", true},
		{`[a=hello]`, "hell", false},
		{`[a^=he]`, "hello", true},
		{`[a^=lo]`, "hello", false},
		{`[a$=lo]`, "hello", true},
		{`[a$=he]`, "hello", false},
		{`[a*=ell]`, "hello", true},
		{`[a*=xyz]`, "hello", false},
		{`[a~=b]`, "a b c", true},
		{`[a~=d]`, "a b c", false},
		{`[a|=en]`, "en-GB", true},
		{`[a|=en]`, "en", true},
		{`[a|=en]`, "enx", false},
	}
	for _, c := range cases {
		ss := mustParse(t, c.selector+` { color: red; }`)
		path := []Element{elWithAttrs("box", map[string]string{"a": c.value})}
		got := Resolve(ss, path)["color"] == "red"
		if got != c.match {
			t.Errorf("%s vs %q: match = %v, want %v", c.selector, c.value, got, c.match)
		}
	}
}
