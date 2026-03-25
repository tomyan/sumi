package preview

import (
	"os"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/term"
)

// pvInjectContent writes rich content into the panel areas after sumi renders.
func pvInjectContent() {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	leftBoxW := termW / 2

	// Render VT100 buffer into left panel (actual).
	pvScreen.Buffer().RenderToOffset(os.Stdout, 1, 2)

	// Render styled text into right panel (expected).
	pvRenderExpected(os.Stdout, 1, leftBoxW+2)

	// Highlight left panel border when in interactive mode.
	if pvInteractive {
		pvHighlightBorder(os.Stdout, leftBoxW)
	}

	// Render editor screens into placeholder areas.
	pvInjectEditors(os.Stdout, termW)
}

// pvHighlightBorder redraws the left panel border in green to indicate interactive mode.
func pvHighlightBorder(w *os.File, boxW int) {
	h := pvComponentHeight()
	style := render.Style{FG: render.Color{Name: "green"}, Bold: true}

	// Top edge.
	top := make([]rune, boxW)
	top[0] = '┌'
	for i := 1; i < boxW-1; i++ {
		top[i] = '─'
	}
	top[boxW-1] = '┐'
	render.WriteAt(w, 0, 0, string(top), style)

	// Redraw title over top edge.
	render.WriteAt(w, 0, 2, " Actual ", style)

	// Bottom edge.
	bottom := make([]rune, boxW)
	bottom[0] = '└'
	for i := 1; i < boxW-1; i++ {
		bottom[i] = '─'
	}
	bottom[boxW-1] = '┘'
	render.WriteAt(w, h-1, 0, string(bottom), style)

	// Side edges.
	for r := 1; r < h-1; r++ {
		render.WriteAt(w, r, 0, "│", style)
		render.WriteAt(w, r, boxW-1, "│", style)
	}
}

// pvRenderExpected writes styled text lines into the expected panel area.
func pvRenderExpected(w *os.File, startRow, startCol int) {
	expected := expectedText()
	if expected == "" {
		render.WriteAt(w, startRow, startCol, "(no snapshot)", render.Style{Dim: true})
		return
	}
	lines := strings.Split(expected, "\n")
	for i, line := range lines {
		if i >= pvInfo.Height {
			break
		}
		segments := parseStyledLine(line)
		writeStyledLine(w, startRow+i, startCol, segments)
	}
}

// expectedText returns the snapshot text for the current step.
func expectedText() string {
	if pvCurrent >= len(pvSnapshots) {
		return ""
	}
	return pvSnapshots[pvCurrent].StyledText
}

// pvInjectEditors renders editor screens into placeholder areas.
// Each editor box has a single-line border, so content starts 1 row down and 1 col right.
func pvInjectEditors(w *os.File, termW int) {
	startRow := pvComponentHeight() + 1 // +1 for status bar row
	leftW := termW / 2

	// Editor 1 (source) — top-left, offset by border.
	if pvEditors[0] != nil {
		pvEditors[0].mu.Lock()
		pvEditors[0].screen.Buffer().RenderToOffset(w, startRow+1, 1)
		pvEditors[0].mu.Unlock()
	}

	// Editor 2 (snapshot) — top-right, offset by border.
	if pvEditors[1] != nil {
		pvEditors[1].mu.Lock()
		pvEditors[1].screen.Buffer().RenderToOffset(w, startRow+1, leftW+1)
		pvEditors[1].mu.Unlock()
	}

	// Editor 3 (scenario) — below the two side-by-side editors, offset by border.
	edH := pvEditorHeight()
	scenRow := startRow + edH
	if pvEditors[2] != nil {
		pvEditors[2].mu.Lock()
		pvEditors[2].screen.Buffer().RenderToOffset(w, scenRow+1, 1)
		pvEditors[2].mu.Unlock()
	}
}
