package input

import (
	"io"
	"strconv"
	"strings"
)

// parseOSC consumes an OSC sequence (ESC ] ... terminated by BEL or ESC \).
// An OSC 11 colour report becomes an EventScheme; anything else is swallowed
// and the next event is returned.
func parseOSC(r io.Reader) (Event, error) {
	body, err := readOSCBody(r)
	if err != nil {
		return Event{Kind: EventSpecial, Special: KeyEscape}, nil
	}
	if scheme, ok := schemeFromOSC11(body); ok {
		return Event{Kind: EventScheme, Scheme: scheme}, nil
	}
	return ReadEvent(r)
}

// readOSCBody reads until BEL (0x07) or ST (ESC \).
func readOSCBody(r io.Reader) (string, error) {
	var b strings.Builder
	for {
		ch, err := readByte(r)
		if err != nil {
			return "", err
		}
		switch ch {
		case 0x07:
			return b.String(), nil
		case 0x1b:
			next, err := readByte(r)
			if err != nil || next == '\\' {
				return b.String(), nil
			}
			b.WriteByte(ch)
			b.WriteByte(next)
		default:
			b.WriteByte(ch)
		}
	}
}

// schemeFromOSC11 parses an OSC 11 background-colour report
// ("11;rgb:RRRR/GGGG/BBBB") into "light" or "dark" by luminance.
func schemeFromOSC11(body string) (string, bool) {
	rest, ok := strings.CutPrefix(body, "11;")
	if !ok {
		return "", false
	}
	rest, ok = strings.CutPrefix(rest, "rgb:")
	if !ok {
		return "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 3 {
		return "", false
	}
	var ch [3]float64
	for i, p := range parts {
		n, err := strconv.ParseUint(p, 16, 32)
		if err != nil {
			return "", false
		}
		max := float64(uint64(1)<<(4*len(p)) - 1)
		ch[i] = float64(n) / max
	}
	luminance := 0.299*ch[0] + 0.587*ch[1] + 0.114*ch[2]
	if luminance > 0.5 {
		return "light", true
	}
	return "dark", true
}
