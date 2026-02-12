package input

import "io"

// EventKind distinguishes key events from special key events and mouse events.
type EventKind int

const (
	EventKey     EventKind = iota // regular character key
	EventSpecial                  // special key (arrow, pgup, etc.)
	EventMouse                    // mouse event
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
	KeyTab      SpecialKey = "tab"
	KeyShiftTab SpecialKey = "shift-tab"
)

// Event represents a terminal input event.
type Event struct {
	Kind    EventKind
	Rune    rune       // set for EventKey
	Special SpecialKey // set for EventSpecial
	Mouse   MouseEvent // set for EventMouse
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
		return Event{Kind: EventKey, Rune: 0x1b}, nil
	}

	if b != '[' {
		// ESC followed by something other than [
		return Event{Kind: EventKey, Rune: 0x1b}, nil
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

// parseNumericCSI handles ESC[N~ sequences.
func parseNumericCSI(r io.Reader, firstDigit byte) (Event, error) {
	// Read until we get a ~ or other terminator
	num := int(firstDigit - '0')
	for {
		b, err := readByte(r)
		if err != nil {
			return Event{Kind: EventKey, Rune: 0x1b}, nil
		}
		if b == '~' {
			break
		}
		if b >= '0' && b <= '9' {
			num = num*10 + int(b-'0')
			continue
		}
		// Unknown terminator
		return Event{Kind: EventKey, Rune: 0x1b}, nil
	}

	if special, ok := numericCSIKey(num); ok {
		return Event{Kind: EventSpecial, Special: special}, nil
	}
	return Event{Kind: EventKey, Rune: 0x1b}, nil
}

func numericCSIKey(num int) (SpecialKey, bool) {
	switch num {
	case 5:
		return KeyPgUp, true
	case 6:
		return KeyPgDn, true
	case 1:
		return KeyHome, true
	case 4:
		return KeyEnd, true
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
