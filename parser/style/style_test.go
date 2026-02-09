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

func TestComments(t *testing.T) {
	// Given
	input := `/* header comment */
.title {
	/* color is green */
	color: green;
}
/* footer */`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	if s.Rules[0].Properties["color"] != "green" {
		t.Errorf("color: got %q, want %q", s.Rules[0].Properties["color"], "green")
	}
}

func TestUnterminatedBlock(t *testing.T) {
	// When
	_, err := Parse(`.title { color: green;`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated block, got nil")
	}
}

func TestMissingSelectorBeforeBrace(t *testing.T) {
	// When
	_, err := Parse(`{ color: green; }`)

	// Then
	if err == nil {
		t.Fatal("expected error for missing selector, got nil")
	}
}

func TestPhase4DemoStylesheet(t *testing.T) {
	// Given
	input := `.container {
	border: single;
	padding: 1 2;
}
.title {
	color: green;
	bold: true;
}
.subtitle {
	color: cyan;
	dim: true;
}
.count {
	color: yellow;
	bold: true;
}`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(s.Rules))
	}

	// .container
	if s.Rules[0].Selector != ".container" {
		t.Errorf("rule 0 selector: got %q, want %q", s.Rules[0].Selector, ".container")
	}
	if s.Rules[0].Properties["border"] != "single" {
		t.Errorf("container border: got %q, want %q", s.Rules[0].Properties["border"], "single")
	}
	if s.Rules[0].Properties["padding"] != "1 2" {
		t.Errorf("container padding: got %q, want %q", s.Rules[0].Properties["padding"], "1 2")
	}

	// .title
	if s.Rules[1].Selector != ".title" {
		t.Errorf("rule 1 selector: got %q, want %q", s.Rules[1].Selector, ".title")
	}
	if s.Rules[1].Properties["color"] != "green" {
		t.Errorf("title color: got %q, want %q", s.Rules[1].Properties["color"], "green")
	}
	if s.Rules[1].Properties["bold"] != "true" {
		t.Errorf("title bold: got %q, want %q", s.Rules[1].Properties["bold"], "true")
	}

	// .subtitle
	if s.Rules[2].Selector != ".subtitle" {
		t.Errorf("rule 2 selector: got %q, want %q", s.Rules[2].Selector, ".subtitle")
	}
	if s.Rules[2].Properties["color"] != "cyan" {
		t.Errorf("subtitle color: got %q, want %q", s.Rules[2].Properties["color"], "cyan")
	}
	if s.Rules[2].Properties["dim"] != "true" {
		t.Errorf("subtitle dim: got %q, want %q", s.Rules[2].Properties["dim"], "true")
	}

	// .count
	if s.Rules[3].Selector != ".count" {
		t.Errorf("rule 3 selector: got %q, want %q", s.Rules[3].Selector, ".count")
	}
	if s.Rules[3].Properties["color"] != "yellow" {
		t.Errorf("count color: got %q, want %q", s.Rules[3].Properties["color"], "yellow")
	}
	if s.Rules[3].Properties["bold"] != "true" {
		t.Errorf("count bold: got %q, want %q", s.Rules[3].Properties["bold"], "true")
	}
}

func TestMixedElementAndClassSelectors(t *testing.T) {
	// Given
	input := `text { color: white; }
.highlight { bold: true; }`

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(s.Rules))
	}
	if s.Rules[0].Selector != "text" {
		t.Errorf("first selector: got %q, want %q", s.Rules[0].Selector, "text")
	}
	if s.Rules[1].Selector != ".highlight" {
		t.Errorf("second selector: got %q, want %q", s.Rules[1].Selector, ".highlight")
	}
}

func TestUnterminatedComment(t *testing.T) {
	// When
	_, err := Parse(`/* unterminated comment .title { color: green; }`)

	// Then
	if err == nil {
		t.Fatal("expected error for unterminated comment, got nil")
	}
}

func TestNoSemicolonMultipleProperties(t *testing.T) {
	// When
	s, err := Parse(`.x { color: red; bold: true }`)

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
	if s.Rules[0].Properties["bold"] != "true" {
		t.Errorf("bold: got %q, want %q", s.Rules[0].Properties["bold"], "true")
	}
}

func TestExtraSpacesAroundColon(t *testing.T) {
	// When
	s, err := Parse(`.x { color :  green ; }`)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(s.Rules))
	}
	if s.Rules[0].Properties["color"] != "green" {
		t.Errorf("color: got %q, want %q", s.Rules[0].Properties["color"], "green")
	}
}

func TestTabsInProperties(t *testing.T) {
	// Given
	input := ".x {\n\tcolor:\tgreen;\n\tbold:\ttrue;\n}"

	// When
	s, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Rules[0].Properties["color"] != "green" {
		t.Errorf("color: got %q, want %q", s.Rules[0].Properties["color"], "green")
	}
	if s.Rules[0].Properties["bold"] != "true" {
		t.Errorf("bold: got %q, want %q", s.Rules[0].Properties["bold"], "true")
	}
}
