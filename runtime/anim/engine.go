package anim

import "github.com/tomyan/sumi/runtime/render"

// Engine manages animation state for all nodes. It intercepts style
// changes and produces interpolated styles when transitions are active.
type Engine struct {
	clock        Clock
	nodes        map[string]*nodeState
	requestFrame func()
}

type nodeState struct {
	prevStyle   render.Style
	transitions []activeTransition
	initialized bool
}

type activeTransition struct {
	property string
	from     render.Style
	to       render.Style
	startMs  int64
	duration int
	delay    int
	timing   TimingFunction
}

// NewEngine creates an animation engine with the given clock and frame requester.
func NewEngine(clock Clock, requestFrame func()) *Engine {
	return &Engine{
		clock:        clock,
		nodes:        make(map[string]*nodeState),
		requestFrame: requestFrame,
	}
}

// BeforeRender processes a node's current style against its previous state.
// If a transition is configured and the style changed, it starts interpolating.
// Returns the style to use for rendering (interpolated or current).
func (e *Engine) BeforeRender(nodeID string, currentStyle render.Style, transitions []TransitionSpec) render.Style {
	ns := e.nodes[nodeID]
	if ns == nil {
		ns = &nodeState{}
		e.nodes[nodeID] = ns
	}

	// First render: just record baseline.
	if !ns.initialized {
		ns.prevStyle = currentStyle
		ns.initialized = true
		return currentStyle
	}

	// Check for style changes and start transitions.
	if currentStyle != ns.prevStyle && len(transitions) > 0 {
		now := e.clock.Now()
		for _, spec := range transitions {
			if !propertyChanged(spec.Property, ns.prevStyle, currentStyle) {
				continue
			}
			// Replace any existing transition for this property.
			ns.transitions = replaceTransition(ns.transitions, activeTransition{
				property: spec.Property,
				from:     ns.prevStyle,
				to:       currentStyle,
				startMs:  now,
				duration: spec.DurationMs,
				delay:    spec.DelayMs,
				timing:   spec.TimingFunction,
			})
		}
		ns.prevStyle = currentStyle
	}

	// Evaluate active transitions.
	if len(ns.transitions) == 0 {
		return currentStyle
	}

	now := e.clock.Now()
	result := currentStyle
	var active []activeTransition
	for _, at := range ns.transitions {
		elapsed := now - at.startMs
		if elapsed < int64(at.delay) {
			// Still in delay period: use from style.
			result = applyTransitionProperty(result, at.property, at.from)
			active = append(active, at)
			continue
		}
		progress := float64(elapsed-int64(at.delay)) / float64(at.duration)
		if progress >= 1.0 {
			// Transition complete: use to style (already currentStyle).
			continue
		}
		t := at.timing.Evaluate(progress)
		result = applyTransitionLerp(result, at.property, at.from, at.to, t)
		active = append(active, at)
	}
	ns.transitions = active
	return result
}

// HasActive returns true if any node has running transitions.
func (e *Engine) HasActive() bool {
	for _, ns := range e.nodes {
		if len(ns.transitions) > 0 {
			return true
		}
	}
	return false
}

// propertyChanged checks if the given property differs between old and new styles.
func propertyChanged(property string, old, new render.Style) bool {
	switch property {
	case "color":
		return old.FG != new.FG
	case "background":
		return old.BG != new.BG
	case "all":
		return old != new
	default:
		return old != new
	}
}

// applyTransitionProperty replaces a specific property in result with the source style's value.
func applyTransitionProperty(result render.Style, property string, source render.Style) render.Style {
	switch property {
	case "color":
		result.FG = source.FG
	case "background":
		result.BG = source.BG
	case "all":
		return source
	}
	return result
}

// applyTransitionLerp interpolates a specific property between from and to at parameter t.
func applyTransitionLerp(result render.Style, property string, from, to render.Style, t float64) render.Style {
	switch property {
	case "color":
		result.FG = render.LerpColor(from.FG, to.FG, t)
	case "background":
		result.BG = render.LerpColor(from.BG, to.BG, t)
	case "all":
		return LerpStyle(from, to, t)
	}
	return result
}

// replaceTransition replaces an existing transition for the same property, or appends.
func replaceTransition(transitions []activeTransition, at activeTransition) []activeTransition {
	for i, existing := range transitions {
		if existing.property == at.property {
			transitions[i] = at
			return transitions
		}
	}
	return append(transitions, at)
}
