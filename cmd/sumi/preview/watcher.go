package preview

import (
	"os"
	"time"
)

// Watcher polls files for mtime changes and calls a callback when a change is detected.
type Watcher struct {
	stop chan struct{}
	done chan struct{}
}

// NewWatcher starts a background goroutine that polls paths at the given interval.
// When a file's mtime changes, onChanged is called with the path.
func NewWatcher(paths []string, interval time.Duration, onChanged func(string)) *Watcher {
	w := &Watcher{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	// Record initial mtimes.
	mtimes := make(map[string]time.Time)
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil {
			mtimes[p] = info.ModTime()
		}
	}

	go func() {
		defer close(w.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.stop:
				return
			case <-ticker.C:
				for _, p := range paths {
					info, err := os.Stat(p)
					if err != nil {
						continue
					}
					mt := info.ModTime()
					if prev, ok := mtimes[p]; ok && mt.After(prev) {
						mtimes[p] = mt
						onChanged(p)
					} else if !ok {
						mtimes[p] = mt
					}
				}
			}
		}
	}()

	return w
}

// Stop terminates the watcher goroutine.
func (w *Watcher) Stop() {
	select {
	case <-w.stop:
		return // already stopped
	default:
		close(w.stop)
	}
	<-w.done
}
