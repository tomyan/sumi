package pty

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Open creates a new PTY pair and returns the master and slave file descriptors.
// Uses /dev/ptmx with the macOS TIOCPTYGRANT/TIOCPTYUNLK/TIOCPTYGNAME ioctls.
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
