package input

import (
	"io"
	"syscall"
)

// EventKind distinguishes key events from special key events and mouse events.
type EventKind int

const (
	EventKey     EventKind = iota // regular character key
	EventSpecial                  // special key (arrow, pgup, etc.)
	EventMouse                    // mouse event
	EventSignal                   // OS signal (SIGINT, SIGTERM, etc.)
	EventFrame                    // animation frame tick
)

// SpecialKey identifies a special key.
type SpecialKey string

const (
	KeyUp       SpecialKey = "up"
	KeyDown     SpecialKey = "down"
	KeyLeft     SpecialKey = "left"
	KeyRight    SpecialKey = "right"
	KeyPgUp     SpecialKey = "pgup"
	KeyPgDn     SpecialKey = "pgdn"
	KeyHome     SpecialKey = "home"
	KeyEnd      SpecialKey = "end"
	KeyTab       SpecialKey = "tab"
	KeyShiftTab  SpecialKey = "shift-tab"
	KeyEnter     SpecialKey = "enter"
	KeyEscape    SpecialKey = "escape"
	KeyBackspace SpecialKey = "backspace"
	KeyDelete    SpecialKey = "delete"
)

// Event represents a terminal input event.
type Event struct {
	Kind    EventKind
	Rune    rune           // set for EventKey
	Ctrl    bool           // true if Ctrl modifier was held
	Shift   bool           // true if Shift modifier was held
	Alt     bool           // true if Alt modifier was held
	Special SpecialKey     // set for EventSpecial
	Mouse   MouseEvent     // set for EventMouse
	Signal  syscall.Signal // set for EventSignal
}

// ReadEvent reads a single input event from the reader.
// Parses escape sequences for arrow keys, PgUp/PgDn, Home/End.
// Uses ReadKey for the initial character to handle multi-byte UTF-8.
func ReadEvent(r io.Reader) (Event, error) {
	ch, err := ReadKey(r)
	if err != nil {
		return Event{}, err
	}

	if ch == '\t' {
		return Event{Kind: EventSpecial, Special: KeyTab}, nil
	}
	if ch == '\r' || ch == '\n' {
		return Event{Kind: EventSpecial, Special: KeyEnter}, nil
	}
	if ch == 127 || ch == 8 {
		return Event{Kind: EventSpecial, Special: KeyBackspace}, nil
	}

	// Control characters (Ctrl+A through Ctrl+Z, excluding Tab and Enter)
	if ch >= 1 && ch <= 26 {
		return Event{Kind: EventKey, Rune: 'a' + ch - 1, Ctrl: true}, nil
	}
	// Ctrl+/ and Ctrl+_ both send 0x1F
	if ch == 0x1f {
		return Event{Kind: EventKey, Rune: '/', Ctrl: true}, nil
	}

	if ch != 0x1b {
		return Event{Kind: EventKey, Rune: ch}, nil
	}

	return parseEscapeSequence(r)
}

// parseEscapeSequence handles input starting with ESC (0x1b).
func parseEscapeSequence(r io.Reader) (Event, error) {
	b, err := readByte(r)
	if err != nil {
		// bare ESC with no following bytes
		return Event{Kind: EventSpecial, Special: KeyEscape}, nil
	}

	if b != '[' {
		// ESC followed by a printable character is Alt+key
		if b >= 32 && b < 127 {
			return Event{Kind: EventKey, Rune: rune(b), Alt: true}, nil
		}
		return Event{Kind: EventSpecial, Special: KeyEscape}, nil
	}

	return parseCSISequence(r)
}

