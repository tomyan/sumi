package anim

import "testing"

func TestTestClockStartsAtZero(t *testing.T) {
	c := NewTestClock()
	if got := c.Now(); got != 0 {
		t.Errorf("Now() = %d, want 0", got)
	}
}

func TestTestClockAdvance(t *testing.T) {
	c := NewTestClock()
	c.Advance(100)
	if got := c.Now(); got != 100 {
		t.Errorf("Now() = %d, want 100", got)
	}
	c.Advance(50)
	if got := c.Now(); got != 150 {
		t.Errorf("Now() = %d, want 150", got)
	}
}

func TestWallClockReturnsPositive(t *testing.T) {
	c := WallClock{}
	if got := c.Now(); got <= 0 {
		t.Errorf("WallClock.Now() = %d, expected positive", got)
	}
}
