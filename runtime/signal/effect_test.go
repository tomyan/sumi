package signal_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/signal"
)

func TestEffectRunsImmediately(t *testing.T) {
	// Given
	count := signal.New(5)
	var observed int

	// When
	signal.Effect(func() {
		observed = count.Get()
	})

	// Then
	if observed != 5 {
		t.Errorf("observed = %d, want 5", observed)
	}
}

func TestEffectRerunsOnDependencyChange(t *testing.T) {
	// Given
	count := signal.New(0)
	var observed int
	signal.Effect(func() {
		observed = count.Get()
	})

	// When
	count.Set(42)

	// Then
	if observed != 42 {
		t.Errorf("observed = %d, want 42", observed)
	}
}

func TestEffectTracksMultipleDependencies(t *testing.T) {
	// Given
	a := signal.New(1)
	b := signal.New(2)
	var observed int
	signal.Effect(func() {
		observed = a.Get() + b.Get()
	})

	// When
	a.Set(10)

	// Then
	if observed != 12 {
		t.Errorf("observed = %d, want 12", observed)
	}

	// When
	b.Set(20)

	// Then
	if observed != 30 {
		t.Errorf("observed = %d, want 30", observed)
	}
}

func TestEffectRetracksDynamicDependencies(t *testing.T) {
	// Given
	cond := signal.New(true)
	a := signal.New(1)
	b := signal.New(2)
	var observed int
	signal.Effect(func() {
		if cond.Get() {
			observed = a.Get()
		} else {
			observed = b.Get()
		}
	})

	// Sanity
	if observed != 1 {
		t.Fatalf("initial = %d, want 1", observed)
	}

	// When — switch branch
	cond.Set(false)
	if observed != 2 {
		t.Errorf("after switch = %d, want 2", observed)
	}

	// When — change a (no longer tracked)
	a.Set(99)
	if observed != 2 {
		t.Errorf("after a change = %d, want 2 (not tracked)", observed)
	}

	// When — change b (now tracked)
	b.Set(50)
	if observed != 50 {
		t.Errorf("after b change = %d, want 50", observed)
	}
}

func TestEffectDispose(t *testing.T) {
	// Given
	count := signal.New(0)
	runCount := 0
	dispose := signal.Effect(func() {
		_ = count.Get()
		runCount++
	})

	// Sanity — ran once
	if runCount != 1 {
		t.Fatalf("runCount = %d, want 1", runCount)
	}

	// When — dispose then change
	dispose()
	count.Set(10)

	// Then — should not have run again
	if runCount != 1 {
		t.Errorf("runCount = %d, want 1 after dispose", runCount)
	}
}

func TestEffectOnComputed(t *testing.T) {
	// Given — effect depends on a computed
	count := signal.New(2)
	doubled := signal.From(func() int { return count.Get() * 2 })
	var observed int
	signal.Effect(func() {
		observed = doubled.Get()
	})

	// Sanity
	if observed != 4 {
		t.Fatalf("initial = %d, want 4", observed)
	}

	// When
	count.Set(5)

	// Then
	if observed != 10 {
		t.Errorf("observed = %d, want 10", observed)
	}
}
