package anim

import "testing"

// E1: steps() easing.

func TestParseSteps(t *testing.T) {
	tf, err := ParseTimingFunction("steps(4)")
	if err != nil || tf.Steps != 4 || tf.JumpStart {
		t.Fatalf("steps(4) = %+v (err %v)", tf, err)
	}
	tf, err = ParseTimingFunction("steps(3, start)")
	if err != nil || tf.Steps != 3 || !tf.JumpStart {
		t.Fatalf("steps(3, start) = %+v (err %v)", tf, err)
	}
	if _, err := ParseTimingFunction("steps(0)"); err == nil {
		t.Error("steps(0) should error")
	}
}

func TestStepEndEvaluation(t *testing.T) {
	tf, _ := ParseTimingFunction("steps(4)")
	cases := []struct{ t, want float64 }{
		{0.0, 0}, {0.1, 0}, {0.26, 0.25}, {0.6, 0.5}, {0.99, 0.75}, {1.0, 1},
	}
	for _, c := range cases {
		if got := tf.Evaluate(c.t); got != c.want {
			t.Errorf("steps(4).Evaluate(%v) = %v, want %v", c.t, got, c.want)
		}
	}
}

func TestStepStartEvaluation(t *testing.T) {
	tf, _ := ParseTimingFunction("step-start")
	if got := tf.Evaluate(0.01); got != 1 {
		t.Errorf("step-start at 0.01 = %v, want 1", got)
	}
	tf, _ = ParseTimingFunction("step-end")
	if got := tf.Evaluate(0.99); got != 0 {
		t.Errorf("step-end at 0.99 = %v, want 0", got)
	}
}
