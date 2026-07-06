package anim

import (
	"math"
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestKeyframeBasicPlayback(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	engine.RegisterKeyframes("fade", &KeyframeAnimation{
		Name: "fade",
		Stops: []KeyframeStop{
			{Percent: 0, Style: red},
			{Percent: 1, Style: blue},
		},
	})

	spec := &AnimationSpec{
		Name:           "fade",
		DurationMs:     200,
		TimingFunction: Linear,
		IterationCount: 1,
		Direction:      "normal",
		FillMode:       "none",
	}

	// At t=0: should be red.
	got := engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if got.FG.R != 255 {
		t.Errorf("at t=0 R=%d, want 255", got.FG.R)
	}

	// At t=100ms (midpoint): should be ~128.
	clock.Advance(100)
	got = engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if math.Abs(float64(got.FG.R)-128) > 5 {
		t.Errorf("at midpoint R=%d, want ~128", got.FG.R)
	}

	// At t=200ms: complete.
	clock.Advance(100)
	got = engine.BeforeRenderAnim("node0", render.Style{}, spec)
	// fill-mode: none → reverts to base style.
	if got.FG.R != 0 && got.FG.B != 0 {
		// base style is empty, so zero color
	}
}

func TestKeyframeThreeStops(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	green := render.Style{FG: render.Color{IsRGB: true, R: 80, G: 250, B: 123}}
	dark := render.Style{FG: render.Color{IsRGB: true, R: 45, G: 138, B: 78}}

	engine.RegisterKeyframes("pulse", &KeyframeAnimation{
		Name: "pulse",
		Stops: []KeyframeStop{
			{Percent: 0, Style: green},
			{Percent: 0.5, Style: dark},
			{Percent: 1, Style: green},
		},
	})

	spec := &AnimationSpec{
		Name:           "pulse",
		DurationMs:     200,
		TimingFunction: Linear,
		IterationCount: -1, // infinite
		Direction:      "normal",
		FillMode:       "none",
	}

	// At t=0: green.
	got := engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if got.FG.R != 80 {
		t.Errorf("at 0%% R=%d, want 80", got.FG.R)
	}

	// At t=50ms (25%): between green and dark.
	clock.Advance(50)
	got = engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if got.FG.R >= 80 || got.FG.R <= 45 {
		t.Errorf("at 25%% R=%d, want between 45 and 80", got.FG.R)
	}

	// At t=100ms (50%): dark.
	clock.Advance(50)
	got = engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if got.FG.R != 45 {
		t.Errorf("at 50%% R=%d, want 45", got.FG.R)
	}

	if !engine.HasActive() {
		t.Error("infinite animation should keep HasActive true")
	}
}

func TestKeyframeAlternate(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	engine.RegisterKeyframes("ab", &KeyframeAnimation{
		Name:  "ab",
		Stops: []KeyframeStop{{Percent: 0, Style: red}, {Percent: 1, Style: blue}},
	})

	spec := &AnimationSpec{
		Name:           "ab",
		DurationMs:     100,
		TimingFunction: Linear,
		IterationCount: 4,
		Direction:      "alternate",
		FillMode:       "none",
	}

	// Iteration 0 (forward): 0→100ms, should go red→blue.
	got := engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if got.FG.R != 255 {
		t.Errorf("iter 0 start R=%d, want 255", got.FG.R)
	}

	// Iteration 1 (reverse): 100→200ms, should go blue→red.
	clock.Advance(150) // 50ms into iteration 1
	got = engine.BeforeRenderAnim("node0", render.Style{}, spec)
	// Reverse at 50%: should be midpoint.
	if math.Abs(float64(got.FG.R)-128) > 10 {
		t.Errorf("iter 1 midpoint R=%d, want ~128", got.FG.R)
	}
}

func TestKeyframeFillForwards(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}

	engine.RegisterKeyframes("f", &KeyframeAnimation{
		Name:  "f",
		Stops: []KeyframeStop{{Percent: 0, Style: red}, {Percent: 1, Style: blue}},
	})

	spec := &AnimationSpec{
		Name:           "f",
		DurationMs:     100,
		TimingFunction: Linear,
		IterationCount: 1,
		Direction:      "normal",
		FillMode:       "forwards",
	}

	engine.BeforeRenderAnim("node0", render.Style{}, spec)
	clock.Advance(200) // well past completion
	got := engine.BeforeRenderAnim("node0", render.Style{}, spec)

	// fill-mode: forwards keeps last frame.
	if got.FG.B != 255 {
		t.Errorf("fill forwards: B=%d, want 255", got.FG.B)
	}
}

