package anim

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// TimingFunction defines an easing curve via cubic bezier control points.
type TimingFunction struct {
	Name               string
	X1, Y1, X2, Y2    float64
}

// Named timing function presets (CSS standard values).
var (
	Linear    = TimingFunction{Name: "linear", X1: 0, Y1: 0, X2: 1, Y2: 1}
	Ease      = TimingFunction{Name: "ease", X1: 0.25, Y1: 0.1, X2: 0.25, Y2: 1.0}
	EaseIn    = TimingFunction{Name: "ease-in", X1: 0.42, Y1: 0, X2: 1.0, Y2: 1.0}
	EaseOut   = TimingFunction{Name: "ease-out", X1: 0, Y1: 0, X2: 0.58, Y2: 1.0}
	EaseInOut = TimingFunction{Name: "ease-in-out", X1: 0.42, Y1: 0, X2: 0.58, Y2: 1.0}
)

// Evaluate returns the eased value for a progress t in [0, 1].
func (tf TimingFunction) Evaluate(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	// Linear special case.
	if tf.X1 == 0 && tf.Y1 == 0 && tf.X2 == 1 && tf.Y2 == 1 {
		return t
	}
	return cubicBezierY(tf.X1, tf.Y1, tf.X2, tf.Y2, t)
}

// cubicBezierY solves the cubic bezier curve for x=t and returns the y value.
// Uses Newton-Raphson iteration to find the parameter u where bezierX(u) = t,
// then returns bezierY(u).
func cubicBezierY(x1, y1, x2, y2, t float64) float64 {
	// Find u such that bezierX(u) = t using Newton-Raphson.
	u := t // initial guess
	for i := 0; i < 8; i++ {
		x := bezierValue(x1, x2, u) - t
		if math.Abs(x) < 1e-7 {
			break
		}
		dx := bezierDerivative(x1, x2, u)
		if math.Abs(dx) < 1e-7 {
			break
		}
		u -= x / dx
	}
	// Clamp u to [0, 1].
	if u < 0 {
		u = 0
	}
	if u > 1 {
		u = 1
	}
	return bezierValue(y1, y2, u)
}

// bezierValue evaluates the cubic bezier polynomial B(u) for one axis.
// B(u) = 3(1-u)²u·p1 + 3(1-u)u²·p2 + u³
func bezierValue(p1, p2, u float64) float64 {
	inv := 1 - u
	return 3*inv*inv*u*p1 + 3*inv*u*u*p2 + u*u*u
}

// bezierDerivative returns dB/du for one axis.
// B'(u) = 3(1-u)²·p1 + 6(1-u)u·(p2-p1) + 3u²·(1-p2)
func bezierDerivative(p1, p2, u float64) float64 {
	inv := 1 - u
	return 3*inv*inv*p1 + 6*inv*u*(p2-p1) + 3*u*u*(1-p2)
}

// ParseTimingFunction parses a CSS timing function string.
func ParseTimingFunction(s string) (TimingFunction, error) {
	s = strings.TrimSpace(s)
	switch s {
	case "linear":
		return Linear, nil
	case "ease":
		return Ease, nil
	case "ease-in":
		return EaseIn, nil
	case "ease-out":
		return EaseOut, nil
	case "ease-in-out":
		return EaseInOut, nil
	}
	if strings.HasPrefix(s, "cubic-bezier(") && strings.HasSuffix(s, ")") {
		inner := s[len("cubic-bezier(") : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) != 4 {
			return TimingFunction{}, fmt.Errorf("cubic-bezier requires 4 values, got %d", len(parts))
		}
		vals := make([]float64, 4)
		for i, p := range parts {
			v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				return TimingFunction{}, fmt.Errorf("invalid cubic-bezier value %q: %w", p, err)
			}
			vals[i] = v
		}
		return TimingFunction{
			Name: s,
			X1:   vals[0], Y1: vals[1],
			X2:   vals[2], Y2: vals[3],
		}, nil
	}
	return TimingFunction{}, fmt.Errorf("unknown timing function: %q", s)
}
