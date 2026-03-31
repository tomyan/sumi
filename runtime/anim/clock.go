package anim

import "time"

// Clock provides the current time in milliseconds for animations.
type Clock interface {
	Now() int64
}

// WallClock uses the system clock.
type WallClock struct{}

// Now returns the current time in milliseconds.
func (WallClock) Now() int64 {
	return time.Now().UnixMilli()
}

// TestClock is a deterministic clock for testing animations.
type TestClock struct {
	now int64
}

// NewTestClock creates a TestClock starting at 0.
func NewTestClock() *TestClock {
	return &TestClock{}
}

// Now returns the current test time.
func (c *TestClock) Now() int64 {
	return c.now
}

// Advance moves the clock forward by ms milliseconds.
func (c *TestClock) Advance(ms int64) {
	c.now += ms
}
