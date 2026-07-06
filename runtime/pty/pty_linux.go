package pty

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// Open creates a new PTY pair and returns the master and slave file
// descriptors. Uses /dev/ptmx with the Linux TIOCSPTLCK unlock and
// TIOCGPTN slave-number ioctls.
func Open() (master, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}

	// Unlock the slave PTY.
	unlock := 0
	if err := unix.IoctlSetPointerInt(int(master.Fd()), unix.TIOCSPTLCK, unlock); err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCSPTLCK: %w", err)
	}

	// Get the slave device number.
	n, err := unix.IoctlGetInt(int(master.Fd()), unix.TIOCGPTN)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("TIOCGPTN: %w", err)
	}

	slaveName := fmt.Sprintf("/dev/pts/%d", n)
	slave, err = os.OpenFile(slaveName, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave %s: %w", slaveName, err)
	}
	return master, slave, nil
}
