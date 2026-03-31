package anim

import (
	"math"
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestEngineNoTransitionPassthrough(t *testing.T) {
	// Given an engine and a style with no transitions
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	style := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}

	// When
	got := engine.BeforeRender("node0", style, nil)

	// Then style passes through unchanged
	if got.FG.R != 255 || got.FG.G != 0 {
		t.Errorf("expected passthrough, got FG=(%d,%d,%d)", got.FG.R, got.FG.G, got.FG.B)
	}
}

func TestEngineTransitionStartsOnStyleChange(t *testing.T) {
	// Given a node with a color transition
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "color", DurationMs: 200, TimingFunction: Linear},
	}

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	// First render: establishes baseline
	engine.BeforeRender("node0", red, transitions)

	// Style changes to blue
	clock.Advance(16)
	got := engine.BeforeRender("node0", blue, transitions)

	// Should still be near red (transition just started, t≈0)
	if got.FG.R < 200 {
		t.Errorf("expected near-red at start, got R=%d", got.FG.R)
	}
	if !engine.HasActive() {
		t.Error("expected HasActive() = true")
	}
}

func TestEngineTransitionMidpoint(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "color", DurationMs: 200, TimingFunction: Linear},
	}

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	// Establish baseline
	engine.BeforeRender("node0", red, transitions)

	// Change to blue at t=0
	clock.Advance(1)
	engine.BeforeRender("node0", blue, transitions)

	// Advance to midpoint
	clock.Advance(100)
	got := engine.BeforeRender("node0", blue, transitions)

	// At t=0.5 linear: R should be ~128, B should be ~128
	if math.Abs(float64(got.FG.R)-128) > 5 {
		t.Errorf("at midpoint R=%d, want ~128", got.FG.R)
	}
	if math.Abs(float64(got.FG.B)-128) > 5 {
		t.Errorf("at midpoint B=%d, want ~128", got.FG.B)
	}
}

func TestEngineTransitionCompletes(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "color", DurationMs: 200, TimingFunction: Linear},
	}

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	engine.BeforeRender("node0", red, transitions)
	clock.Advance(1)
	engine.BeforeRender("node0", blue, transitions)

	// Advance past duration
	clock.Advance(250)
	got := engine.BeforeRender("node0", blue, transitions)

	// Should be blue
	if got.FG.R != 0 || got.FG.B != 255 {
		t.Errorf("after completion: FG=(%d,%d,%d), want (0,0,255)", got.FG.R, got.FG.G, got.FG.B)
	}
	if engine.HasActive() {
		t.Error("expected HasActive() = false after completion")
	}
}

func TestEngineTransitionDelay(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "color", DurationMs: 200, TimingFunction: Linear, DelayMs: 100},
	}

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	engine.BeforeRender("node0", red, transitions)
	clock.Advance(1)
	engine.BeforeRender("node0", blue, transitions)

	// During delay: should still be red
	clock.Advance(50)
	got := engine.BeforeRender("node0", blue, transitions)
	if got.FG.R != 255 {
		t.Errorf("during delay R=%d, want 255", got.FG.R)
	}

	// After delay + half duration: should be midpoint
	clock.Advance(150) // now at 201ms total: 100ms delay + 100ms into 200ms duration
	got = engine.BeforeRender("node0", blue, transitions)
	if math.Abs(float64(got.FG.R)-128) > 5 {
		t.Errorf("at delay+midpoint R=%d, want ~128", got.FG.R)
	}
}

func TestEngineMultiplePropertyTransitions(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "color", DurationMs: 200, TimingFunction: Linear},
		{Property: "background", DurationMs: 400, TimingFunction: Linear},
	}

	from := render.Style{
		FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0},
		BG: render.Color{IsRGB: true, R: 0, G: 0, B: 0},
	}
	to := render.Style{
		FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255},
		BG: render.Color{IsRGB: true, R: 255, G: 255, B: 255},
	}

	engine.BeforeRender("node0", from, transitions)
	clock.Advance(1)
	engine.BeforeRender("node0", to, transitions)

	// At 200ms: FG done, BG halfway.
	clock.Advance(200)
	got := engine.BeforeRender("node0", to, transitions)
	if got.FG.R != 0 || got.FG.B != 255 {
		t.Errorf("at 200ms FG=(%d,%d,%d), want (0,0,255)", got.FG.R, got.FG.G, got.FG.B)
	}
	if math.Abs(float64(got.BG.R)-128) > 5 {
		t.Errorf("at 200ms BG.R=%d, want ~128", got.BG.R)
	}

	// At 400ms: both done.
	clock.Advance(200)
	got = engine.BeforeRender("node0", to, transitions)
	if got.BG.R != 255 || got.BG.G != 255 || got.BG.B != 255 {
		t.Errorf("at 400ms BG=(%d,%d,%d), want (255,255,255)", got.BG.R, got.BG.G, got.BG.B)
	}
	if engine.HasActive() {
		t.Error("expected no active transitions after both complete")
	}
}

func TestEngineTransitionAll(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "all", DurationMs: 200, TimingFunction: Linear},
	}

	from := render.Style{
		FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0},
		BG: render.Color{IsRGB: true, R: 0, G: 0, B: 0},
	}
	to := render.Style{
		FG: render.Color{IsRGB: true, R: 0, G: 255, B: 0},
		BG: render.Color{IsRGB: true, R: 255, G: 255, B: 255},
	}

	engine.BeforeRender("node0", from, transitions)
	clock.Advance(1)
	engine.BeforeRender("node0", to, transitions)

	clock.Advance(100)
	got := engine.BeforeRender("node0", to, transitions)
	// Both FG and BG should be interpolated at midpoint.
	if math.Abs(float64(got.FG.G)-128) > 5 {
		t.Errorf("at midpoint FG.G=%d, want ~128", got.FG.G)
	}
	if math.Abs(float64(got.BG.R)-128) > 5 {
		t.Errorf("at midpoint BG.R=%d, want ~128", got.BG.R)
	}
}

func TestEngineNoStyleChangeNoTransition(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	transitions := []TransitionSpec{
		{Property: "color", DurationMs: 200, TimingFunction: Linear},
	}

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}

	engine.BeforeRender("node0", red, transitions)
	clock.Advance(16)
	got := engine.BeforeRender("node0", red, transitions)

	// Same style, no transition should start
	if got.FG.R != 255 {
		t.Errorf("same style should passthrough, got R=%d", got.FG.R)
	}
	if engine.HasActive() {
		t.Error("no style change should mean no active transitions")
	}
}
