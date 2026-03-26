package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/sumitest"
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
	pvWatcher      *Watcher
	pvOnReload     func() // callback when source/scenario files change (set by caller)
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

	comp := NewPreview(PreviewProps{})

	SetupEditors()
	defer CleanupEditors()

	pvStartWatcher()
	defer pvStopWatcher()

	// Run with post-render injection for editors and content.
	tui.RunWithOptions(comp, tui.RunOptions{
		OnPostRender: pvInjectContent,
		OnResize:     pvResizeEditors,
		SetApp:       func(a *tui.App) { pvApp = a },
	})
}

// isTimeout checks if an error is a read timeout.
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline")
}
