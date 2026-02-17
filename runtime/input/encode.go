package input

import "unicode/utf8"

// EncodeEvent converts an Event back to the raw terminal byte sequence.
// This is the inverse of ReadEvent — encoding events so they can be
// forwarded to a subprocess PTY.
func EncodeEvent(evt Event) []byte {
	switch evt.Kind {
	case EventKey:
		return encodeKeyEvent(evt)
	case EventSpecial:
		return encodeSpecialKey(evt.Special)
	case EventPaste:
		return encodePaste(evt.PasteText)
	default:
		return nil
	}
}

func encodeKeyEvent(evt Event) []byte {
	if evt.Ctrl {
		return encodeCtrlKey(evt)
	}
	if evt.Alt {
		// Alt+key sends ESC followed by the key byte(s).
		inner := encodeRune(evt.Rune)
		buf := make([]byte, 0, 1+len(inner))
		buf = append(buf, 0x1b)
		buf = append(buf, inner...)
		return buf
	}
	return encodeRune(evt.Rune)
}

func encodeCtrlKey(evt Event) []byte {
	switch evt.Rune {
	case '/':
		return []byte{0x1f}
	case '\\':
		return []byte{0x1c}
	default:
		if evt.Rune >= 'a' && evt.Rune <= 'z' {
			return []byte{byte(evt.Rune - 'a' + 1)}
		}
		// Fallback: just send the rune
		return encodeRune(evt.Rune)
	}
}

func encodeRune(r rune) []byte {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return buf[:n]
}

var specialKeySeqs = map[SpecialKey][]byte{
	KeyUp:        {0x1b, '[', 'A'},
	KeyDown:      {0x1b, '[', 'B'},
	KeyRight:     {0x1b, '[', 'C'},
	KeyLeft:      {0x1b, '[', 'D'},
	KeyHome:      {0x1b, '[', 'H'},
	KeyEnd:       {0x1b, '[', 'F'},
	KeyPgUp:      {0x1b, '[', '5', '~'},
	KeyPgDn:      {0x1b, '[', '6', '~'},
	KeyDelete:    {0x1b, '[', '3', '~'},
	KeyTab:       {'\t'},
	KeyShiftTab:  {0x1b, '[', 'Z'},
	KeyEnter:     {'\r'},
	KeyEscape:    {0x1b},
	KeyBackspace: {0x7f},
}

func encodeSpecialKey(key SpecialKey) []byte {
	if seq, ok := specialKeySeqs[key]; ok {
		return seq
	}
	return nil
}

func encodePaste(text string) []byte {
	prefix := "\x1b[200~"
	suffix := "\x1b[201~"
	buf := make([]byte, 0, len(prefix)+len(text)+len(suffix))
	buf = append(buf, prefix...)
	buf = append(buf, text...)
	buf = append(buf, suffix...)
	return buf
}
