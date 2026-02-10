package term

import (
	"golang.org/x/term"
)

// GetSize returns the current terminal width and height for the given fd.
// Falls back to 80x24 if the size cannot be determined.
func GetSize(fd int) (width, height int) {
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	return w, h
}
