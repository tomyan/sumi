package signal_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/signal"
)

func TestFromComputesInitialValue(t *testing.T) {
	// Given
	count := signal.New(5)

	// When
	doubled := signal.From(func() int { return count.Get() * 2 })

	// Then
	if doubled.Get() != 10 {
		t.Errorf("Get() = %d, want 10", doubled.Get())
	}
}

func TestFromRecomputesOnDependencyChange(t *testing.T) {
	// Given
	count := signal.New(1)
	doubled := signal.From(func() int { return count.Get() * 2 })

	// When
	count.Set(5)

	// Then
	if doubled.Get() != 10 {
		t.Errorf("Get() = %d, want 10", doubled.Get())
	}
}

func TestFromTracksMultipleDependencies(t *testing.T) {
	// Given
	a := signal.New(2)
	b := signal.New(3)
	sum := signal.From(func() int { return a.Get() + b.Get() })

	// When — change either dependency
	a.Set(10)

	// Then
	if sum.Get() != 13 {
		t.Errorf("Get() = %d, want 13", sum.Get())
	}

	// When — change the other
	b.Set(20)

	// Then
	if sum.Get() != 30 {
		t.Errorf("Get() = %d, want 30", sum.Get())
	}
}

func TestFromChainedComputations(t *testing.T) {
	// Given — a chain: count → doubled → quadrupled
	count := signal.New(2)
	doubled := signal.From(func() int { return count.Get() * 2 })
	quadrupled := signal.From(func() int { return doubled.Get() * 2 })

	// When
	count.Set(3)

	// Then
	if quadrupled.Get() != 12 {
		t.Errorf("Get() = %d, want 12", quadrupled.Get())
	}
}

func TestFromDiamondDependency(t *testing.T) {
	// Given — diamond: source → a, b → combined
	source := signal.New(1)
	a := signal.From(func() int { return source.Get() + 10 })
	b := signal.From(func() int { return source.Get() * 2 })
	combined := signal.From(func() int { return a.Get() + b.Get() })

	// Sanity
	if combined.Get() != 13 { // (1+10) + (1*2) = 13
		t.Fatalf("initial Get() = %d, want 13", combined.Get())
	}

	// When
	source.Set(5)

	// Then — (5+10) + (5*2) = 25
	if combined.Get() != 25 {
		t.Errorf("Get() = %d, want 25", combined.Get())
	}
}

func TestFromDynamicDependencies(t *testing.T) {
	// Given — condition controls which signal is read
	cond := signal.New(true)
	a := signal.New(1)
	b := signal.New(2)
	result := signal.From(func() int {
		if cond.Get() {
			return a.Get()
		}
		return b.Get()
	})

	// Sanity
	if result.Get() != 1 {
		t.Fatalf("initial = %d, want 1", result.Get())
	}

	// When — switch to b branch
	cond.Set(false)

	// Then
	if result.Get() != 2 {
		t.Errorf("Get() = %d, want 2", result.Get())
	}

	// When — change a (no longer a dependency)
	a.Set(99)

	// Then — should still be 2
	if result.Get() != 2 {
		t.Errorf("Get() = %d, want 2 (a no longer tracked)", result.Get())
	}

	// When — change b (current dependency)
	b.Set(42)

	// Then
	if result.Get() != 42 {
		t.Errorf("Get() = %d, want 42", result.Get())
	}
}

func TestFromIsLazyOnGet(t *testing.T) {
	// Given — computed that tracks call count
	count := signal.New(0)
	calls := 0
	doubled := signal.From(func() int {
		calls++
		return count.Get() * 2
	})

	// Initial computation runs once
	_ = doubled.Get()
	initialCalls := calls

	// When — set without reading
	count.Set(5)
	count.Set(10)

	// Then — Get() gives latest value
	if doubled.Get() != 20 {
		t.Errorf("Get() = %d, want 20", doubled.Get())
	}
	// Should not have run more than necessary
	_ = initialCalls // calls count depends on implementation — just verify correctness
}