// parseCSISequence handles CSI sequences (ESC [ ...).
func parseCSISequence(r io.Reader) (Event, error) {
	b, err := readByte(r)
	if err != nil {
		return Event{Kind: EventKey, Rune: 0x1b}, nil
	}

	// SGR mouse: ESC[<...
	if b == '<' {
		return parseSGRMouse(func() (byte, error) { return readByte(r) }, "")
	}

	// Single letter final byte: arrow keys, Home, End
	if special, ok := singleByteCSI(b); ok {
		return Event{Kind: EventSpecial, Special: special}, nil
	}

	// Numeric sequences like ESC[5~ (PgUp), ESC[6~ (PgDn)
	if b >= '0' && b <= '9' {
		return parseNumericCSI(r, b)
	}

	return Event{Kind: EventKey, Rune: 0x1b}, nil
}

func singleByteCSI(b byte) (SpecialKey, bool) {
	switch b {
	case 'A':
		return KeyUp, true
	case 'B':
		return KeyDown, true
	case 'C':
		return KeyRight, true
	case 'D':
		return KeyLeft, true
	case 'H':
		return KeyHome, true
	case 'F':
		return KeyEnd, true
	case 'Z':
		return KeyShiftTab, true
	default:
		return "", false
	}
}

// parseNumericCSI handles ESC[N~ and ESC[N;M<final> sequences.
// The latter encodes modifier keys: M is (1+bitmask) where bit 0=Shift, bit 1=Alt, bit 2=Ctrl.
func parseNumericCSI(r io.Reader, firstDigit byte) (Event, error) {
	num := int(firstDigit - '0')
	for {
		b, err := readByte(r)
		if err != nil {
			return Event{Kind: EventKey, Rune: 0x1b}, nil
		}
		if b == '~' {
			if special, ok := numericCSIKey(num); ok {
				return Event{Kind: EventSpecial, Special: special}, nil
			}
			return Event{Kind: EventKey, Rune: 0x1b}, nil
		}
		if b == ';' {
			return parseModifiedCSI(r, num)
		}
		if b >= '0' && b <= '9' {
			num = num*10 + int(b-'0')
			continue
		}
		return Event{Kind: EventKey, Rune: 0x1b}, nil
	}
}

// parseModifiedCSI handles CSI sequences with modifier: ESC[num;modifier<final>.
// The final byte determines the key; modifier encodes Shift/Alt/Ctrl.
func parseModifiedCSI(r io.Reader, num int) (Event, error) {
	// Read modifier number
	modifier := 0
	for {
		b, err := readByte(r)
		if err != nil {
			return Event{Kind: EventKey, Rune: 0x1b}, nil
		}
		if b >= '0' && b <= '9' {
			modifier = modifier*10 + int(b-'0')
			continue
		}
		// b is the final byte (or ~)
		shift, alt, ctrl := decodeModifier(modifier)
		if b == '~' {
			if special, ok := numericCSIKey(num); ok {
				return Event{Kind: EventSpecial, Special: special, Shift: shift, Alt: alt, Ctrl: ctrl}, nil
			}
			return Event{Kind: EventKey, Rune: 0x1b}, nil
		}
		if special, ok := singleByteCSI(b); ok {
			return Event{Kind: EventSpecial, Special: special, Shift: shift, Alt: alt, Ctrl: ctrl}, nil
		}
		return Event{Kind: EventKey, Rune: 0x1b}, nil
	}
}

// decodeModifier decodes the CSI modifier parameter into flags.
// Modifier param = 1 + bitmask: bit 0=Shift, bit 1=Alt, bit 2=Ctrl.
func decodeModifier(modifier int) (shift, alt, ctrl bool) {
	bits := modifier - 1
	shift = bits&1 != 0
	alt = bits&2 != 0
	ctrl = bits&4 != 0
	return
}

func numericCSIKey(num int) (SpecialKey, bool) {
	switch num {
	case 1:
		return KeyHome, true
	case 3:
		return KeyDelete, true
	case 4:
		return KeyEnd, true
	case 5:
		return KeyPgUp, true
	case 6:
		return KeyPgDn, true
	default:
		return "", false
	}
}

func readByte(r io.Reader) (byte, error) {
	var buf [1]byte
	n, err := r.Read(buf[:])
	if n == 0 {
		return 0, err
	}
	return buf[0], nil
}
