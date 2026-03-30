package style

import (
	"testing"
)

func TestEmptyStylesheet(t *testing.T) {
	// When
	s, err := Parse("")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(s.Rules))
	}
}

func TestWhitespaceOnlyStylesheet(t *testing.T) {
	// When
	s, err := Parse("   \n\n\t  \n")

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(s.Rules))
	}
}

func TestSingleRuleClassSelector(t *testing.T) {
	// When
	s, err := Parse(`.title { color: green; }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	r := s.Rules[0]
	if r.Selector != ".title" {
		t.Errorf("selector: got %q, want %q", r.Selector, ".title")
	}
	if len(r.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(r.Properties))
	}
	if r.Properties["color"] != "green" {
		t.Errorf("color: got %q, want %q", r.Properties["color"], "green")
	}
}

func TestMultipleProperties(t *testing.T) {
	// When
	s, err := Parse(`.title { color: green; bold: true; }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	r := s.Rules[0]
	if r.Selector != ".title" {
		t.Errorf("selector: got %q, want %q", r.Selector, ".title")
	}
	if len(r.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(r.Properties))
	}
	if r.Properties["color"] != "green" {
		t.Errorf("color: got %q, want %q", r.Properties["color"], "green")
	}
	if r.Properties["bold"] != "true" {
		t.Errorf("bold: got %q, want %q", r.Properties["bold"], "true")
	}
}

func TestMultipleRules(t *testing.T) {
	// Given
	input := `.title { color: green; }
.subtitle { color: cyan; dim: true; }`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(s.Rules))
	}
	if s.Rules[0].Selector != ".title" {
		t.Errorf("first selector: got %q, want %q", s.Rules[0].Selector, ".title")
	}
	if s.Rules[0].Properties["color"] != "green" {
		t.Errorf("first color: got %q, want %q", s.Rules[0].Properties["color"], "green")
	}
	if s.Rules[1].Selector != ".subtitle" {
		t.Errorf("second selector: got %q, want %q", s.Rules[1].Selector, ".subtitle")
	}
	if s.Rules[1].Properties["color"] != "cyan" {
		t.Errorf("second color: got %q, want %q", s.Rules[1].Properties["color"], "cyan")
	}
	if s.Rules[1].Properties["dim"] != "true" {
		t.Errorf("second dim: got %q, want %q", s.Rules[1].Properties["dim"], "true")
	}
}

func TestElementSelector(t *testing.T) {
	// When
	s, err := Parse(`text { color: white; }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	if s.Rules[0].Selector != "text" {
		t.Errorf("selector: got %q, want %q", s.Rules[0].Selector, "text")
	}
	if s.Rules[0].Properties["color"] != "white" {
		t.Errorf("color: got %q, want %q", s.Rules[0].Properties["color"], "white")
	}
}

func TestBoxElementSelector(t *testing.T) {
	// When
	s, err := Parse(`box { border: single; direction: row; }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	if s.Rules[0].Selector != "box" {
		t.Errorf("selector: got %q, want %q", s.Rules[0].Selector, "box")
	}
	if s.Rules[0].Properties["border"] != "single" {
		t.Errorf("border: got %q, want %q", s.Rules[0].Properties["border"], "single")
	}
	if s.Rules[0].Properties["direction"] != "row" {
		t.Errorf("direction: got %q, want %q", s.Rules[0].Properties["direction"], "row")
	}
}

func TestNoSemicolonOnLastProperty(t *testing.T) {
	// When
	s, err := Parse(`.x { color: red }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	if s.Rules[0].Properties["color"] != "red" {
		t.Errorf("color: got %q, want %q", s.Rules[0].Properties["color"], "red")
	}
}

func TestMultilineWhitespace(t *testing.T) {
	// Given
	input := `
		.title {
			color: green;
			bold: true;
		}

		.subtitle {
			color: cyan;
		}
	`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(s.Rules))
	}
	if s.Rules[0].Selector != ".title" {
		t.Errorf("first selector: got %q, want %q", s.Rules[0].Selector, ".title")
	}
	if s.Rules[0].Properties["color"] != "green" {
		t.Errorf("first color: got %q, want %q", s.Rules[0].Properties["color"], "green")
	}
	if s.Rules[0].Properties["bold"] != "true" {
		t.Errorf("first bold: got %q, want %q", s.Rules[0].Properties["bold"], "true")
	}
	if s.Rules[1].Selector != ".subtitle" {
		t.Errorf("second selector: got %q, want %q", s.Rules[1].Selector, ".subtitle")
	}
	if s.Rules[1].Properties["color"] != "cyan" {
		t.Errorf("second color: got %q, want %q", s.Rules[1].Properties["color"], "cyan")
	}
}

func TestPropertyValueWithSpaces(t *testing.T) {
	// When
	s, err := Parse(`.container { padding: 1 2; }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	if s.Rules[0].Properties["padding"] != "1 2" {
		t.Errorf("padding: got %q, want %q", s.Rules[0].Properties["padding"], "1 2")
	}
}

func TestAllSupportedProperties(t *testing.T) {
	// Given
	input := `.styled {
		color: green;
		background: black;
		bold: true;
		dim: true;
		italic: true;
		underline: true;
		strikethrough: true;
		inverse: true;
		border: single;
		padding: 1 2 3 4;
		direction: row;
		border-color: cyan;
	}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	r := s.Rules[0]
	expected := map[string]string{
		"color":         "green",
		"background":    "black",
		"bold":          "true",
		"dim":           "true",
		"italic":        "true",
		"underline":     "true",
		"strikethrough": "true",
		"inverse":       "true",
		"border":        "single",
		"padding":       "1 2 3 4",
		"direction":     "row",
		"border-color":  "cyan",
	}
	if len(r.Properties) != len(expected) {
		t.Fatalf("expected %d properties, got %d", len(expected), len(r.Properties))
	}
	for k, want := range expected {
		if got := r.Properties[k]; got != want {
			t.Errorf("%s: got %q, want %q", k, got, want)
		}
	}
}

func TestHoverPseudoClass(t *testing.T) {
	// Given
	input := `.tab:hover { color: white; dim: false; }`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	r := s.Rules[0]
	if r.Selector != ".tab" {
		t.Errorf("selector = %q, want %q", r.Selector, ".tab")
	}
	if r.Pseudo != "hover" {
		t.Errorf("pseudo = %q, want %q", r.Pseudo, "hover")
	}
	if r.Properties["color"] != "white" {
		t.Errorf("color = %q, want %q", r.Properties["color"], "white")
	}
}

func TestBaseAndHoverRules(t *testing.T) {
	// Given both base and hover rules for the same class
	input := `.tab { dim: true; }
.tab:hover { dim: false; color: white; }`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(s.Rules))
	}
	if s.Rules[0].Pseudo != "" {
		t.Errorf("base rule pseudo = %q, want empty", s.Rules[0].Pseudo)
	}
	if s.Rules[1].Pseudo != "hover" {
		t.Errorf("hover rule pseudo = %q, want %q", s.Rules[1].Pseudo, "hover")
	}
}
