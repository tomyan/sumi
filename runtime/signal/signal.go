package signal

// Signal is a reactive value that notifies dependents when it changes.
type Signal[T any] struct {
	value T
	subs  []*computation
}

// New creates a new signal with the given initial value.
func New[T any](initial T) *Signal[T] {
	return &Signal[T]{value: initial}
}

// Get returns the current value. If called during a tracked computation
// (From or Effect), the computation is registered as a dependent.
func (s *Signal[T]) Get() T {
	if cur := currentComputation(); cur != nil {
		s.subscribe(cur)
	}
	return s.value
}

// Set updates the value and notifies all dependents.
func (s *Signal[T]) Set(v T) {
	s.value = v
	s.notify()
}

// Update applies a function to the current value and sets the result.
func (s *Signal[T]) Update(fn func(T) T) {
	s.Set(fn(s.value))
}

func (s *Signal[T]) subscribe(c *computation) {
	for _, sub := range s.subs {
		if sub == c {
			return
		}
	}
	s.subs = append(s.subs, c)
	c.sources = append(c.sources, &signalRef{unsubscribe: func() { s.removeSub(c) }})
}

func (s *Signal[T]) removeSub(c *computation) {
	for i, sub := range s.subs {
		if sub == c {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

func (s *Signal[T]) notify() {
	// Copy to avoid issues if subs changes during notification.
	subs := make([]*computation, len(s.subs))
	copy(subs, s.subs)
	for _, c := range subs {
		scheduleOrRun(c)
	}
}
