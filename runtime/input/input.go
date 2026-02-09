package input

import (
	"io"
	"unicode/utf8"

	"golang.org/x/term"
)

// EnableRawMode puts the terminal in raw mode (no line buffering, no echo).
// Returns a restore function that MUST be called to restore the terminal.
// The restore function is safe to call multiple times.
func EnableRawMode(fd int) (restore func(), err error) {
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return func() {
		term.Restore(fd, oldState)
	}, nil
}

// ReadKey reads a single keypress from the reader.
// Returns the rune and any error.
func ReadKey(r io.Reader) (rune, error) {
	var buf [4]byte

	// Read the first byte.
	n, err := r.Read(buf[:1])
	if n == 0 {
		if err != nil {
			return 0, err
		}
		return 0, io.EOF
	}

	b := buf[0]

	// Single-byte ASCII.
	if b < 0x80 {
		return rune(b), nil
	}

	// Determine how many bytes this UTF-8 character needs.
	var size int
	switch {
	case b&0xE0 == 0xC0:
		size = 2
	case b&0xF0 == 0xE0:
		size = 3
	case b&0xF8 == 0xF0:
		size = 4
	default:
		// Invalid leading byte — return replacement character.
		return utf8.RuneError, nil
	}

	// Read remaining bytes.
	remaining := size - 1
	total, err := io.ReadFull(r, buf[1:size])
	if total < remaining {
		return utf8.RuneError, err
	}

	ch, _ := utf8.DecodeRune(buf[:size])
	return ch, nil
}
