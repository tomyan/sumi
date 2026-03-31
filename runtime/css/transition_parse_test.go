package css

import (
	"testing"

	"github.com/tomyan/sumi/runtime/anim"
)

func TestParseTransitionShorthand(t *testing.T) {
	props := map[string]string{
		"transition": "color 200ms ease-out",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	s := specs[0]
	if s.Property != "color" {
		t.Errorf("Property = %q, want %q", s.Property, "color")
	}
	if s.DurationMs != 200 {
		t.Errorf("DurationMs = %d, want 200", s.DurationMs)
	}
	if s.TimingFunction != anim.EaseOut {
		t.Errorf("TimingFunction = %v, want EaseOut", s.TimingFunction)
	}
	if s.DelayMs != 0 {
		t.Errorf("DelayMs = %d, want 0", s.DelayMs)
	}
}

func TestParseTransitionWithDelay(t *testing.T) {
	props := map[string]string{
		"transition": "color 200ms ease-out 50ms",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].DelayMs != 50 {
		t.Errorf("DelayMs = %d, want 50", specs[0].DelayMs)
	}
}

func TestParseTransitionMultiple(t *testing.T) {
	props := map[string]string{
		"transition": "color 200ms ease-out, background 500ms ease-in",
	}
	specs := ParseTransitions(props)
	if len(specs) != 2 {
		t.Fatalf("got %d specs, want 2", len(specs))
	}
	if specs[0].Property != "color" || specs[0].DurationMs != 200 {
		t.Errorf("spec[0] = %+v", specs[0])
	}
	if specs[1].Property != "background" || specs[1].DurationMs != 500 {
		t.Errorf("spec[1] = %+v", specs[1])
	}
}

func TestParseTransitionAll(t *testing.T) {
	props := map[string]string{
		"transition": "all 300ms linear",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].Property != "all" {
		t.Errorf("Property = %q, want %q", specs[0].Property, "all")
	}
	if specs[0].TimingFunction != anim.Linear {
		t.Errorf("TimingFunction = %v, want Linear", specs[0].TimingFunction)
	}
}

func TestParseTransitionDurationSeconds(t *testing.T) {
	props := map[string]string{
		"transition": "color 0.5s ease",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].DurationMs != 500 {
		t.Errorf("DurationMs = %d, want 500", specs[0].DurationMs)
	}
}

func TestParseTransitionDurationSecondsWholeNumber(t *testing.T) {
	props := map[string]string{
		"transition": "color 2s ease",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].DurationMs != 2000 {
		t.Errorf("DurationMs = %d, want 2000", specs[0].DurationMs)
	}
}

func TestParseTransitionDefaultEasing(t *testing.T) {
	// No easing specified — default to ease.
	props := map[string]string{
		"transition": "color 200ms",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	if specs[0].TimingFunction != anim.Ease {
		t.Errorf("TimingFunction = %v, want Ease (default)", specs[0].TimingFunction)
	}
}

func TestParseTransitionEmpty(t *testing.T) {
	specs := ParseTransitions(map[string]string{})
	if len(specs) != 0 {
		t.Errorf("got %d specs for empty props, want 0", len(specs))
	}
}

func TestParseTransitionLonghand(t *testing.T) {
	props := map[string]string{
		"transition-property":        "color",
		"transition-duration":        "200ms",
		"transition-timing-function": "ease-in",
		"transition-delay":           "100ms",
	}
	specs := ParseTransitions(props)
	if len(specs) != 1 {
		t.Fatalf("got %d specs, want 1", len(specs))
	}
	s := specs[0]
	if s.Property != "color" || s.DurationMs != 200 || s.TimingFunction != anim.EaseIn || s.DelayMs != 100 {
		t.Errorf("spec = %+v", s)
	}
}
