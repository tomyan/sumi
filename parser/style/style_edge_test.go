package style

import (
	"testing"
)

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
