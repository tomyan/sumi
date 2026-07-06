package tui

import (
	"fmt"
	"syscall"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
)

// SuspendHooks abstracts terminal teardown, process stop, and terminal
// re-entry so tests can observe sequencing. Nil uses real POSIX
// behaviour (SIGTSTP; execution resumes after SIGCONT).
type SuspendHooks struct {
	ExitTerminal  func()
	Stop          func()
	EnterTerminal func()
}

// Suspend hands control back to the shell — the Ctrl+Z default action:
// restore the terminal, stop the process, and on resume re-enter the
// terminal and force a clean full repaint (the shell drew over us).
// Unset hook fields fall back to the real behaviour individually.
func (a *App) Suspend() {
	h := a.SuspendHooks
	if h == nil {
		h = &SuspendHooks{}
	}
	runHook(h.ExitTerminal, a.exitTerminal)
	runHook(h.Stop, func() { _ = syscall.Kill(syscall.Getpid(), syscall.SIGTSTP) })
	runHook(h.EnterTerminal, a.enterTerminal)
	a.NeedsFullRedraw = true
	a.Dirty = true
}

func runHook(hook, fallback func()) {
	if hook != nil {
		hook()
		return
	}
	fallback()
}

// isSuspendChord reports whether the event is the Ctrl+Z suspend chord.
func isSuspendChord(evt input.Event) bool {
	return evt.Kind == input.EventKey && evt.Ctrl && evt.Rune == 'z'
}

// enterTerminal puts the terminal into application state: raw mode,
// alternate screen, bracketed paste, and mouse tracking when enabled.
func (a *App) enterTerminal() {
	if fd, ok := a.inFd(); ok {
		a.termRestore, _ = input.EnableRawMode(fd)
	}
	if a.Inline {
		render.HideCursor(a.out())
	} else {
		render.EnterAlternateScreen(a.out())
	}
	fmt.Fprint(a.out(), input.PasteEnableSeq)
	fmt.Fprint(a.out(), input.KittyEnableSeq)
	if a.HasMouse {
		fmt.Fprint(a.out(), input.MouseEnableSeq)
	}
}

// exitTerminal restores the shell's terminal state (reverse order of
// enterTerminal).
func (a *App) exitTerminal() {
	if a.HasMouse {
		fmt.Fprint(a.out(), input.MouseDisableSeq)
	}
	fmt.Fprint(a.out(), input.KittyDisableSeq)
	fmt.Fprint(a.out(), input.PasteDisableSeq)
	if a.Inline {
		// Park the cursor after the content and leave the final frame
		// in scrollback; forget the zone (suspend resumes fresh).
		a.out().Write(a.inlineZone().Finish())
		a.inlineZone().Reset()
	} else {
		render.ExitAlternateScreen(a.out())
	}
	if a.termRestore != nil {
		a.termRestore()
		a.termRestore = nil
	}
}
