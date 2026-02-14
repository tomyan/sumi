package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/vt100"
)

const headerRows = 3

type previewTUI struct {
	client  *sumitest.Client
	master  *os.File
	info    *sumitest.InfoResponse
	screen  *vt100.Screen
	current int // current step index
	w       io.Writer
}

func runPreviewTUI(client *sumitest.Client, master *os.File, info *sumitest.InfoResponse) error {
	tui := &previewTUI{
		client: client,
		master: master,
		info:   info,
		screen: vt100.NewScreen(info.Width, info.Height),
		w:      os.Stdout,
	}
	return tui.run()
}

func (p *previewTUI) run() error {
	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(p.w)
	defer render.ExitAlternateScreen(p.w)

	// Render initial step.
	if err := p.stepTo(0); err != nil {
		return fmt.Errorf("initial step: %w", err)
	}

	// Input loop.
	for {
		evt, err := input.ReadEvent(os.Stdin)
		if err != nil {
			return nil
		}

		switch {
		case isQuitKey(evt):
			return nil
		case isNextKey(evt):
			if p.current < len(p.info.Steps)-1 {
				if err := p.stepTo(p.current + 1); err != nil {
					return err
				}
			}
		case isPrevKey(evt):
			if p.current > 0 {
				if err := p.stepTo(p.current - 1); err != nil {
					return err
				}
			}
		}
	}
}

func (p *previewTUI) stepTo(index int) error {
	p.screen.ResetSentinel()

	resp, err := p.client.Step(index)
	if err != nil {
		return fmt.Errorf("step %d: %w", index, err)
	}
	_ = resp

	if err := p.readUntilSentinel(); err != nil {
		return fmt.Errorf("read pty: %w", err)
	}

	p.current = index
	p.render()
	return nil
}

func (p *previewTUI) readUntilSentinel() error {
	buf := make([]byte, 4096)
	deadline := time.Now().Add(5 * time.Second)

	for !p.screen.SentinelSeen() {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for sentinel")
		}
		p.master.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := p.master.Read(buf)
		if n > 0 {
			p.screen.Write(buf[:n])
		}
		if err != nil && !isTimeout(err) {
			return err
		}
	}
	return nil
}

func (p *previewTUI) render() {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	render.ClearScreen(p.w)

	// Header.
	stepName := p.info.Steps[p.current]
	header := fmt.Sprintf(" %s  |  Frame %d/%d  |  %s",
		p.info.Name, p.current+1, len(p.info.Steps), stepName)
	render.WriteAt(p.w, 0, 0, header, render.Style{Bold: true})
	render.WriteAt(p.w, 1, 0, strings.Repeat("─", min(termW, 60)), render.Style{Dim: true})

	// Component area — copy VT100 screen cells to terminal at offset.
	compBuf := p.screen.Buffer()
	compBuf.RenderToOffset(p.w, headerRows-1, 0)

	// Footer.
	footerRow := headerRows + p.info.Height
	render.WriteAt(p.w, footerRow, 0, strings.Repeat("─", min(termW, 60)), render.Style{Dim: true})
	controls := " ←/→ step  |  q quit"
	render.WriteAt(p.w, footerRow+1, 0, controls, render.Style{Dim: true})
}

func isQuitKey(evt input.Event) bool {
	return evt.Kind == input.EventKey && evt.Rune == 'q'
}

func isNextKey(evt input.Event) bool {
	if evt.Kind == input.EventKey && (evt.Rune == '\r' || evt.Rune == '\n' || evt.Rune == 'l') {
		return true
	}
	if evt.Kind == input.EventSpecial && evt.Special == input.KeyRight {
		return true
	}
	return false
}

func isPrevKey(evt input.Event) bool {
	if evt.Kind == input.EventKey && evt.Rune == 'h' {
		return true
	}
	if evt.Kind == input.EventSpecial && evt.Special == input.KeyLeft {
		return true
	}
	return false
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
