package style

import "testing"

// A10a: @media blocks parse; rules carry their query in source order.

func TestMediaBlockRulesCarryQuery(t *testing.T) {
	// Given
	input := `
.card { color: red; }
@media (display-mode: terminal) {
	.card { border: single; }
}
.other { color: blue; }
`

	// When
	ss, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ss.Rules) != 3 {
		t.Fatalf("expected 3 rules (flat, source order), got %d", len(ss.Rules))
	}
	if ss.Rules[0].Media != "" || ss.Rules[2].Media != "" {
		t.Errorf("rules outside @media must have empty Media: %+v", ss.Rules)
	}
	if ss.Rules[1].Media != "(display-mode: terminal)" {
		t.Errorf("Media = %q", ss.Rules[1].Media)
	}
	if ss.Rules[1].Properties["border"] != "single" {
		t.Errorf("media rule lost properties: %+v", ss.Rules[1])
	}
}

func TestMediaBlockWithMultipleRules(t *testing.T) {
	ss, err := Parse(`@media (display-mode: browser) { .a { color: red; } .b { color: blue; } }`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ss.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(ss.Rules))
	}
	for _, r := range ss.Rules {
		if r.Media != "(display-mode: browser)" {
			t.Errorf("Media = %q", r.Media)
		}
	}
}

func TestMediaBlockWithKeyframesUntouched(t *testing.T) {
	// @keyframes still parses outside media blocks alongside them.
	ss, err := Parse(`@media (display-mode: terminal) { .a { color: red; } } @keyframes pulse { from { color: red; } to { color: blue; } }`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ss.Rules) != 1 || len(ss.Keyframes) != 1 {
		t.Errorf("rules=%d keyframes=%d", len(ss.Rules), len(ss.Keyframes))
	}
}

func TestUnterminatedMediaBlockErrors(t *testing.T) {
	if _, err := Parse(`@media (display-mode: terminal) { .a { color: red; }`); err == nil {
		t.Error("unterminated @media should error")
	}
}
