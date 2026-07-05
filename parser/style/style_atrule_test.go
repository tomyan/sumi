package style

import "testing"

// A2: unknown at-rules parse and drop silently (never a parse error).
// @keyframes remains the only consumed at-rule until @media lands (A10).

func TestUnknownAtRuleSkippedSilently(t *testing.T) {
	// Given
	input := `
@media (display-mode: browser) {
	.card { box-shadow: 0 2px 8px #0004; }
}
.card { color: red; }
`

	// When
	ss, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("unknown at-rule must not error, got: %v", err)
	}
	if len(ss.Rules) != 1 {
		t.Fatalf("expected 1 rule (the @media block skipped), got %d", len(ss.Rules))
	}
	if ss.Rules[0].Properties["color"] != "red" {
		t.Errorf("rule after skipped at-rule lost properties: %+v", ss.Rules[0])
	}
}

func TestUnknownAtRuleWithNestedBracesSkipped(t *testing.T) {
	// Given
	input := `
@supports (display: grid) {
	.a { color: blue; }
	.b { color: green; }
}
text { color: cyan; }
`

	// When
	ss, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("nested at-rule must not error, got: %v", err)
	}
	if len(ss.Rules) != 1 || ss.Rules[0].Selector != "text" {
		t.Fatalf("expected only the trailing text rule, got %+v", ss.Rules)
	}
}

func TestStatementAtRuleWithoutBlockSkipped(t *testing.T) {
	// Given: statement-style at-rule terminated by semicolon
	input := `@import "theme.css";
.x { color: yellow; }`

	// When
	ss, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("statement at-rule must not error, got: %v", err)
	}
	if len(ss.Rules) != 1 || ss.Rules[0].Selector != ".x" {
		t.Fatalf("expected only .x rule, got %+v", ss.Rules)
	}
}

func TestKeyframesStillParsed(t *testing.T) {
	// Given
	input := `@keyframes pulse { from { color: red; } to { color: blue; } }`

	// When
	ss, err := Parse(input)

	// Then
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ss.Keyframes) != 1 || ss.Keyframes[0].Name != "pulse" {
		t.Fatalf("@keyframes must still be consumed, got %+v", ss.Keyframes)
	}
}

func TestUnterminatedAtRuleErrors(t *testing.T) {
	// Given: an at-rule block that never closes
	input := `@media (min-width: 40) { .a { color: red; }`

	// When
	_, err := Parse(input)

	// Then
	if err == nil {
		t.Fatal("unterminated at-rule block should error")
	}
}
