package style

import "testing"

// RS1: Serialize round-trips a stylesheet through CSS text so codegen can
// embed it in generated Go for runtime resolution.

func roundTrip(t *testing.T, src string) *Stylesheet {
	t.Helper()
	ss, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := Parse(Serialize(ss))
	if err != nil {
		t.Fatalf("reparse serialized form: %v\n%s", err, Serialize(ss))
	}
	return out
}

func TestSerializeRoundTripsRules(t *testing.T) {
	// Given / When
	ss := roundTrip(t, `.title { color: red; font-weight: bold; } text { opacity: dim; }`)

	// Then
	if len(ss.Rules) != 2 {
		t.Fatalf("rules = %d, want 2", len(ss.Rules))
	}
	if ss.Rules[0].Properties["color"] != "red" || ss.Rules[0].Selector != ".title" {
		t.Errorf("rule 0 = %+v", ss.Rules[0])
	}
}

func TestSerializeRoundTripsPseudoAndCombinators(t *testing.T) {
	ss := roundTrip(t, `.panel > text:hover { color: cyan; } .a + .b:focus { color: red; }`)
	if ss.Rules[0].Pseudo != "hover" || len(ss.Rules[0].Parsed.Parts) != 2 {
		t.Errorf("rule 0 = %+v", ss.Rules[0])
	}
	if ss.Rules[1].Pseudo != "focus" || ss.Rules[1].Parsed.Combinators[0] != '+' {
		t.Errorf("rule 1 = %+v", ss.Rules[1])
	}
}

func TestSerializeRoundTripsMedia(t *testing.T) {
	ss := roundTrip(t, `
.card { color: red; }
@media (display-mode: terminal) { .card { border: single; } .other { color: blue; } }
`)
	if len(ss.Rules) != 3 {
		t.Fatalf("rules = %d, want 3", len(ss.Rules))
	}
	if ss.Rules[1].Media != "(display-mode: terminal)" || ss.Rules[2].Media != "(display-mode: terminal)" {
		t.Errorf("media lost: %+v", ss.Rules)
	}
}

func TestSerializeRoundTripsKeyframes(t *testing.T) {
	ss := roundTrip(t, `@keyframes pulse { from { color: red; } 50% { color: yellow; } to { color: blue; } }`)
	if len(ss.Keyframes) != 1 || len(ss.Keyframes[0].Stops) != 3 {
		t.Fatalf("keyframes = %+v", ss.Keyframes)
	}
	if ss.Keyframes[0].Stops[1].Percent != 0.5 {
		t.Errorf("mid stop = %+v", ss.Keyframes[0].Stops[1])
	}
}
