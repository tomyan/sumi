package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
	"github.com/tomyan/sumi/runtime/vt100"
)

// Package-level state shared between the .sumi component and the bridge.
var (
	pvClient       *sumitest.Client
	pvMaster       *os.File
	pvInfo         *sumitest.InfoResponse
	pvScreen       *vt100.Screen
	pvSnapshots    []sumitest.Frame
	pvSnapPath     string
	pvActualStyled string
	pvSourceLines  []string
	pvCurrent      int  // current step index, updated by pvStepTo
	pvInteractive  bool // true when in interactive mode
	pvCompDir      string
	pvEditors      [3]*Editor // source, snapshot, scenario
	pvApp          *tui.App   // reference for Wake()
)

// Setup stores subprocess handles and loads snapshots and source.
func Setup(client *sumitest.Client, master *os.File, info *sumitest.InfoResponse, compDir string) {
	pvClient = client
	pvMaster = master
	pvInfo = info
	pvScreen = vt100.NewScreen(info.Width, info.Height)
	pvCompDir = compDir

	// Load snapshots.
	pvSnapPath = filepath.Join(compDir, "testdata", info.Name+".snapshot")
	frames, err := sumitest.ReadSnapshot(pvSnapPath)
	if err == nil {
		pvSnapshots = frames
	}

	// Load source file (for fallback display).
	if info.SourceFile != "" {
		absPath := filepath.Join(compDir, info.SourceFile)
		if data, err := os.ReadFile(absPath); err == nil {
			pvSourceLines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		}
	}
}

// pvStepTo steps the subprocess to the given index and reads PTY output.
func pvStepTo(index int) error {
	pvScreen.ResetSentinel()

	resp, err := pvClient.Step(index)
	if err != nil {
		return fmt.Errorf("step %d: %w", index, err)
	}
	pvActualStyled = resp.StyledText
	pvCurrent = index

	if err := readUntilSentinel(); err != nil {
		return fmt.Errorf("read pty: %w", err)
	}
	return nil
}

// pvSendInput sends an input event to the subprocess and reads the updated output.
func pvSendInput(evt input.Event) error {
	pvScreen.ResetSentinel()

	resp, err := pvClient.Input(evt)
	if err != nil {
		return fmt.Errorf("input: %w", err)
	}
	pvActualStyled = resp.StyledText

	if err := readUntilSentinel(); err != nil {
		return err
	}
	return nil
}

// pvEnterInteractive enters interactive mode.
func pvEnterInteractive() { pvInteractive = true }

// pvExitInteractive exits interactive mode.
func pvExitInteractive() { pvInteractive = false }

// pvIsInteractive returns whether interactive mode is active.
func pvIsInteractive() bool { return pvInteractive }

// readUntilSentinel reads PTY output until the VT100 sentinel is seen.
func readUntilSentinel() error {
	buf := make([]byte, 4096)
	deadline := time.Now().Add(5 * time.Second)

	for !pvScreen.SentinelSeen() {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for sentinel")
		}
		pvMaster.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		n, err := pvMaster.Read(buf)
		if n > 0 {
			pvScreen.Write(buf[:n])
		}
		if err != nil && !isTimeout(err) {
			return err
		}
	}
	return nil
}

// pvMatches returns the match status for a step: 0=no snapshot, 1=match, 2=diff.
func pvMatches(index int) int {
	if index >= len(pvSnapshots) || pvSnapshots[index].StyledText == "" {
		return 0
	}
	if pvActualStyled == pvSnapshots[index].StyledText {
		return 1
	}
	return 2
}

// pvUpdateSnapshot writes the current actual styled text as the snapshot for a step.
func pvUpdateSnapshot(index int) error {
	stepName := pvInfo.Steps[index]

	for len(pvSnapshots) <= index {
		pvSnapshots = append(pvSnapshots, sumitest.Frame{})
	}

	pvSnapshots[index] = sumitest.Frame{
		Name:       stepName,
		StyledText: pvActualStyled,
	}

	return sumitest.WriteSnapshot(pvSnapPath, pvSnapshots)
}

// pvStepCount returns the number of steps in the scenario.
func pvStepCount() int {
	return len(pvInfo.Steps)
}

// pvStepName returns the name of a step by index.
func pvStepName(index int) string {
	if index < 0 || index >= len(pvInfo.Steps) {
		return ""
	}
	return pvInfo.Steps[index]
}

// pvScenarioName returns the scenario name.
func pvScenarioName() string {
	return pvInfo.Name
}

// pvComponentHeight returns the panel height (component height + 2 for borders, min 10).
func pvComponentHeight() int {
	h := pvInfo.Height + 2
	if h < 10 {
		return 10
	}
	return h
}

// pvSourceTitle returns the title for the source panel.
func pvSourceTitle() string {
	if pvInfo.SourceFile != "" {
		return pvInfo.SourceFile
	}
	return "Source"
}

