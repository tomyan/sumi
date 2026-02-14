package sumitest

import "github.com/tomyan/sumi/runtime/tui"

// Scenario defines a sequence of interaction steps against a component.
type Scenario struct {
	Name       string
	Width      int
	Height     int
	NewApp     func(w, h int) *tui.App
	Steps      []Step
	SourceFile string // optional path to .sumi source shown in preview
}

// Step is a single named interaction within a scenario.
// A nil Action captures the current frame without dispatching any event.
type Step struct {
	Name   string
	Action func(h *Harness)
}

// Frame holds the captured output for a single step.
type Frame struct {
	Name      string
	StyledText string
}
