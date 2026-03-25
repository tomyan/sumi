package signal

// Computed is a derived reactive value that auto-tracks its dependencies.
// It recomputes when any signal read during its computation changes.
type Computed[T any] struct {
	fn    func() T
	value T
	dirty bool
	comp  *computation
	subs  []*computation
}

// From creates a computed signal. The function fn is called immediately to
// compute the initial value, and re-called whenever its dependencies change.
func From[T any](fn func() T) *Computed[T] {
	c := &Computed[T]{fn: fn, dirty: true}
	c.comp = &computation{fn: c.recompute}
	c.comp.run() // initial computation — establishes dependencies
	return c
}

// Get returns the current value. If called during a tracked computation,
// that computation is registered as a dependent.
func (c *Computed[T]) Get() T {
	if c.dirty {
		c.comp.run()
	}
	if cur := currentComputation(); cur != nil {
		c.subscribe(cur)
	}
	return c.value
}

func (c *Computed[T]) recompute() {
	c.value = c.fn()
	c.dirty = false
	c.notifySubs()
}

func (c *Computed[T]) subscribe(comp *computation) {
	for _, sub := range c.subs {
		if sub == comp {
			return
		}
	}
	c.subs = append(c.subs, comp)
	comp.sources = append(comp.sources, &signalRef{unsubscribe: func() { c.removeSub(comp) }})
}

func (c *Computed[T]) removeSub(comp *computation) {
	for i, sub := range c.subs {
		if sub == comp {
			c.subs = append(c.subs[:i], c.subs[i+1:]...)
			return
		}
	}
}

func (c *Computed[T]) notifySubs() {
	subs := make([]*computation, len(c.subs))
	copy(subs, c.subs)
	for _, sub := range subs {
		scheduleOrRun(sub)
	}
}
