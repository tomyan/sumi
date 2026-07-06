//go:build !js

package term

import (
	"os"
	"os/signal"
	"syscall"
)

// WatchResize returns a channel that receives a value whenever the terminal
// is resized (SIGWINCH). The caller should read from the channel in a select.
// Call the returned stop function to clean up when done.
func WatchResize() (resize <-chan struct{}, stop func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	ch := make(chan struct{}, 1)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-sigCh:
				// Non-blocking send — drop if the channel is full
				select {
				case ch <- struct{}{}:
				default:
				}
			case <-done:
				signal.Stop(sigCh)
				close(ch)
				return
			}
		}
	}()

	return ch, func() { close(done) }
}
