// Minimal app for the Ctrl+Z suspend PTY test. The PTY child is a
// session leader, so its orphaned process group would have a default
// SIGTSTP discarded (POSIX); the test asks for SIGSTOP instead, which
// cannot be discarded — the rest of the suspend path is identical.
package main

import (
	"os"
	"syscall"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/tui"
)

func main() {
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Kind: layout.KindText, Content: "suspend me", CursorCol: -1, CursorRow: -1},
		},
	}}
	tui.RunWithOptions(comp, tui.RunOptions{
		ExitOn: []string{"q"},
		SetApp: func(a *tui.App) {
			if os.Getenv("SUSPEND_TEST_SIGSTOP") != "" {
				a.SuspendHooks = &tui.SuspendHooks{
					Stop: func() { _ = syscall.Kill(syscall.Getpid(), syscall.SIGSTOP) },
				}
			}
		},
	})
}
