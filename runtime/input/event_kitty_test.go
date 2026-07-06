package input

import (
	"strings"
	"testing"
)

// D3: kitty keyboard protocol — CSI <codepoint>(;<mod>(:<event>)?)? u
// key reports decode alongside the legacy paths. Flag 1 (disambiguate
// escape codes) only: plain arrows/functionals keep arriving legacy.

func readSeq(t *testing.T, seq string) Event {
	t.Helper()
	evt, err := ReadEvent(strings.NewReader(seq))
	if err != nil {
		t.Fatalf("read %q: %v", seq, err)
	}
	return evt
}

func TestKittyPrintableKey(t *testing.T) {
	evt := readSeq(t, "\x1b[97u")
	if evt.Kind != EventKey || evt.Rune != 'a' || evt.Ctrl || evt.Shift || evt.Alt {
		t.Errorf("evt = %+v, want plain 'a'", evt)
	}
}

func TestKittyCtrlLetter(t *testing.T) {
	evt := readSeq(t, "\x1b[99;5u")
	if evt.Kind != EventKey || evt.Rune != 'c' || !evt.Ctrl {
		t.Errorf("evt = %+v, want Ctrl+c", evt)
	}
}

func TestKittyShiftAndAltPrintables(t *testing.T) {
	if evt := readSeq(t, "\x1b[97;2u"); !evt.Shift || evt.Rune != 'a' {
		t.Errorf("evt = %+v, want Shift+a", evt)
	}
	if evt := readSeq(t, "\x1b[97;3u"); !evt.Alt || evt.Rune != 'a' {
		t.Errorf("evt = %+v, want Alt+a", evt)
	}
}

func TestKittyFunctionalKeys(t *testing.T) {
	cases := []struct {
		seq  string
		want SpecialKey
	}{
		{"\x1b[13u", KeyEnter},
		{"\x1b[27u", KeyEscape},
		{"\x1b[9u", KeyTab},
		{"\x1b[127u", KeyBackspace},
		{"\x1b[57349u", KeyDelete},
		{"\x1b[57354u", KeyPgUp},
		{"\x1b[57355u", KeyPgDn},
		{"\x1b[57356u", KeyHome},
		{"\x1b[57357u", KeyEnd},
	}
	for _, c := range cases {
		evt := readSeq(t, c.seq)
		if evt.Kind != EventSpecial || evt.Special != c.want {
			t.Errorf("%q = %+v, want %s", c.seq, evt, c.want)
		}
	}
}

func TestKittyCtrlEnter(t *testing.T) {
	// The combo legacy cannot express — the point of flag 1.
	evt := readSeq(t, "\x1b[13;5u")
	if evt.Kind != EventSpecial || evt.Special != KeyEnter || !evt.Ctrl {
		t.Errorf("evt = %+v, want Ctrl+Enter", evt)
	}
}

func TestKittyShiftTabKeepsConvention(t *testing.T) {
	evt := readSeq(t, "\x1b[9;2u")
	if evt.Kind != EventSpecial || evt.Special != KeyShiftTab {
		t.Errorf("evt = %+v, want shift-tab", evt)
	}
}

func TestKittyEventTypeSubParamIgnored(t *testing.T) {
	// CSI 13;5:1u — press event type; sub-parameter is discarded.
	evt := readSeq(t, "\x1b[13;5:1u")
	if evt.Kind != EventSpecial || evt.Special != KeyEnter || !evt.Ctrl {
		t.Errorf("evt = %+v, want Ctrl+Enter (event type ignored)", evt)
	}
}

func TestLegacySequencesStillDecode(t *testing.T) {
	// Flag 1 leaves plain arrows/functionals on the legacy encoding.
	if evt := readSeq(t, "\x1b[A"); evt.Special != KeyUp {
		t.Errorf("legacy up = %+v", evt)
	}
	if evt := readSeq(t, "\x1b[5~"); evt.Special != KeyPgUp {
		t.Errorf("legacy pgup = %+v", evt)
	}
}
