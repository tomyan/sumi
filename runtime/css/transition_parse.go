package css

import (
	"math"
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/anim"
)

// ParseTransitions extracts transition specs from CSS properties.
// Supports shorthand: transition: color 200ms ease-out 50ms
// Supports multiple: transition: color 200ms, background 500ms
// Supports longhand: transition-property, transition-duration, etc.
func ParseTransitions(props map[string]string) []anim.TransitionSpec {
	if shorthand, ok := props["transition"]; ok {
		return parseTransitionShorthand(shorthand)
	}
	// Try longhand properties.
	prop, ok := props["transition-property"]
	if !ok {
		return nil
	}
	dur := parseDuration(props["transition-duration"])
	tf := anim.Ease
	if tfStr, ok := props["transition-timing-function"]; ok {
		if parsed, err := anim.ParseTimingFunction(tfStr); err == nil {
			tf = parsed
		}
	}
	delay := parseDuration(props["transition-delay"])
	return []anim.TransitionSpec{{
		Property:       strings.TrimSpace(prop),
		DurationMs:     dur,
		TimingFunction: tf,
		DelayMs:        delay,
	}}
}

func parseTransitionShorthand(s string) []anim.TransitionSpec {
	// Split on commas for multiple transitions.
	parts := strings.Split(s, ",")
	var specs []anim.TransitionSpec
	for _, part := range parts {
		if spec, ok := parseSingleTransition(strings.TrimSpace(part)); ok {
			specs = append(specs, spec)
		}
	}
	return specs
}

func parseSingleTransition(s string) (anim.TransitionSpec, bool) {
	tokens := strings.Fields(s)
	if len(tokens) < 2 {
		return anim.TransitionSpec{}, false
	}

	spec := anim.TransitionSpec{
		Property:       tokens[0],
		TimingFunction: anim.Ease, // CSS default
	}

	// Second token is always duration.
	spec.DurationMs = parseDuration(tokens[1])

	// Remaining tokens: optional timing function and delay.
	for i := 2; i < len(tokens); i++ {
		if tf, err := anim.ParseTimingFunction(tokens[i]); err == nil {
			spec.TimingFunction = tf
		} else {
			// Assume it's a delay.
			spec.DelayMs = parseDuration(tokens[i])
		}
	}

	return spec, true
}

// parseDuration parses a CSS duration string like "200ms" or "0.5s".
func parseDuration(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "ms") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
		if err != nil {
			return 0
		}
		return int(math.Round(v))
	}
	if strings.HasSuffix(s, "s") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
		if err != nil {
			return 0
		}
		return int(math.Round(v * 1000))
	}
	// Bare number — assume ms.
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int(math.Round(v))
}
