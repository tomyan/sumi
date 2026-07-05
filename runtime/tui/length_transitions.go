package tui

import (
	"github.com/tomyan/sumi/runtime/anim"
	"github.com/tomyan/sumi/runtime/layout"
)

// stepLengthTransitions interpolates width/height for nodes whose CSS
// transitions cover them, between style resolution (which stamps the
// target values) and layout. Returns true while any transition runs so
// the caller keeps scheduling frames. Only CSS-driven lengths transition:
// the resolver re-stamps the target each pass.
func stepLengthTransitions(root *layout.Input, engine *anim.Engine) bool {
	if root == nil {
		return false
	}
	active := false
	var walk func(n *layout.Input)
	walk = func(n *layout.Input) {
		if n == nil {
			return
		}
		if len(n.Transitions) > 0 {
			if n.LengthAnim == nil {
				n.LengthAnim = &anim.LengthState{}
			}
			var running bool
			if n.FixedWidth > 0 {
				n.FixedWidth, running = engine.StepLength(n.LengthAnim, "width", n.FixedWidth, n.Transitions)
				active = active || running
			}
			if n.FixedHeight > 0 {
				n.FixedHeight, running = engine.StepLength(n.LengthAnim, "height", n.FixedHeight, n.Transitions)
				active = active || running
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return active
}
