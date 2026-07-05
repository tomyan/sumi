package anim

import "math"

// LengthState tracks in-flight whole-cell transitions for one node's
// layout lengths. The runtime keeps one per element (build-once inputs
// are stable), so no global registry is needed.
type LengthState struct {
	props map[string]*lengthTransition
}

type lengthTransition struct {
	target   int // last target value seen
	seen     bool
	from, to int
	startMs  int64
	duration int
	delay    int
	timing   TimingFunction
	active   bool
}

// StepLength returns the value to lay out with for a length property.
// When the target changes and a transition spec covers the property, it
// interpolates from the currently displayed value in whole cells.
// The second return reports whether a transition is still running (the
// caller schedules another frame while true).
func (e *Engine) StepLength(state *LengthState, property string, target int, specs []TransitionSpec) (int, bool) {
	if state.props == nil {
		state.props = map[string]*lengthTransition{}
	}
	lt := state.props[property]
	if lt == nil {
		lt = &lengthTransition{}
		state.props[property] = lt
	}
	now := e.clock.Now()

	if !lt.seen {
		lt.seen = true
		lt.target = target
		return target, false
	}

	if target != lt.target {
		spec := lengthSpecFor(property, specs)
		if spec == nil {
			lt.target = target
			lt.active = false
			return target, false
		}
		lt.from = e.currentLength(lt, now, lt.target)
		lt.to = target
		lt.target = target
		lt.startMs = now
		lt.duration = spec.DurationMs
		lt.delay = spec.DelayMs
		lt.timing = spec.TimingFunction
		lt.active = true
	}

	if !lt.active {
		return target, false
	}
	value := e.currentLength(lt, now, target)
	if value == lt.to && now-lt.startMs >= int64(lt.delay+lt.duration) {
		lt.active = false
		return target, false
	}
	return value, true
}

// currentLength evaluates a transition's displayed value at time now.
func (e *Engine) currentLength(lt *lengthTransition, now int64, settled int) int {
	if !lt.active {
		return settled
	}
	elapsed := now - lt.startMs - int64(lt.delay)
	if elapsed < 0 {
		return lt.from
	}
	if lt.duration <= 0 || elapsed >= int64(lt.duration) {
		return lt.to
	}
	t := lt.timing.Evaluate(float64(elapsed) / float64(lt.duration))
	return int(math.Round(float64(lt.from) + float64(lt.to-lt.from)*t))
}

// lengthSpecFor finds the transition spec covering a length property.
func lengthSpecFor(property string, specs []TransitionSpec) *TransitionSpec {
	for i := range specs {
		if specs[i].Property == property || specs[i].Property == "all" {
			return &specs[i]
		}
	}
	return nil
}
