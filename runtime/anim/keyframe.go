package anim

import (
	"sort"

	"github.com/tomyan/sumi/runtime/render"
)

// KeyframeAnimation defines a named sequence of style stops.
type KeyframeAnimation struct {
	Name  string
	Stops []KeyframeStop
}

// KeyframeStop is a single percentage stop with a resolved style.
type KeyframeStop struct {
	Percent float64      // 0.0 to 1.0
	Style   render.Style // resolved style at this stop
}

// RegisterKeyframes adds a keyframe animation to the engine.
func (e *Engine) RegisterKeyframes(name string, kf *KeyframeAnimation) {
	if e.keyframes == nil {
		e.keyframes = make(map[string]*KeyframeAnimation)
	}
	e.keyframes[name] = kf
}

// evaluateKeyframe computes the interpolated style for a keyframe animation at the given time.
func (e *Engine) evaluateKeyframe(spec *AnimationSpec, nodeID string, baseStyle render.Style) render.Style {
	kf, ok := e.keyframes[spec.Name]
	if !ok || len(kf.Stops) == 0 {
		return baseStyle
	}

	ns := e.nodes[nodeID]
	if ns == nil {
		ns = &nodeState{}
		e.nodes[nodeID] = ns
	}
	if !ns.animStarted {
		ns.animStarted = true
		ns.animStart = e.clock.Now()
		ns.animSpec = spec
	}

	elapsed := e.clock.Now() - ns.animStart - int64(spec.DelayMs)
	if elapsed < 0 {
		// In delay period.
		if spec.FillMode == "backwards" || spec.FillMode == "both" {
			return kf.Stops[0].Style
		}
		return baseStyle
	}

	dur := int64(spec.DurationMs)
	if dur <= 0 {
		return baseStyle
	}

	// Determine iteration and position within cycle.
	iteration := elapsed / dur
	cyclePos := float64(elapsed%dur) / float64(dur)

	// Check if animation is complete.
	if spec.IterationCount >= 0 && iteration >= int64(spec.IterationCount) {
		ns.animDone = true
		if spec.FillMode == "forwards" || spec.FillMode == "both" {
			return kf.Stops[len(kf.Stops)-1].Style
		}
		return baseStyle
	}

	// Apply direction.
	reverse := false
	switch spec.Direction {
	case "reverse":
		reverse = true
	case "alternate":
		reverse = iteration%2 == 1
	case "alternate-reverse":
		reverse = iteration%2 == 0
	}
	if reverse {
		cyclePos = 1.0 - cyclePos
	}

	return interpolateKeyframes(kf.Stops, cyclePos, spec.TimingFunction)
}

// interpolateKeyframes finds the two surrounding stops and lerps between them.
func interpolateKeyframes(stops []KeyframeStop, pos float64, timing TimingFunction) render.Style {
	if len(stops) == 0 {
		return render.Style{}
	}
	if pos <= stops[0].Percent {
		return stops[0].Style
	}
	if pos >= stops[len(stops)-1].Percent {
		return stops[len(stops)-1].Style
	}

	// Find the segment.
	idx := sort.Search(len(stops), func(i int) bool {
		return stops[i].Percent > pos
	})
	if idx == 0 {
		return stops[0].Style
	}

	from := stops[idx-1]
	to := stops[idx]
	segLen := to.Percent - from.Percent
	if segLen <= 0 {
		return from.Style
	}
	t := (pos - from.Percent) / segLen
	t = timing.Evaluate(t)

	return LerpStyle(from.Style, to.Style, t)
}
