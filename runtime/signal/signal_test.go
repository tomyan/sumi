package signal_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/signal"
)

func TestNewSignalGetReturnsInitialValue(t *testing.T) {
	// Given
	s := signal.New(42)

	// When
	v := s.Get()

	// Then
	if v != 42 {
		t.Errorf("Get() = %d, want 42", v)
	}
}

func TestSignalSetUpdatesValue(t *testing.T) {
	// Given
	s := signal.New(0)

	// When
	s.Set(10)

	// Then
	if s.Get() != 10 {
		t.Errorf("Get() = %d, want 10", s.Get())
	}
}

func TestSignalUpdateReadModifyWrite(t *testing.T) {
	// Given
	s := signal.New(5)

	// When
	s.Update(func(v int) int { return v * 2 })

	// Then
	if s.Get() != 10 {
		t.Errorf("Get() = %d, want 10", s.Get())
	}
}
