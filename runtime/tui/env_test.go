package tui_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

func TestEnvSignalReturnsSignal(t *testing.T) {
	// Given/When
	width := tui.Env[int]("width")

	// Then — should return a signal (not nil)
	if width == nil {
		t.Fatal("Env('width') returned nil")
	}
}

func TestEnvSignalSameInstance(t *testing.T) {
	// Given
	w1 := tui.Env[int]("width")
	w2 := tui.Env[int]("width")

	// Then — same signal instance
	if w1 != w2 {
		t.Error("Env('width') should return the same instance")
	}
}

func TestEnvSignalUpdatable(t *testing.T) {
	// Given
	width := tui.Env[int]("width")

	// When — framework updates it
	tui.SetEnv("width", 120)

	// Then
	if width.Get() != 120 {
		t.Errorf("width.Get() = %d, want 120", width.Get())
	}
}

func TestEnvSignalReactive(t *testing.T) {
	// Given
	width := tui.Env[int]("width")
	tui.SetEnv("width", 80)
	var observed int
	signal.Effect(func() {
		observed = width.Get()
	})

	// When
	tui.SetEnv("width", 120)

	// Then
	if observed != 120 {
		t.Errorf("observed = %d, want 120", observed)
	}
}
