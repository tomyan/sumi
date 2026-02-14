package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// sourcePanel manages a scrollable source code view.
type sourcePanel struct {
	lines     []string
	scrollOff int
	path      string // display path
}

// loadSource reads the source file and populates the panel.
func loadSource(compDir, relPath string) *sourcePanel {
	if relPath == "" {
		return nil
	}
	absPath := filepath.Join(compDir, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	return &sourcePanel{lines: lines, path: relPath}
}

// visibleLines returns the number of source lines that fit on screen.
func (sp *sourcePanel) visibleLines(startRow int) int {
	_, termH := term.GetSize(int(os.Stdout.Fd()))
	available := termH - startRow - 1 // leave room for scroll hint
	if available < 1 {
		return 1
	}
	return available
}

// scrollDown moves the view down by n lines.
func (sp *sourcePanel) scrollDown(n int) {
	sp.scrollOff += n
	maxOff := len(sp.lines) - 1
	if maxOff < 0 {
		maxOff = 0
	}
	if sp.scrollOff > maxOff {
		sp.scrollOff = maxOff
	}
}

// scrollUp moves the view up by n lines.
func (sp *sourcePanel) scrollUp(n int) {
	sp.scrollOff -= n
	if sp.scrollOff < 0 {
		sp.scrollOff = 0
	}
}

// renderSource draws the source panel starting at the given row.
func (sp *sourcePanel) renderSource(w interface{ Write([]byte) (int, error) }, startRow int) {
	if sp == nil || len(sp.lines) == 0 {
		return
	}

	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	divider := strings.Repeat("─", min(termW, 80))
	render.WriteAt(w, startRow, 0, divider, render.Style{Dim: true})

	label := " Source: " + sp.path
	render.WriteAt(w, startRow+1, 0, label, render.Style{Dim: true, Bold: true})

	visible := sp.visibleLines(startRow + 2)
	for i := 0; i < visible; i++ {
		lineIdx := sp.scrollOff + i
		if lineIdx >= len(sp.lines) {
			break
		}
		line := sp.lines[lineIdx]
		if len(line) > termW-3 {
			line = line[:termW-3]
		}
		render.WriteAt(w, startRow+2+i, 1, line, render.Style{Dim: true})
	}

	// Scroll indicator.
	if sp.scrollOff > 0 || sp.scrollOff+visible < len(sp.lines) {
		hint := " j/k scroll"
		render.WriteAt(w, startRow+2+visible, 0, hint, render.Style{Dim: true})
	}
}
