package input

import "io"

// PasteEnableSeq enables bracketed paste mode.
const PasteEnableSeq = "\x1b[?2004h"

// PasteDisableSeq disables bracketed paste mode.
const PasteDisableSeq = "\x1b[?2004l"

// readBracketedPaste reads bytes until the end marker ESC[201~ and returns the content.
func readBracketedPaste(r io.Reader) (string, error) {
	var buf []byte
	// End marker is: ESC [ 2 0 1 ~
	end := []byte{0x1b, '[', '2', '0', '1', '~'}
	for {
		b, err := readByte(r)
		if err != nil {
			return string(buf), err
		}
		buf = append(buf, b)
		if len(buf) >= len(end) && matchSuffix(buf, end) {
			return string(buf[:len(buf)-len(end)]), nil
		}
	}
}

// matchSuffix checks if buf ends with suffix.
func matchSuffix(buf, suffix []byte) bool {
	start := len(buf) - len(suffix)
	for i, b := range suffix {
		if buf[start+i] != b {
			return false
		}
	}
	return true
}
