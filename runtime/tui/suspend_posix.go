//go:build !js

package tui

import "syscall"

// stopSelf raises SIGTSTP; execution resumes here after SIGCONT.
func stopSelf() { _ = syscall.Kill(syscall.Getpid(), syscall.SIGTSTP) }
