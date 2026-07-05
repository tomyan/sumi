package input

import (
	"strconv"
	"strings"
)

// MouseAction identifies what the mouse did.
type MouseAction int

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMotion
	MouseScroll
)

// MouseButton identifies which button or scroll direction.
type MouseButton int

const (
	ButtonLeft   MouseButton = 0
	ButtonMiddle MouseButton = 1
	ButtonRight  MouseButton = 2
	ScrollUp     MouseButton = 64
	ScrollDown   MouseButton = 65
)

// MouseEvent holds parsed mouse event data.
type MouseEvent struct {
	Action MouseAction
	Button MouseButton
	Shift  bool // shift key held during mouse event
	X, Y   int  // 0-indexed buffer coordinates
}

// MouseEnableSeq enables SGR extended mouse mode + any-event tracking.
const MouseEnableSeq = "\x1b[?1006h\x1b[?1003h"

// MouseDisableSeq disables mouse tracking.
const MouseDisableSeq = "\x1b[?1003l\x1b[?1006l"

// parseSGRMouse parses a SGR mouse sequence after ESC[< has been consumed.
// Format: button;x;yM (press) or button;x;ym (release)
func parseSGRMouse(r readByteFunc, params string) (Event, error) {
	// Read remaining bytes of the sequence: digits, semicolons, then M or m
	var buf strings.Builder
	buf.WriteString(params)
	for {
		b, err := r()
		if err != nil {
			return Event{Kind: EventKey, Rune: 0x1b}, nil
		}
		if b == 'M' || b == 'm' {
			return decodeSGRMouse(buf.String(), b)
		}
		buf.WriteByte(b)
	}
}

type readByteFunc func() (byte, error)

func decodeSGRMouse(params string, terminator byte) (Event, error) {
	parts := strings.Split(params, ";")
	if len(parts) != 3 {
		return Event{Kind: EventKey, Rune: 0x1b}, nil
	}

	code, _ := strconv.Atoi(parts[0])
	x, _ := strconv.Atoi(parts[1])
	y, _ := strconv.Atoi(parts[2])

	// Terminal coords are 1-indexed; convert to 0-indexed
	x--
	y--

	me := MouseEvent{
		X: x,
		Y: y,
	}

	me.Shift = code&4 != 0
	me.Action, me.Button = decodeMouseButton(code, terminator)

	return Event{Kind: EventMouse, Mouse: me}, nil
}

func decodeMouseButton(code int, terminator byte) (MouseAction, MouseButton) {
	// Strip modifier bits (shift=4, meta=8, ctrl=16) for button identification.
	base := code &^ (4 | 8 | 16)

	if terminator == 'm' {
		return MouseRelease, MouseButton(base & 0x03)
	}

	// Motion flag: bit 5 (32)
	if base&32 != 0 {
		return MouseMotion, MouseButton(base & ^32)
	}

	// Scroll: button codes 64, 65
	if base >= 64 && base <= 67 {
		return MouseScroll, MouseButton(base)
	}

	return MousePress, MouseButton(base & 0x03)
}
