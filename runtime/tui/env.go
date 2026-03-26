package tui

import "github.com/tomyan/sumi/runtime/signal"

// envSignals stores framework-provided environment signals by name.
var envSignals = make(map[string]any)

// Env returns a reactive signal for the named environment value.
// Framework updates these (e.g., terminal width/height on SIGWINCH).
// Returns the same signal instance for repeated calls with the same name.
func Env[T any](name string) *signal.Signal[T] {
	if s, ok := envSignals[name]; ok {
		return s.(*signal.Signal[T])
	}
	s := signal.New(*new(T))
	envSignals[name] = s
	return s
}

// updateEnvSignals sets the width and height env signals to current terminal dimensions.
func updateEnvSignals(w, h int) {
	SetEnv("width", w)
	SetEnv("height", h)
}

// SetEnv updates a named environment signal.
func SetEnv[T any](name string, value T) {
	if s, ok := envSignals[name]; ok {
		s.(*signal.Signal[T]).Set(value)
		return
	}
	s := signal.New(value)
	envSignals[name] = s
}
