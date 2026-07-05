package input

import (
	"strings"
	"testing"
)

// A9b: OSC 11 colour-scheme reports.

func TestOSC11DarkBackground(t *testing.T) {
	// Given: a dark background report, BEL-terminated
	r := strings.NewReader("\x1b]11;rgb:1e1e/1e1e/1e1e\x07")

	// When
	evt, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("ReadEvent: %v", err)
	}
	if evt.Kind != EventScheme || evt.Scheme != "dark" {
		t.Errorf("evt = %+v, want dark scheme", evt)
	}
}

func TestOSC11LightBackgroundSTTerminated(t *testing.T) {
	r := strings.NewReader("\x1b]11;rgb:ffff/ffff/f0f0\x1b\\")
	evt, err := ReadEvent(r)
	if err != nil {
		t.Fatalf("ReadEvent: %v", err)
	}
	if evt.Kind != EventScheme || evt.Scheme != "light" {
		t.Errorf("evt = %+v, want light scheme", evt)
	}
}

func TestOSC11EightBitChannels(t *testing.T) {
	r := strings.NewReader("\x1b]11;rgb:ff/ff/ff\x07")
	evt, _ := ReadEvent(r)
	if evt.Scheme != "light" {
		t.Errorf("evt = %+v, want light", evt)
	}
}

func TestUnknownOSCSwallowedNextEventReturned(t *testing.T) {
	// Given: an unrelated OSC report followed by a keypress
	r := strings.NewReader("\x1b]52;c;aGk=\x07x")

	// When
	evt, err := ReadEvent(r)

	// Then
	if err != nil {
		t.Fatalf("ReadEvent: %v", err)
	}
	if evt.Kind != EventKey || evt.Rune != 'x' {
		t.Errorf("evt = %+v, want the following keypress", evt)
	}
}
