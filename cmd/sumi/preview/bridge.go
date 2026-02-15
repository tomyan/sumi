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
)

// Setup stores subprocess handles and loads snapshots and source.
func Setup(client *sumitest.Client, master *os.File, info *sumitest.InfoResponse, compDir string) {
	pvClient = client
	pvMaster = master
	pvInfo = info
	pvScreen = vt100.NewScreen(info.Width, info.Height)

	// Load snapshots.
	pvSnapPath = filepath.Join(compDir, "testdata", info.Name+".snapshot")
	frames, err := sumitest.ReadSnapshot(pvSnapPath)
	if err == nil {
		pvSnapshots = frames
	}

	// Load source file.
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

	return readUntilSentinel()
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

// pvComponentHeight returns the panel height (component height + 2 for borders).
func pvComponentHeight() int {
	return pvInfo.Height + 2
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

	app := CreateApp(0, 0)
	app.TestBuffer = nil
	app.OnPostRender = pvInjectContent
	app.Run()
}

// pvInjectContent writes rich content into the panel areas after sumi renders.
func pvInjectContent() {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	leftPanelW := (termW - 3) / 2 // 3 border chars

	// Render VT100 buffer into left panel (actual).
	pvScreen.Buffer().RenderToOffset(os.Stdout, 1, 2)

	// Render styled text into right panel (expected).
	pvRenderExpected(os.Stdout, 1, leftPanelW+3)
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
