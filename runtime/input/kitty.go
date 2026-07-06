package input

// Kitty keyboard protocol, progressive-enhancement flag 1 (disambiguate
// escape codes) only: key reports arrive as
// CSI <codepoint> (';' <modifier> (':' <event-type>)?)? 'u'.
// Codepoint and modifier are consumed; the event-type sub-parameter is
// discarded (press/release reporting is not requested). Plain arrows
// and functionals keep arriving in the legacy encoding under flag 1.

// KittyEnableSeq pushes disambiguate-escape-codes mode. Terminals
// without kitty support ignore the push and keep the legacy encoding —
// that is the whole fallback strategy (no capability query).
const KittyEnableSeq = "\x1b[>1u"

// KittyDisableSeq pops our entry off the terminal's mode stack.
const KittyDisableSeq = "\x1b[<u"

// kittyFunctional maps kitty functional-key codepoints (ASCII controls
// plus private-use-area codes) to special keys.
var kittyFunctional = map[int]SpecialKey{
	9:     KeyTab,
	13:    KeyEnter,
	27:    KeyEscape,
	127:   KeyBackspace,
	57349: KeyDelete,
	57354: KeyPgUp,
	57355: KeyPgDn,
	57356: KeyHome,
	57357: KeyEnd,
}

// kittyKeyEvent builds the event for a kitty key report. The modifier
// parameter uses the standard bitmask+1 encoding (1 = none).
func kittyKeyEvent(codepoint, modifier int) Event {
	shift, alt, ctrl := decodeModifier(modifier)
	if special, ok := kittyFunctional[codepoint]; ok {
		if special == KeyTab && shift {
			return Event{Kind: EventSpecial, Special: KeyShiftTab, Alt: alt, Ctrl: ctrl}
		}
		return Event{Kind: EventSpecial, Special: special, Shift: shift, Alt: alt, Ctrl: ctrl}
	}
	return Event{Kind: EventKey, Rune: rune(codepoint), Shift: shift, Alt: alt, Ctrl: ctrl}
}
