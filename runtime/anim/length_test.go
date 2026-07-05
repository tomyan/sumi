package anim

import "testing"

func widthSpec(durationMs int) []TransitionSpec {
	return []TransitionSpec{{Property: "width", DurationMs: durationMs, TimingFunction: Linear}}
}

func TestStepLengthFirstRenderAdoptsTarget(t *testing.T) {
	// Given
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	state := &LengthState{}

	// When / Then — no transition on the first sighting
	got, active := engine.StepLength(state, "width", 30, widthSpec(200))
	if got != 30 || active {
		t.Errorf("first render = (%d, %v), want (30, false)", got, active)
	}
}

func TestStepLengthInterpolatesWholeCells(t *testing.T) {
	// Given — width settles at 10, then the target jumps to 30
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	state := &LengthState{}
	engine.StepLength(state, "width", 10, widthSpec(200))

	// When — halfway through a 200ms transition
	got, active := engine.StepLength(state, "width", 30, widthSpec(200))
	if got != 10 || !active {
		t.Fatalf("at t=0: (%d, %v), want (10, true)", got, active)
	}
	clock.Advance(100)
	got, active = engine.StepLength(state, "width", 30, widthSpec(200))

	// Then — whole-cell midpoint
	if got != 20 || !active {
		t.Errorf("at midpoint: (%d, %v), want (20, true)", got, active)
	}

	// When — complete
	clock.Advance(100)
	got, active = engine.StepLength(state, "width", 30, widthSpec(200))
	if got != 30 || active {
		t.Errorf("at end: (%d, %v), want (30, false)", got, active)
	}
}

func TestStepLengthRetargetsMidFlight(t *testing.T) {
	// Given — transition from 10 toward 30, interrupted at the midpoint
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	state := &LengthState{}
	engine.StepLength(state, "width", 10, widthSpec(200))
	engine.StepLength(state, "width", 30, widthSpec(200))
	clock.Advance(100)
	engine.StepLength(state, "width", 30, widthSpec(200)) // displays 20

	// When — target changes back to 10; the new transition starts at 20
	got, _ := engine.StepLength(state, "width", 10, widthSpec(200))
	if got != 20 {
		t.Fatalf("retarget start = %d, want 20 (current displayed value)", got)
	}
	clock.Advance(100)
	got, _ = engine.StepLength(state, "width", 10, widthSpec(200))

	// Then — halfway from 20 down to 10
	if got != 15 {
		t.Errorf("retarget midpoint = %d, want 15", got)
	}
}

func TestStepLengthNoSpecSnapsImmediately(t *testing.T) {
	// Given — specs cover only color
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	state := &LengthState{}
	specs := []TransitionSpec{{Property: "color", DurationMs: 200, TimingFunction: Linear}}
	engine.StepLength(state, "width", 10, specs)

	// When
	got, active := engine.StepLength(state, "width", 30, specs)

	// Then
	if got != 30 || active {
		t.Errorf("uncovered property = (%d, %v), want snap (30, false)", got, active)
	}
}

func TestStepLengthHonorsDelay(t *testing.T) {
	// Given
	clock := NewTestClock()
	engine := NewEngine(clock, func() {})
	state := &LengthState{}
	specs := []TransitionSpec{{Property: "width", DurationMs: 100, DelayMs: 100, TimingFunction: Linear}}
	engine.StepLength(state, "width", 10, specs)
	engine.StepLength(state, "width", 30, specs)

	// When — still inside the delay
	clock.Advance(50)
	got, active := engine.StepLength(state, "width", 30, specs)

	// Then — holds the from value
	if got != 10 || !active {
		t.Errorf("during delay = (%d, %v), want (10, true)", got, active)
	}
}
