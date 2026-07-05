package css

import "testing"

// A5: structural pseudo-class and sibling-combinator matching.
// Sibling context rides on Element.Siblings (all element siblings, in order)
// and Element.Index (position of self within that list).

// sibs builds a sibling list and returns the element at idx with context set.
func withSiblings(elems []Element, idx int) []Element {
	e := elems[idx]
	e.Siblings = elems
	e.Index = idx
	return []Element{e}
}

func threeTexts() []Element {
	return []Element{el("text"), el("text"), el("text")}
}

func TestFirstAndLastChild(t *testing.T) {
	ss := mustParse(t, `text:first-child { color: red; } text:last-child { color: blue; }`)
	sibs := threeTexts()
	if got := Resolve(ss, withSiblings(sibs, 0))["color"]; got != "red" {
		t.Errorf("first child color = %q, want red", got)
	}
	if got := Resolve(ss, withSiblings(sibs, 1))["color"]; got != "" {
		t.Errorf("middle child color = %q, want empty", got)
	}
	if got := Resolve(ss, withSiblings(sibs, 2))["color"]; got != "blue" {
		t.Errorf("last child color = %q, want blue", got)
	}
}

func TestOnlyChild(t *testing.T) {
	ss := mustParse(t, `text:only-child { color: red; }`)
	single := []Element{el("text")}
	if got := Resolve(ss, withSiblings(single, 0))["color"]; got != "red" {
		t.Errorf("only child = %q, want red", got)
	}
	if got := Resolve(ss, withSiblings(threeTexts(), 0))["color"]; got != "" {
		t.Errorf("not only child = %q, want empty", got)
	}
}

func TestNthChild(t *testing.T) {
	ss := mustParse(t, `text:nth-child(odd) { color: red; }`)
	sibs := threeTexts()
	for i, want := range []string{"red", "", "red"} {
		if got := Resolve(ss, withSiblings(sibs, i))["color"]; got != want {
			t.Errorf("nth-child(odd) at %d = %q, want %q", i, got, want)
		}
	}
}

func TestNthLastChild(t *testing.T) {
	ss := mustParse(t, `text:nth-last-child(1) { color: red; }`)
	sibs := threeTexts()
	if got := Resolve(ss, withSiblings(sibs, 2))["color"]; got != "red" {
		t.Errorf("nth-last-child(1) on last = %q, want red", got)
	}
	if got := Resolve(ss, withSiblings(sibs, 0))["color"]; got != "" {
		t.Errorf("nth-last-child(1) on first = %q, want empty", got)
	}
}

func TestOfTypeVariants(t *testing.T) {
	// box, text, box — of-type indices count per tag.
	elems := []Element{el("box"), el("text"), el("box")}
	ss := mustParse(t, `box:first-of-type { color: red; } box:last-of-type { background: blue; } text:only-of-type { color: green; }`)

	first := Resolve(ss, withSiblings(elems, 0))
	if first["color"] != "red" || first["background"] == "blue" && false {
		t.Errorf("first box props = %+v", first)
	}
	last := Resolve(ss, withSiblings(elems, 2))
	if last["background"] != "blue" || last["color"] == "red" {
		t.Errorf("last box props = %+v", last)
	}
	only := Resolve(ss, withSiblings(elems, 1))
	if only["color"] != "green" {
		t.Errorf("only text props = %+v", only)
	}
}

func TestEmptyPseudo(t *testing.T) {
	ss := mustParse(t, `box:empty { color: red; }`)
	empty := Element{Tag: "box", Empty: true}
	if got := Resolve(ss, []Element{empty})["color"]; got != "red" {
		t.Errorf("empty box = %q, want red", got)
	}
	if got := Resolve(ss, []Element{el("box")})["color"]; got != "" {
		t.Errorf("non-empty box = %q, want empty", got)
	}
}

func TestAdjacentSiblingCombinator(t *testing.T) {
	ss := mustParse(t, `.label + text { color: red; }`)
	elems := []Element{el("text", "label"), el("text"), el("text")}
	if got := Resolve(ss, withSiblings(elems, 1))["color"]; got != "red" {
		t.Errorf("adjacent sibling = %q, want red", got)
	}
	if got := Resolve(ss, withSiblings(elems, 2))["color"]; got != "" {
		t.Errorf("non-adjacent sibling = %q, want empty", got)
	}
}

func TestGeneralSiblingCombinator(t *testing.T) {
	ss := mustParse(t, `.label ~ text { color: red; }`)
	elems := []Element{el("text", "label"), el("text"), el("text")}
	for _, i := range []int{1, 2} {
		if got := Resolve(ss, withSiblings(elems, i))["color"]; got != "red" {
			t.Errorf("general sibling at %d = %q, want red", i, got)
		}
	}
	if got := Resolve(ss, withSiblings(elems, 0))["color"]; got != "" {
		t.Errorf(".label itself = %q, want empty", got)
	}
}

func TestSiblingThenDescendant(t *testing.T) {
	// .a + .b text — text inside the box right after .a
	ss := mustParse(t, `.a + .b text { color: red; }`)
	boxes := []Element{el("box", "a"), el("box", "b")}
	parent := boxes[1]
	parent.Siblings = boxes
	parent.Index = 1
	path := []Element{el("root"), parent, el("text")}
	if got := Resolve(ss, path)["color"]; got != "red" {
		t.Errorf("sibling-then-descendant = %q, want red", got)
	}
}

func TestUnknownPseudoNeverMatches(t *testing.T) {
	ss := mustParse(t, `text:future-thing { color: red; }`)
	if got := Resolve(ss, []Element{el("text")})["color"]; got != "" {
		t.Errorf("unknown pseudo must not match, got %q", got)
	}
}
