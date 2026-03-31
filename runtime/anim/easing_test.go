package anim

import (
	"math"
	"testing"
)

func TestLinearEasing(t *testing.T) {
	tests := []struct{ t, want float64 }{
		{0.0, 0.0},
		{0.25, 0.25},
		{0.5, 0.5},
		{0.75, 0.75},
		{1.0, 1.0},
	}
	for _, tc := range tests {
		got := Linear.Evaluate(tc.t)
		if math.Abs(got-tc.want) > 0.001 {
			t.Errorf("Linear.Evaluate(%v) = %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestEasingBoundaryValues(t *testing.T) {
	// All easing functions must return 0 at t=0 and 1 at t=1.
	easings := []struct {
		name string
		fn   TimingFunction
	}{
		{"Linear", Linear},
		{"Ease", Ease},
		{"EaseIn", EaseIn},
		{"EaseOut", EaseOut},
		{"EaseInOut", EaseInOut},
	}
	for _, e := range easings {
		if got := e.fn.Evaluate(0.0); math.Abs(got) > 0.001 {
			t.Errorf("%s.Evaluate(0) = %v, want 0", e.name, got)
		}
		if got := e.fn.Evaluate(1.0); math.Abs(got-1.0) > 0.001 {
			t.Errorf("%s.Evaluate(1) = %v, want 1", e.name, got)
		}
	}
}

func TestEasingMonotonic(t *testing.T) {
	// All standard easings should be monotonically increasing.
	easings := []struct {
		name string
		fn   TimingFunction
	}{
		{"Linear", Linear},
		{"Ease", Ease},
		{"EaseIn", EaseIn},
		{"EaseOut", EaseOut},
		{"EaseInOut", EaseInOut},
	}
	for _, e := range easings {
		prev := 0.0
		for i := 1; i <= 100; i++ {
			tt := float64(i) / 100.0
			got := e.fn.Evaluate(tt)
			if got < prev-0.001 {
				t.Errorf("%s not monotonic: Evaluate(%v)=%v < previous %v", e.name, tt, got, prev)
				break
			}
			prev = got
		}
	}
}

func TestEaseInSlowerAtStart(t *testing.T) {
	// EaseIn should be below linear at t=0.25.
	if got := EaseIn.Evaluate(0.25); got >= 0.25 {
		t.Errorf("EaseIn.Evaluate(0.25) = %v, expected < 0.25", got)
	}
}

func TestEaseOutFasterAtStart(t *testing.T) {
	// EaseOut should be above linear at t=0.25.
	if got := EaseOut.Evaluate(0.25); got <= 0.25 {
		t.Errorf("EaseOut.Evaluate(0.25) = %v, expected > 0.25", got)
	}
}

func TestParseTimingFunction(t *testing.T) {
	tests := []struct {
		input string
		want  TimingFunction
	}{
		{"linear", Linear},
		{"ease", Ease},
		{"ease-in", EaseIn},
		{"ease-out", EaseOut},
		{"ease-in-out", EaseInOut},
	}
	for _, tc := range tests {
		got, err := ParseTimingFunction(tc.input)
		if err != nil {
			t.Errorf("ParseTimingFunction(%q) error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseTimingFunction(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseCubicBezier(t *testing.T) {
	got, err := ParseTimingFunction("cubic-bezier(0, 0, 1, 1)")
	if err != nil {
		t.Fatal(err)
	}
	// cubic-bezier(0,0,1,1) is linear.
	for _, tt := range []float64{0.25, 0.5, 0.75} {
		val := got.Evaluate(tt)
		if math.Abs(val-tt) > 0.01 {
			t.Errorf("cubic-bezier(0,0,1,1).Evaluate(%v) = %v, want ~%v", tt, val, tt)
		}
	}
}

func TestParseTimingFunctionError(t *testing.T) {
	_, err := ParseTimingFunction("bogus")
	if err == nil {
		t.Error("expected error for unknown timing function")
	}
}
