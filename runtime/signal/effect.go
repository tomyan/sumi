package signal

// Effect creates a reactive side effect. The function fn runs immediately,
// and re-runs whenever any signal read during its execution changes.
// Returns a dispose function that stops the effect from re-running.
func Effect(fn func()) func() {
	c := &computation{fn: fn}
	c.run()
	return func() {
		c.cleanup()
		// Mark as disposed so it won't be re-added.
		c.fn = nil
	}
}
