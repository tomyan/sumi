package vt100_test

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/vt100"
)

func TestParseModeSequencesIgnored(t *testing.T) {
	// Given — a stream with alt screen, cursor hide, mouse mode, paste mode
	screen := vt100.NewScreen(10, 1)
	stream := "" +
		"\x1b[?1049h" + // alt screen on
		"\x1b[?25l" + // cursor hide
		"\x1b[?1003h" + // mouse mode on
		"\x1b[?1006h" + // SGR mouse on
		"\x1b[?2004h" + // bracketed paste on
		"\x1b[1;1HHi" + // write text
		"\x1b[?1003l" + // mouse mode off
		"\x1b[?1006l" + // SGR mouse off
		"\x1b[?2004l" + // bracketed paste off
		"\x1b[?25h" + // cursor show
		"\x1b[?1049l" // alt screen off

	// When
	_, err := screen.Write([]byte(stream))

	// Then — mode sequences consumed without error, text written correctly
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := screen.Cell(0, 0).Ch; got != 'H' {
		t.Errorf("cell (0,0): got %q, want 'H'", got)
	}
	if got := screen.Cell(0, 1).Ch; got != 'i' {
		t.Errorf("cell (0,1): got %q, want 'i'", got)
	}
}

func TestParseOSCTitleIgnored(t *testing.T) {
	// Given — OSC 2 (set title) sequence surrounding text
	screen := vt100.NewScreen(10, 1)
	stream := "\x1b]2;My App Title\x07" + "\x1b[1;1HOK"

	// When
	_, err := screen.Write([]byte(stream))

	// Then — title consumed, text written
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := screen.Cell(0, 0).Ch; got != 'O' {
		t.Errorf("cell (0,0): got %q, want 'O'", got)
	}
}

func TestParseSavRestoreTitleIgnored(t *testing.T) {
	// Given — save/restore title sequences
	screen := vt100.NewScreen(10, 1)
	stream := "\x1b[22;2t" + "\x1b[1;1HAB" + "\x1b[23;2t"

	// When
	_, err := screen.Write([]byte(stream))

	// Then — consumed, text written
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := screen.Cell(0, 0).Ch; got != 'A' {
		t.Errorf("cell (0,0): got %q, want 'A'", got)
	}
}

func TestParseSentinelDetected(t *testing.T) {
	// Given — content + sentinel
	screen := vt100.NewScreen(10, 1)
	stream := "\x1b[1;1HX" +
		"\x1b]999;done\x07" // sentinel

	// When
	_, err := screen.Write([]byte(stream))

	// Then — text written, sentinel detected
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := screen.Cell(0, 0).Ch; got != 'X' {
		t.Errorf("cell (0,0): got %q, want 'X'", got)
	}
	if !screen.SentinelSeen() {
		t.Error("expected sentinel to be detected")
	}
}

func TestParseSentinelNotDetectedWithoutOSC(t *testing.T) {
	// Given — content without sentinel
	screen := vt100.NewScreen(10, 1)

	// When
	screen.Write([]byte("\x1b[1;1HX"))

	// Then
	if screen.SentinelSeen() {
		t.Error("expected no sentinel")
	}
}

func TestParseFullRoundTripWithModes(t *testing.T) {
	// Given — a styled buffer rendered with full screen setup
	buf := render.NewBuffer(10, 2)
	buf.WriteStyledText(0, 0, "Hello", render.Style{Bold: true})
	buf.WriteText(1, 0, "World")

	// When — wrap with alt screen + clear + sentinel, parse back
	screen := vt100.NewScreen(10, 2)
	screen.Write([]byte("\x1b[?1049h\x1b[?25l\x1b[?1003h\x1b[?1006h\x1b[?2004h"))
	screen.Write([]byte("\x1b[2J\x1b[H"))

	var ansi []byte
	ansi = append(ansi, renderToBytes(buf)...)
	ansi = append(ansi, []byte("\x1b]999;done\x07")...)
	screen.Write(ansi)

	// Then — cells match, sentinel seen
	assertFullCellsMatch(t, buf, screen)
	if !screen.SentinelSeen() {
		t.Error("expected sentinel")
	}
}

func renderToBytes(buf *render.Buffer) []byte {
	var b []byte
	w := &byteWriter{buf: &b}
	buf.RenderTo(w)
	return *w.buf
}

type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
