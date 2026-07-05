package style

import "testing"

// A5: structural pseudo-classes parse into Structural; state pseudos stay in
// Pseudo; sibling combinators tokenize correctly.

func TestParseStructuralPseudoSeparatedFromState(t *testing.T) {
	sels, err := ParseSelectorList("text:first-child:hover")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	p := sels[0].Parts[0]
	if p.Pseudo != "hover" {
		t.Errorf("Pseudo = %q, want hover", p.Pseudo)
	}
	if len(p.Structural) != 1 || p.Structural[0] != "first-child" {
		t.Errorf("Structural = %+v, want [first-child]", p.Structural)
	}
}

func TestParseNthChildArgument(t *testing.T) {
	sels, err := ParseSelectorList("text:nth-child(2n+1)")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	p := sels[0].Parts[0]
	if len(p.Structural) != 1 || p.Structural[0] != "nth-child(2n+1)" {
		t.Errorf("Structural = %+v", p.Structural)
	}
}

func TestParseSiblingCombinators(t *testing.T) {
	cases := []struct {
		sel  string
		comb byte
	}{
		{".a + .b", '+'},
		{".a ~ .b", '~'},
		{".a+.b", '+'},
		{".a~.b", '~'},
	}
	for _, c := range cases {
		sels, err := ParseSelectorList(c.sel)
		if err != nil {
			t.Fatalf("parse error for %q: %v", c.sel, err)
		}
		s := sels[0]
		if len(s.Parts) != 2 || s.Combinators[0] != c.comb {
			t.Errorf("%q → %+v, want combinator %q", c.sel, s, string(c.comb))
		}
	}
}

func TestParsePlusInsideNthNotACombinator(t *testing.T) {
	sels, err := ParseSelectorList(".list :nth-child(2n+1)")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	s := sels[0]
	if len(s.Parts) != 2 || s.Combinators[0] != ' ' {
		t.Errorf("expected descendant with nth argument intact, got %+v", s)
	}
	if s.Parts[1].Structural[0] != "nth-child(2n+1)" {
		t.Errorf("nth argument mangled: %+v", s.Parts[1])
	}
}

func TestStructuralPseudoCountsInSpecificity(t *testing.T) {
	lo, _ := ParseSelectorList("text")
	hi, _ := ParseSelectorList("text:first-child")
	if !lo[0].Specificity().Less(hi[0].Specificity()) {
		t.Error(":first-child should add class-level specificity")
	}
}
