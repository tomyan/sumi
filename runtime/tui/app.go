package tui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// App owns the terminal event loop for a Sumi application.
// Generated code builds the tree, defines handlers, and calls app.Run().
type App struct {
	OnRender  func()            // called to perform a render
	OnEvent   func(input.Event) // called for each input event
	OnResize  func()            // called on terminal resize
	HasMouse  bool              // enable SGR mouse mode
	Title     string            // static terminal title (saved/set/restored around Run)
	SaveTitle bool              // save/restore terminal title only (for dynamic titles set in doRender)
	Dirty     bool              // set by handlers to trigger re-render
}

// Run sets up the terminal, starts the event reader, and runs the event loop.
func (a *App) Run() {
	if a.Title != "" || a.SaveTitle {
		fmt.Fprint(os.Stdout, "\033[22;2t") // save current title
	}

	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(os.Stdout)
	defer render.ExitAlternateScreen(os.Stdout)

	if a.HasMouse {
		fmt.Fprint(os.Stdout, input.MouseEnableSeq)
		defer fmt.Fprint(os.Stdout, input.MouseDisableSeq)
	}

	if a.Title != "" || a.SaveTitle {
		defer fmt.Fprint(os.Stdout, "\033[23;2t") // restore title
	}
	if a.Title != "" {
		fmt.Fprintf(os.Stdout, "\033]2;%s\007", a.Title)
	}

	eventCh := make(chan input.Event, 64)
	go func() {
		for {
			evt, err := input.ReadEvent(os.Stdin)
			if err != nil {
				close(eventCh)
				return
			}
			eventCh <- evt
		}
	}()

	resizeCh, stopResize := term.WatchResize()
	defer stopResize()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	// Convert os.Signal channel to struct{} channel for runLoop
	shutdownCh := make(chan struct{}, 1)
	go func() {
		<-sigCh
		shutdownCh <- struct{}{}
	}()

	a.runLoop(eventCh, resizeCh, shutdownCh)
}

// runLoop is the internal event loop, injectable for testing.
// eventCh delivers input events, resizeCh delivers resize signals,
// sigCh delivers shutdown signals. Nil channels are ignored.
func (a *App) runLoop(eventCh <-chan input.Event, resizeCh <-chan struct{}, sigCh <-chan struct{}) {
	// Initial render
	a.Dirty = true
	a.OnRender()
	a.Dirty = false

	for {
		// Wait for at least one event
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return
			}
			if isQuit(evt) {
				return
			}
			a.dispatchEvent(evt)

		case <-resizeCh:
			a.dispatchResize()

		case <-sigCh:
			return
		}

		// Drain pending events before rendering
		a.drain(eventCh, resizeCh, sigCh)

		if a.Dirty {
			a.OnRender()
			a.Dirty = false
		}
	}
}

// isQuit returns true for Ctrl+C (rune 3).
func isQuit(evt input.Event) bool {
	return evt.Kind == input.EventKey && evt.Rune == 3
}

// dispatchEvent calls OnEvent if set.
func (a *App) dispatchEvent(evt input.Event) {
	if a.OnEvent != nil {
		a.OnEvent(evt)
	}
}

// dispatchResize calls OnResize if set and marks dirty.
func (a *App) dispatchResize() {
	if a.OnResize != nil {
		a.OnResize()
	}
	a.Dirty = true
}

// drain non-blocking reads all pending events from all channels.
// Returns true if a quit/signal was received.
func (a *App) drain(eventCh <-chan input.Event, resizeCh <-chan struct{}, sigCh <-chan struct{}) bool {
	for {
		select {
		case evt, ok := <-eventCh:
			if !ok {
				return true
			}
			if isQuit(evt) {
				return true
			}
			a.dispatchEvent(evt)

		case <-resizeCh:
			a.dispatchResize()

		case <-sigCh:
			return true

		default:
			return false
		}
	}
}