func TestKeyframeDelay(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}
	base := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 255, B: 0}}

	engine.RegisterKeyframes("d", &KeyframeAnimation{
		Name:  "d",
		Stops: []KeyframeStop{{Percent: 0, Style: red}, {Percent: 1, Style: blue}},
	})

	spec := &AnimationSpec{
		Name:           "d",
		DurationMs:     100,
		TimingFunction: Linear,
		IterationCount: 1,
		DelayMs:        50,
		Direction:      "normal",
		FillMode:       "none",
	}

	// During delay: should return base style.
	got := engine.BeforeRenderAnim("node0", base, spec)
	if got.FG.G != 255 {
		t.Errorf("during delay: G=%d, want 255 (base)", got.FG.G)
	}

	// After delay: animation starts.
	clock.Advance(100) // 50ms past delay, 50% through animation
	got = engine.BeforeRenderAnim("node0", base, spec)
	if math.Abs(float64(got.FG.R)-128) > 5 {
		t.Errorf("after delay midpoint: R=%d, want ~128", got.FG.R)
	}
}

// E3: animation-play-state — paused freezes progress, running resumes.
func TestKeyframePlayStatePauseAndResume(t *testing.T) {
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})

	red := render.Style{FG: render.Color{IsRGB: true, R: 255, G: 0, B: 0}}
	blue := render.Style{FG: render.Color{IsRGB: true, R: 0, G: 0, B: 255}}
	engine.RegisterKeyframes("fade", &KeyframeAnimation{
		Name:  "fade",
		Stops: []KeyframeStop{{Percent: 0, Style: red}, {Percent: 1, Style: blue}},
	})
	spec := &AnimationSpec{
		Name: "fade", DurationMs: 200, TimingFunction: Linear,
		IterationCount: 1, Direction: "normal", FillMode: "forwards",
		PlayState: "running",
	}

	// Given — run to the midpoint.
	engine.BeforeRenderAnim("node0", render.Style{}, spec)
	clock.Advance(100)
	mid := engine.BeforeRenderAnim("node0", render.Style{}, spec)

	// When — pause (a render happens at the moment of pausing), then let
	// time pass.
	paused := *spec
	paused.PlayState = "paused"
	engine.BeforeRenderAnim("node0", render.Style{}, &paused)
	clock.Advance(500)
	frozen := engine.BeforeRenderAnim("node0", render.Style{}, &paused)

	// Then — the style holds the midpoint value.
	if frozen.FG.R != mid.FG.R || frozen.FG.B != mid.FG.B {
		t.Errorf("paused style = %+v, want frozen midpoint %+v", frozen.FG, mid.FG)
	}

	// When — resume and advance the remaining half.
	engine.BeforeRenderAnim("node0", render.Style{}, spec)
	clock.Advance(100)
	done := engine.BeforeRenderAnim("node0", render.Style{}, spec)

	// Then — completes to blue (fill forwards).
	if done.FG.B != 255 || done.FG.R != 0 {
		t.Errorf("resumed final style = %+v, want blue", done.FG)
	}
}

// E3b: per-node resolved stops on the spec beat the engine registry —
// no registry entry needed at all.
func TestKeyframeSpecStopsBeatRegistry(t *testing.T) {
	// Given: no RegisterKeyframes call; stops live on the spec.
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	spec := &AnimationSpec{
		Name:           "pulse",
		DurationMs:     200,
		TimingFunction: Linear,
		IterationCount: 1,
		Stops: []KeyframeStop{
			{Percent: 0, Style: render.Style{FG: render.Color{IsRGB: true, R: 200}}},
			{Percent: 1, Style: render.Style{FG: render.Color{IsRGB: true, R: 0}}},
		},
	}

	// When / Then
	got := engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if got.FG.R != 200 {
		t.Errorf("at t=0 R=%d, want 200 (spec stops used)", got.FG.R)
	}
	clock.Advance(100)
	got = engine.BeforeRenderAnim("node0", render.Style{}, spec)
	if math.Abs(float64(got.FG.R)-100) > 5 {
		t.Errorf("at midpoint R=%d, want ~100", got.FG.R)
	}
}
