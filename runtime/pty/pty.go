//go:build darwin || linux

package pty

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// SetSize sets the terminal window size on the PTY master.
func SetSize(master *os.File, rows, cols int) error {
	ws := unix.Winsize{Row: uint16(rows), Col: uint16(cols)}
	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL, master.Fd(),
		uintptr(unix.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	); errno != 0 {
		return fmt.Errorf("TIOCSWINSZ: %w", errno)
	}
	return nil
}

// Start launches cmd with stdin/stdout/stderr connected to a PTY slave.
// Returns the master end. The caller must close master and wait on cmd.
func Start(cmd *exec.Cmd, rows, cols int) (*os.File, error) {
	master, slave, err := Open()
	if err != nil {
		return nil, err
	}
	defer slave.Close()

	if err := SetSize(master, rows, cols); err != nil {
		master.Close()
		return nil, err
	}

	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}

	if err := cmd.Start(); err != nil {
		master.Close()
		return nil, fmt.Errorf("start: %w", err)
	}
	return master, nil
}

// cstrLen returns the length of a null-terminated byte slice.
func cstrLen(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}
