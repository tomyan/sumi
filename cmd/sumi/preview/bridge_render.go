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
	leftBoxW := (termW + 1) / 2 // first flex child gets remainder

	// Render VT100 buffer into left panel (actual).
	pvScreen.Buffer().RenderToOffset(os.Stdout, 1, 2)

	// Render styled text into right panel (expected).
	pvRenderExpected(os.Stdout, 1, leftBoxW+2)

	// Highlight left panel border when in interactive mode.
	if pvInteractive {
		pvHighlightBox(os.Stdout, 0, 0, leftBoxW, pvComponentHeight(), "Actual",
			render.Style{FG: render.Color{IsRGB: true, R: 0x50, G: 0xff, B: 0x50}, Bold: true})
	}

	// Render editor screens into placeholder areas.
	pvInjectEditors(os.Stdout, termW)

	// Highlight focused editor border.
	pvHighlightFocusedEditor(os.Stdout, termW)
}

// pvHighlightBox redraws a box border in the given style to indicate focus.
func pvHighlightBox(w *os.File, row, col, boxW, boxH int, title string, s render.Style) {
	// Top edge.
	top := make([]rune, boxW)
	top[0] = '┌'
	for i := 1; i < boxW-1; i++ {
		top[i] = '─'
	}
	top[boxW-1] = '┐'
	render.WriteAt(w, row, col, string(top), s)

	// Redraw title over top edge.
	if title != "" {
		render.WriteAt(w, row, col+2, " "+title+" ", s)
	}

	// Bottom edge.
	bottom := make([]rune, boxW)
	bottom[0] = '└'
	for i := 1; i < boxW-1; i++ {
		bottom[i] = '─'
	}
	bottom[boxW-1] = '┘'
	render.WriteAt(w, row+boxH-1, col, string(bottom), s)

	// Side edges.
	for r := row + 1; r < row+boxH-1; r++ {
		render.WriteAt(w, r, col, "│", s)
		render.WriteAt(w, r, col+boxW-1, "│", s)
	}
}

// pvHighlightFocusedEditor redraws all editor borders in their base colour,
// then highlights the focused one in a brighter variant with a bold title.
func pvHighlightFocusedEditor(w *os.File, termW int) {
	startRow := pvComponentHeight() + 1
	leftW := pvLeftEditorWidth()
	rightW := pvRightEditorWidth()
	topH := pvEditorHeight()
	botH := pvScenarioHeight()

	base := render.Style{FG: render.Color{Name: "magenta"}}
	bright := render.Style{FG: render.Color{IsRGB: true, R: 0xff, G: 0x60, B: 0xff}, Bold: true}

	// Redraw all editor borders in the base colour to clear any previous highlight.
	pvHighlightBox(w, startRow, 0, leftW, topH, pvSourceTitle(), base)
	pvHighlightBox(w, startRow, leftW, rightW, topH, pvSnapshotTitle(), base)
	pvHighlightBox(w, startRow+topH, 0, termW, botH, pvScenarioTitle(), base)

	// Highlight the focused editor.
	switch pvFocus {
	case FocusEditor1:
		pvHighlightBox(w, startRow, 0, leftW, topH, pvSourceTitle(), bright)
	case FocusEditor2:
		pvHighlightBox(w, startRow, leftW, rightW, topH, pvSnapshotTitle(), bright)
	case FocusEditor3:
		pvHighlightBox(w, startRow+topH, 0, termW, botH, pvScenarioTitle(), bright)
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
	leftBoxW := pvLeftEditorWidth()

	// Editor 1 (source) — top-left, offset by border.
	if pvEditors[0] != nil {
		pvEditors[0].mu.Lock()
		pvEditors[0].screen.Buffer().RenderToOffset(w, startRow+1, 1)
		pvEditors[0].mu.Unlock()
	}

	// Editor 2 (snapshot) — top-right, offset by border.
	if pvEditors[1] != nil {
		pvEditors[1].mu.Lock()
		pvEditors[1].screen.Buffer().RenderToOffset(w, startRow+1, leftBoxW+1)
		pvEditors[1].mu.Unlock()
	}

	// Editor 3 (scenario) — below the two side-by-side editors, offset by border.
	topEdH := pvEditorHeight()
	scenRow := startRow + topEdH
	if pvEditors[2] != nil {
		pvEditors[2].mu.Lock()
		pvEditors[2].screen.Buffer().RenderToOffset(w, scenRow+1, 1)
		pvEditors[2].mu.Unlock()
	}
}
