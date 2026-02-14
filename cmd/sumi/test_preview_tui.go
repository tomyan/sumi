package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	client    *sumitest.Client
	master    *os.File
	info      *sumitest.InfoResponse
	screen    *vt100.Screen
	current   int
	actual    string           // styled text from step response
	snapshots []sumitest.Frame // loaded from snapshot file
	source    *sourcePanel
	w         io.Writer
}

func runPreviewTUI(client *sumitest.Client, master *os.File, info *sumitest.InfoResponse, compDir string) error {
	tui := &previewTUI{
		client: client,
		master: master,
		info:   info,
		screen: vt100.NewScreen(info.Width, info.Height),
		w:      os.Stdout,
	}
	tui.loadSnapshots(compDir)
	tui.source = loadSource(compDir, info.SourceFile)
	return tui.run()
}

func (p *previewTUI) loadSnapshots(compDir string) {
	path := filepath.Join(compDir, "testdata", p.info.Name+".snapshot")
	frames, err := sumitest.ReadSnapshot(path)
	if err != nil {
		return // no snapshot file — that's fine
	}
	p.snapshots = frames
}

func (p *previewTUI) run() error {
	restore, _ := input.EnableRawMode(int(os.Stdin.Fd()))
	defer restore()
	render.EnterAlternateScreen(p.w)
	defer render.ExitAlternateScreen(p.w)

	if err := p.stepTo(0); err != nil {
		return fmt.Errorf("initial step: %w", err)
	}

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
		case isScrollDown(evt):
			if p.source != nil {
				p.source.scrollDown(scrollAmount(evt))
				p.render()
			}
		case isScrollUp(evt):
			if p.source != nil {
				p.source.scrollUp(scrollAmount(evt))
				p.render()
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
	p.actual = resp.StyledText

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

	stepName := p.info.Steps[p.current]
	match := p.snapshotMatches()
	statusIcon := "?"
	if len(p.snapshots) > 0 {
		if match {
			statusIcon = "✓"
		} else {
			statusIcon = "✗"
		}
	}

	// Header.
	header := fmt.Sprintf(" %s  |  Frame %d/%d  |  %s  [%s]",
		p.info.Name, p.current+1, len(p.info.Steps), stepName, statusIcon)
	render.WriteAt(p.w, 0, 0, header, render.Style{Bold: true})
	render.WriteAt(p.w, 1, 0, strings.Repeat("─", min(termW, 80)), render.Style{Dim: true})

	// Side-by-side panels.
	panelW := p.info.Width + 2
	p.renderActualPanel(headerRows, 0, panelW)
	p.renderExpectedPanel(headerRows, panelW+1)

	// Source panel.
	sourceStart := headerRows + p.info.Height + 1
	if p.source != nil {
		p.source.renderSource(p.w, sourceStart)
	} else {
		// Footer when no source.
		render.WriteAt(p.w, sourceStart, 0, strings.Repeat("─", min(termW, 80)), render.Style{Dim: true})
		controls := " ←/→ step  |  q quit"
		render.WriteAt(p.w, sourceStart+1, 0, controls, render.Style{Dim: true})
	}
}

func (p *previewTUI) renderActualPanel(startRow, startCol, width int) {
	// Label.
	render.WriteAt(p.w, startRow-1, startCol, " Actual", render.Style{Dim: true})

	// Render VT100 screen cells at offset.
	compBuf := p.screen.Buffer()
	compBuf.RenderToOffset(p.w, startRow-1, startCol)
}

func (p *previewTUI) renderExpectedPanel(startRow, startCol int) {
	render.WriteAt(p.w, startRow-1, startCol, " Expected", render.Style{Dim: true})

	expected := p.expectedText()
	if expected == "" {
		render.WriteAt(p.w, startRow, startCol+1, "(no snapshot)", render.Style{Dim: true})
		return
	}
	lines := strings.Split(expected, "\n")
	for i, line := range lines {
		if i >= p.info.Height {
			break
		}
		render.WriteAt(p.w, startRow+i, startCol+1, line, render.Style{})
	}
}

func (p *previewTUI) expectedText() string {
	if p.current >= len(p.snapshots) {
		return ""
	}
	return p.snapshots[p.current].StyledText
}

func (p *previewTUI) snapshotMatches() bool {
	expected := p.expectedText()
	if expected == "" {
		return false
	}
	return p.actual == expected
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
