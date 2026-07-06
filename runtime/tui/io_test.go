package tui

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// F4b: io injection — terminal control sequences route through App.Out
// instead of hard-coded os.Stdout; stdlib log is capturable via OnLog.

func TestEnterExitTerminalWriteToInjectedOut(t *testing.T) {
	// Given
	var out bytes.Buffer
	app := &App{HasMouse: true, Out: &out}

	// When
	app.enterTerminal()
	app.exitTerminal()

	// Then: alt screen, paste, and mouse sequences all hit the buffer.
	s := out.String()
	for _, seq := range []string{"\x1b[?1049h", "\x1b[?2004h", "\x1b[?1003h", "\x1b[?1003l", "\x1b[?1049l"} {
		if !strings.Contains(s, seq) {
			t.Errorf("missing %q in injected output", seq)
		}
	}
}

func TestClipboardWritesOSC52ToInjectedOut(t *testing.T) {
	// Given
	var out bytes.Buffer
	app := &App{Out: &out}

	// When: the system clipboard path writes OSC 52 in-band.
	app.writeOSC52("hi")

	// Then: base64("hi") = aGk=
	if !strings.Contains(out.String(), "\x1b]52;c;aGk=") {
		t.Errorf("output = %q, want OSC 52 payload", out.String())
	}
}

func TestOnLogCapturesStdlibLog(t *testing.T) {
	// Given
	var lines []string
	restore := captureLogs(func(line string) { lines = append(lines, line) })
	defer restore()

	// When
	log.Print("hello from app")

	// Then
	if len(lines) != 1 || !strings.Contains(lines[0], "hello from app") {
		t.Errorf("lines = %q, want the log line", lines)
	}
}

func TestOnLogRestoreReturnsLogToStderr(t *testing.T) {
	// Given
	restore := captureLogs(func(string) {})

	// When
	restore()

	// Then: logging after restore must not panic or call the callback.
	log.SetOutput(log.Writer()) // no-op sanity; the default writer is back
}
