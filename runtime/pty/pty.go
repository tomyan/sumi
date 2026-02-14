package pty

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Open creates a new PTY pair and returns the master and slave file descriptors.
// Uses /dev/ptmx on macOS with TIOCPTYGRANT/TIOCPTYUNLK/TIOCPTYGNAME ioctls.
func Open() (master, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}

	// Grant access to the slave PTY.
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), unix.TIOCPTYGRANT, 0); errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCPTYGRANT: %w", errno)
	}

	// Unlock the slave PTY.
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), unix.TIOCPTYUNLK, 0); errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCPTYUNLK: %w", errno)
	}

	// Get the slave device name.
	var nameBuf [128]byte
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), unix.TIOCPTYGNAME, uintptr(unsafe.Pointer(&nameBuf[0]))); errno != 0 {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCPTYGNAME: %w", errno)
	}

	slaveName := string(nameBuf[:cstrLen(nameBuf[:])])
	slave, err = os.OpenFile(slaveName, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave %s: %w", slaveName, err)
	}
	return master, slave, nil
}

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
