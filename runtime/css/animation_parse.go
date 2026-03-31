package css

import (
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/anim"
)

// ParseAnimation extracts an animation spec from CSS properties.
// Supports shorthand: animation: name duration [timing] [delay] [iteration] [direction] [fill-mode]
// Supports longhand: animation-name, animation-duration, etc.
func ParseAnimation(props map[string]string) *anim.AnimationSpec {
	if shorthand, ok := props["animation"]; ok {
		return parseAnimationShorthand(shorthand)
	}
	name, ok := props["animation-name"]
	if !ok {
		return nil
	}
	spec := &anim.AnimationSpec{
		Name:           name,
		DurationMs:     parseDuration(props["animation-duration"]),
		TimingFunction: anim.Ease,
		DelayMs:        parseDuration(props["animation-delay"]),
		IterationCount: 1,
		Direction:      "normal",
		FillMode:       "none",
		PlayState:      "running",
	}
	if tf, ok := props["animation-timing-function"]; ok {
		if parsed, err := anim.ParseTimingFunction(tf); err == nil {
			spec.TimingFunction = parsed
		}
	}
	if ic, ok := props["animation-iteration-count"]; ok {
		spec.IterationCount = parseIterationCount(ic)
	}
	if d, ok := props["animation-direction"]; ok {
		spec.Direction = d
	}
	if fm, ok := props["animation-fill-mode"]; ok {
		spec.FillMode = fm
	}
	if ps, ok := props["animation-play-state"]; ok {
		spec.PlayState = ps
	}
	return spec
}

func parseAnimationShorthand(s string) *anim.AnimationSpec {
	tokens := strings.Fields(s)
	if len(tokens) < 2 {
		return nil
	}

	spec := &anim.AnimationSpec{
		Name:           tokens[0],
		TimingFunction: anim.Ease,
		IterationCount: 1,
		Direction:      "normal",
		FillMode:       "none",
		PlayState:      "running",
	}

	// Second token is always duration.
	spec.DurationMs = parseDuration(tokens[1])

	// Parse remaining tokens by type.
	for i := 2; i < len(tokens); i++ {
		tok := tokens[i]
		if tf, err := anim.ParseTimingFunction(tok); err == nil {
			spec.TimingFunction = tf
			continue
		}
		if tok == "infinite" {
			spec.IterationCount = -1
			continue
		}
		if isDirection(tok) {
			spec.Direction = tok
			continue
		}
		if isFillMode(tok) {
			spec.FillMode = tok
			continue
		}
		if tok == "running" || tok == "paused" {
			spec.PlayState = tok
			continue
		}
		// Try as iteration count.
		if n, err := strconv.Atoi(tok); err == nil {
			spec.IterationCount = n
			continue
		}
		// Try as delay (duration format).
		if isDuration(tok) {
			spec.DelayMs = parseDuration(tok)
		}
	}

	return spec
}

func parseIterationCount(s string) int {
	s = strings.TrimSpace(s)
	if s == "infinite" {
		return -1
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 1
	}
	return n
}

func isDirection(s string) bool {
	switch s {
	case "normal", "reverse", "alternate", "alternate-reverse":
		return true
	}
	return false
}

func isFillMode(s string) bool {
	switch s {
	case "none", "forwards", "backwards", "both":
		return true
	}
	return false
}

func isDuration(s string) bool {
	return strings.HasSuffix(s, "ms") || strings.HasSuffix(s, "s")
}
