package signal

// batchDepth tracks nested Batch calls. Notifications are deferred while > 0.
var batchDepth int

// pendingComputations collects computations to run when the outermost batch ends.
var pendingComputations []*computation

// Batch executes fn, deferring all reactive notifications until fn returns.
// Multiple signal.Set() calls within a batch trigger dependents only once.
func Batch(fn func()) {
	batchDepth++
	fn()
	batchDepth--
	if batchDepth == 0 {
		flushPending()
	}
}

// scheduleOrRun either runs a computation immediately or defers it for batch.
func scheduleOrRun(c *computation) {
	if batchDepth > 0 {
		for _, p := range pendingComputations {
			if p == c {
				return // already scheduled
			}
		}
		pendingComputations = append(pendingComputations, c)
		return
	}
	c.run()
}

func flushPending() {
	for len(pendingComputations) > 0 {
		// Take the current batch and clear it — computations may schedule more.
		batch := pendingComputations
		pendingComputations = nil
		for _, c := range batch {
			c.run()
		}
	}
}