// RunPreview creates and runs the preview sumi app.
func RunPreview() {
	// Step to initial frame before creating the app.
	if err := pvStepTo(0); err != nil {
		fmt.Fprintf(os.Stderr, "initial step: %v\n", err)
		return
	}

	SetupEditors()
	defer CleanupEditors()

	app := CreateApp(0, 0)
	pvApp = app
	app.TestBuffer = nil
	app.OnPostRender = pvInjectContent
	app.Run()
}

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

// isTimeout checks if an error is a read timeout.
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline")
}

// pvEditorHeight returns the height for each editor row.
func pvEditorHeight() int {
	_, termH := term.GetSize(int(os.Stdout.Fd()))
	// Layout: component panels + status bar (1) + 2 editor rows
	used := pvComponentHeight() + 1
	remaining := termH - used
	if remaining < 6 {
		return 3
	}
	return remaining / 2
}

// pvEditorWidth returns the width of each side-by-side editor.
func pvEditorWidth() int {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	return termW / 2
}

// pvScenarioEditorWidth returns the width of the full-width scenario editor.
func pvScenarioEditorWidth() int {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	return termW
}

// SetupEditors starts nvim instances for source, snapshot, and scenario files.
func SetupEditors() {
	edH := pvEditorHeight()
	edW := pvEditorWidth()
	scenW := pvScenarioEditorWidth()

	wake := func() {
		if pvApp != nil {
			pvApp.Wake()
		}
	}

	// Editor 1: source file.
	if pvInfo.SourceFile != "" {
		path := filepath.Join(pvCompDir, pvInfo.SourceFile)
		ed, err := NewEditor(path, edH, edW, wake)
		if err == nil {
			pvEditors[0] = ed
		}
	}

	// Editor 2: snapshot file.
	snapPath := filepath.Join(pvCompDir, "testdata", pvInfo.Name+".snapshot")
	ed, err := NewEditor(snapPath, edH, edW, wake)
	if err == nil {
		pvEditors[1] = ed
	}

	// Editor 3: scenario file.
	if pvInfo.ScenarioFile != "" {
		path := filepath.Join(pvCompDir, pvInfo.ScenarioFile)
		ed, err := NewEditor(path, edH, scenW, wake)
		if err == nil {
			pvEditors[2] = ed
		}
	}
}

// CleanupEditors terminates all nvim processes.
func CleanupEditors() {
	for i, ed := range pvEditors {
		if ed != nil {
			ed.Close()
			pvEditors[i] = nil
		}
	}
}

// pvResizeEditors recalculates editor dimensions and resizes all PTYs.
func pvResizeEditors() {
	edH := pvEditorHeight()
	edW := pvEditorWidth()
	scenW := pvScenarioEditorWidth()

	if pvEditors[0] != nil {
		pvEditors[0].Resize(edH, edW)
	}
	if pvEditors[1] != nil {
		pvEditors[1].Resize(edH, edW)
	}
	if pvEditors[2] != nil {
		pvEditors[2].Resize(edH, scenW)
	}
}

// pvInjectEditors renders editor screens into placeholder areas.
func pvInjectEditors(w *os.File, termW int) {
	startRow := pvComponentHeight()
	leftW := termW / 2

	// Editor 1 (source) — top-left.
	if pvEditors[0] != nil {
		pvEditors[0].mu.Lock()
		pvEditors[0].screen.Buffer().RenderToOffset(w, startRow, 0)
		pvEditors[0].mu.Unlock()
	}

	// Editor 2 (snapshot) — top-right.
	if pvEditors[1] != nil {
		pvEditors[1].mu.Lock()
		pvEditors[1].screen.Buffer().RenderToOffset(w, startRow, leftW)
		pvEditors[1].mu.Unlock()
	}

	// Editor 3 (scenario) — below the two side-by-side editors.
	edH := pvEditorHeight()
	scenRow := startRow + edH
	if pvEditors[2] != nil {
		pvEditors[2].mu.Lock()
		pvEditors[2].screen.Buffer().RenderToOffset(w, scenRow, 0)
		pvEditors[2].mu.Unlock()
	}
}

// pvSnapshotTitle returns the title for the snapshot editor panel.
func pvSnapshotTitle() string {
	return pvInfo.Name + ".snapshot"
}

// pvScenarioTitle returns the title for the scenario editor panel.
func pvScenarioTitle() string {
	if pvInfo.ScenarioFile != "" {
		return pvInfo.ScenarioFile
	}
	return "scenario.go"
}

// pvFocusName returns the current focus state name for the status bar.
func pvFocusName() string {
	return pvFocus.Name()
}

// pvIsEditorFocused returns true if any editor is focused.
func pvIsEditorFocused() bool {
	return pvFocus >= FocusEditor1 && pvFocus <= FocusEditor3
}
