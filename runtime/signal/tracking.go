package signal

// signalRef holds an unsubscribe function for cleaning up dependency tracking.
type signalRef struct {
	unsubscribe func()
}

// computation represents a tracked reactive computation (From or Effect).
type computation struct {
	fn      func()
	sources []*signalRef
}

// run clears old subscriptions and re-executes the computation,
// re-establishing dependency tracking from scratch.
func (c *computation) run() {
	if c.fn == nil {
		return // disposed
	}
	c.cleanup()
	trackingStack = append(trackingStack, c)
	c.fn()
	trackingStack = trackingStack[:len(trackingStack)-1]
}

// cleanup removes all subscriptions from source signals.
func (c *computation) cleanup() {
	for _, ref := range c.sources {
		ref.unsubscribe()
	}
	c.sources = c.sources[:0]
}

// trackingStack is the global stack of active computations.
// When a Signal.Get() is called, the top of this stack (if any)
// is registered as a dependent.
var trackingStack []*computation

// currentComputation returns the computation currently being tracked, or nil.
func currentComputation() *computation {
	if len(trackingStack) == 0 {
		return nil
	}
	return trackingStack[len(trackingStack)-1]
}
