package signal_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/signal"
)

func TestBatchDefersNotifications(t *testing.T) {
	// Given
	a := signal.New(1)
	b := signal.New(2)
	runCount := 0
	signal.Effect(func() {
		_ = a.Get() + b.Get()
		runCount++
	})
	runCount = 0 // reset after initial run

	// When — batch multiple sets
	signal.Batch(func() {
		a.Set(10)
		b.Set(20)
	})

	// Then — effect should have run only once, not twice
	if runCount != 1 {
		t.Errorf("runCount = %d, want 1 (batched)", runCount)
	}
}

func TestBatchStillUpdatesValues(t *testing.T) {
	// Given
	s := signal.New(0)

	// When
	signal.Batch(func() {
		s.Set(42)
	})

	// Then
	if s.Get() != 42 {
		t.Errorf("Get() = %d, want 42", s.Get())
	}
}

func TestBatchNestedBatches(t *testing.T) {
	// Given
	s := signal.New(0)
	runCount := 0
	signal.Effect(func() {
		_ = s.Get()
		runCount++
	})
	runCount = 0

	// When — nested batches
	signal.Batch(func() {
		s.Set(1)
		signal.Batch(func() {
			s.Set(2)
		})
		s.Set(3)
	})

	// Then — effect should run once after outermost batch
	if runCount != 1 {
		t.Errorf("runCount = %d, want 1", runCount)
	}
	if s.Get() != 3 {
		t.Errorf("Get() = %d, want 3", s.Get())
	}
}
