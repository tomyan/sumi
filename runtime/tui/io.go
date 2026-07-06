package tui

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// out returns the terminal writer (injected Out or stdout).
func (a *App) out() io.Writer {
	if a.Out != nil {
		return a.Out
	}
	return os.Stdout
}

// in returns the terminal reader (injected In or stdin).
func (a *App) in() io.Reader {
	if a.In != nil {
		return a.In
	}
	return os.Stdin
}

// outFd returns the file descriptor behind the output when it is a real
// file (terminal size queries need one). ok is false for injected
// writers like buffers or pipes without a descriptor.
func (a *App) outFd() (int, bool) {
	if f, ok := a.out().(*os.File); ok {
		return int(f.Fd()), true
	}
	return 0, false
}

// inFd returns the descriptor behind the input for raw-mode control.
func (a *App) inFd() (int, bool) {
	if f, ok := a.in().(*os.File); ok {
		return int(f.Fd()), true
	}
	return 0, false
}

// terminalSize returns the viewport dimensions: test overrides first,
// then the output terminal's real size, then an 80x24 fallback for
// injected writers without a descriptor.
func (a *App) terminalSize() (int, int) {
	if a.TestWidth > 0 {
		return a.TestWidth, a.TestHeight
	}
	if fd, ok := a.outFd(); ok {
		return term.GetSize(fd)
	}
	return 80, 24
}

// writeOSC52 writes the in-band OSC 52 clipboard sequence to the
// terminal writer.
func (a *App) writeOSC52(text string) {
	render.CopyToClipboard(a.out(), text)
}

// captureLogs routes the stdlib log package to onLog (one call per
// line) while a fullscreen app owns the terminal — direct writes would
// corrupt the frame. The returned func restores logging to stderr.
func captureLogs(onLog func(string)) func() {
	log.SetOutput(&lineWriter{onLog: onLog})
	return func() { log.SetOutput(os.Stderr) }
}

// lineWriter splits writes into lines for the onLog callback.
type lineWriter struct {
	onLog func(string)
	buf   bytes.Buffer
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	for {
		line, err := w.buf.ReadString('\n')
		if err != nil {
			w.buf.WriteString(line) // incomplete line — keep for later
			break
		}
		w.onLog(line[:len(line)-1])
	}
	return len(p), nil
}
