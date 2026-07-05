package style

import "testing"

// A5: attribute selectors — [a], [a=v], [a^=v], [a$=v], [a*=v], [a~=v], [a|=v].

func attrOf(t *testing.T, selector string) []AttrMatcher {
	t.Helper()
	sels, err := ParseSelectorList(selector)
	if err != nil {
		t.Fatalf("parse error for %q: %v", selector, err)
	}
	return sels[0].Parts[0].Attrs
}

func TestParseAttributePresence(t *testing.T) {
	attrs := attrOf(t, "[focusable]")
	if len(attrs) != 1 || attrs[0].Name != "focusable" || attrs[0].Op != "" {
		t.Errorf("attrs = %+v", attrs)
	}
}

func TestParseAttributeOperators(t *testing.T) {
	cases := []struct {
		sel, op string
	}{
		{`[a=v]`, "="},
		{`[a^=v]`, "^="},
		{`[a$=v]`, "$="},
		{`[a*=v]`, "*="},
		{`[a~=v]`, "~="},
		{`[a|=v]`, "|="},
	}
	for _, c := range cases {
		attrs := attrOf(t, c.sel)
		if len(attrs) != 1 || attrs[0].Op != c.op || attrs[0].Value != "v" {
			t.Errorf("%s → %+v, want op %q value v", c.sel, attrs, c.op)
		}
	}
}

func TestParseAttributeQuotedValue(t *testing.T) {
	attrs := attrOf(t, `[title="hello"]`)
	if attrs[0].Value != "hello" {
		t.Errorf("quoted value = %q, want hello", attrs[0].Value)
	}
}

func TestParseAttributeOnCompound(t *testing.T) {
	sels, err := ParseSelectorList(`box.panel[focusable=true]`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	p := sels[0].Parts[0]
	if p.Tag != "box" || len(p.Classes) != 1 || len(p.Attrs) != 1 {
		t.Errorf("compound = %+v", p)
	}
}

func TestAttributeSelectorSpecificityCountsAsClass(t *testing.T) {
	lo, _ := ParseSelectorList("box")
	hi, _ := ParseSelectorList("[focusable]")
	if !lo[0].Specificity().Less(hi[0].Specificity()) {
		t.Error("attribute selector should outrank bare type")
	}
}
