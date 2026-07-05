package style

import (
	"reflect"
	"testing"
)

// A4: structured selectors — compounds, combinators, lists, specificity.

func TestParseSelectorSimpleClass(t *testing.T) {
	// Given / When
	sels, err := ParseSelectorList(".title")

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(sels) != 1 || len(sels[0].Parts) != 1 {
		t.Fatalf("expected 1 selector with 1 compound, got %+v", sels)
	}
	got := sels[0].Parts[0]
	if got.Tag != "" || !reflect.DeepEqual(got.Classes, []string{"title"}) {
		t.Errorf("compound = %+v", got)
	}
}

func TestParseSelectorTagWithClassAndID(t *testing.T) {
	sels, err := ParseSelectorList("box.panel#main")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := sels[0].Parts[0]
	if got.Tag != "box" || got.ID != "main" || !reflect.DeepEqual(got.Classes, []string{"panel"}) {
		t.Errorf("compound = %+v", got)
	}
}

func TestParseSelectorUniversal(t *testing.T) {
	sels, err := ParseSelectorList("*")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	got := sels[0].Parts[0]
	if got.Tag != "" || got.ID != "" || len(got.Classes) != 0 {
		t.Errorf("universal should be empty compound, got %+v", got)
	}
}

func TestParseSelectorDescendant(t *testing.T) {
	sels, err := ParseSelectorList(".panel text")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := sels[0]
	if len(s.Parts) != 2 || len(s.Combinators) != 1 || s.Combinators[0] != ' ' {
		t.Fatalf("expected descendant chain, got %+v", s)
	}
	if s.Parts[1].Tag != "text" {
		t.Errorf("subject = %+v", s.Parts[1])
	}
}

func TestParseSelectorChild(t *testing.T) {
	sels, err := ParseSelectorList(".panel > .row")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := sels[0]
	if len(s.Parts) != 2 || s.Combinators[0] != '>' {
		t.Fatalf("expected child chain, got %+v", s)
	}
}

func TestParseSelectorList(t *testing.T) {
	sels, err := ParseSelectorList("text, .hint")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(sels) != 2 {
		t.Fatalf("expected 2 selectors, got %+v", sels)
	}
}

func TestParseSelectorPseudoOnSubject(t *testing.T) {
	sels, err := ParseSelectorList(".panel .btn:hover")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := sels[0]
	if s.Parts[1].Pseudo != "hover" {
		t.Errorf("subject pseudo = %q, want hover", s.Parts[1].Pseudo)
	}
	if s.Parts[0].Pseudo != "" {
		t.Errorf("ancestor pseudo = %q, want empty", s.Parts[0].Pseudo)
	}
}

func TestSpecificityOrdering(t *testing.T) {
	cases := []struct {
		lower, higher string
	}{
		{"text", ".title"},
		{".title", "#main"},
		{".a", ".a.b"},
		{"text", "box text"},
		{"*", "text"},
	}
	for _, c := range cases {
		lo, _ := ParseSelectorList(c.lower)
		hi, _ := ParseSelectorList(c.higher)
		if !lo[0].Specificity().Less(hi[0].Specificity()) {
			t.Errorf("specificity(%s) should be < specificity(%s)", c.lower, c.higher)
		}
	}
}

func TestRuleParsingExplodesSelectorLists(t *testing.T) {
	// Given / When
	ss, err := Parse(`text, .hint { color: red; }`)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ss.Rules) != 2 {
		t.Fatalf("expected list exploded into 2 rules, got %d", len(ss.Rules))
	}
	for _, r := range ss.Rules {
		if r.Properties["color"] != "red" {
			t.Errorf("rule %q missing properties", r.Selector)
		}
	}
}

func TestRuleParsingPopulatesParsedSelector(t *testing.T) {
	ss, err := Parse(`.panel > text { color: red; }`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	r := ss.Rules[0]
	if len(r.Parsed.Parts) != 2 || r.Parsed.Combinators[0] != '>' {
		t.Errorf("Parsed = %+v", r.Parsed)
	}
}

func TestRuleParsingKeepsSubjectPseudo(t *testing.T) {
	ss, err := Parse(`.btn:hover { color: red; }`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if ss.Rules[0].Pseudo != "hover" {
		t.Errorf("Pseudo = %q, want hover", ss.Rules[0].Pseudo)
	}
}
