package preview

import (
	"os"
	"path/filepath"
	"time"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/term"
)

// pvEditorRowRemaining returns the total height available for the two editor rows.
func pvEditorRowRemaining() int {
	_, termH := term.GetSize(int(os.Stdout.Fd()))
	remaining := termH - pvComponentHeight() - 1
	if remaining < 6 {
		return 6
	}
	return remaining
}

// pvEditorHeight returns the height for the first (top) editor row (including border).
// The first flex child gets the remainder from integer division.
func pvEditorHeight() int {
	remaining := pvEditorRowRemaining()
	return (remaining + 1) / 2
}

// pvScenarioHeight returns the height for the second (bottom) editor row (including border).
func pvScenarioHeight() int {
	remaining := pvEditorRowRemaining()
	return remaining / 2
}

// pvEditorContentHeight returns the inner height for editor PTY (minus border).
func pvEditorContentHeight() int {
	h := pvEditorHeight() - 2
	if h < 1 {
		return 1
	}
	return h
}

// pvScenarioContentHeight returns the inner height for scenario editor PTY (minus border).
func pvScenarioContentHeight() int {
	h := pvScenarioHeight() - 2
	if h < 1 {
		return 1
	}
	return h
}

// pvLeftEditorWidth returns the width of the left side-by-side editor (including border).
// The first flex child gets the remainder from integer division.
func pvLeftEditorWidth() int {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	return (termW + 1) / 2
}

// pvRightEditorWidth returns the width of the right side-by-side editor (including border).
func pvRightEditorWidth() int {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	return termW / 2
}

// pvLeftEditorContentWidth returns the inner width for the left editor PTY (minus border).
func pvLeftEditorContentWidth() int {
	w := pvLeftEditorWidth() - 2
	if w < 1 {
		return 1
	}
	return w
}

// pvRightEditorContentWidth returns the inner width for the right editor PTY (minus border).
func pvRightEditorContentWidth() int {
	w := pvRightEditorWidth() - 2
	if w < 1 {
		return 1
	}
	return w
}

// pvScenarioEditorWidth returns the width of the full-width scenario editor (including border).
func pvScenarioEditorWidth() int {
	termW, _ := term.GetSize(int(os.Stdout.Fd()))
	return termW
}

// pvScenarioContentWidth returns the inner width for scenario editor PTY (minus border).
func pvScenarioContentWidth() int {
	w := pvScenarioEditorWidth() - 2
	if w < 1 {
		return 1
	}
	return w
}

// SetupEditors starts nvim instances for source, snapshot, and scenario files.
func SetupEditors() {
	edH := pvEditorContentHeight()
	leftW := pvLeftEditorContentWidth()
	rightW := pvRightEditorContentWidth()
	scenH := pvScenarioContentHeight()
	scenW := pvScenarioContentWidth()

	wake := func() {
		if pvApp != nil {
			pvApp.Wake()
		}
	}

	// Editor 1: source file (left).
	if pvInfo.SourceFile != "" {
		path := filepath.Join(pvCompDir, pvInfo.SourceFile)
		ed, err := NewEditor(path, edH, leftW, wake)
		if err == nil {
			pvEditors[0] = ed
		}
	}

	// Editor 2: snapshot file (right).
	snapPath := filepath.Join(pvCompDir, "testdata", pvInfo.Name+".snapshot")
	ed, err := NewEditor(snapPath, edH, rightW, wake)
	if err == nil {
		pvEditors[1] = ed
	}

	// Editor 3: scenario file (full-width bottom).
	if pvInfo.ScenarioFile != "" {
		path := filepath.Join(pvCompDir, pvInfo.ScenarioFile)
		ed, err := NewEditor(path, scenH, scenW, wake)
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
	edH := pvEditorContentHeight()
	leftW := pvLeftEditorContentWidth()
	rightW := pvRightEditorContentWidth()
	scenH := pvScenarioContentHeight()
	scenW := pvScenarioContentWidth()

	if pvEditors[0] != nil {
		pvEditors[0].Resize(edH, leftW)
	}
	if pvEditors[1] != nil {
		pvEditors[1].Resize(edH, rightW)
	}
	if pvEditors[2] != nil {
		pvEditors[2].Resize(scenH, scenW)
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

// SetOnReload sets a callback for when source/scenario files are modified.
// The callback should regenerate code, restart the subprocess, and re-step.
func SetOnReload(fn func()) {
	pvOnReload = fn
}

// pvStartWatcher starts the file watcher for source, snapshot, and scenario files.
func pvStartWatcher() {
	var paths []string
	if pvInfo.SourceFile != "" {
		paths = append(paths, filepath.Join(pvCompDir, pvInfo.SourceFile))
	}
	paths = append(paths, pvSnapPath)
	if pvInfo.ScenarioFile != "" {
		paths = append(paths, filepath.Join(pvCompDir, pvInfo.ScenarioFile))
	}

	if len(paths) == 0 {
		return
	}

	pvWatcher = NewWatcher(paths, 500*time.Millisecond, pvHandleFileChange)
}

// pvStopWatcher stops the file watcher.
func pvStopWatcher() {
	if pvWatcher != nil {
		pvWatcher.Stop()
		pvWatcher = nil
	}
}

// pvHandleFileChange is called when a watched file's mtime changes.
func pvHandleFileChange(path string) {
	// Snapshot file changed — re-read snapshots and update match status.
	if path == pvSnapPath {
		frames, err := sumitest.ReadSnapshot(pvSnapPath)
		if err == nil {
			pvSnapshots = frames
		}
		if pvApp != nil {
			pvApp.Dirty = true
			pvApp.Wake()
		}
		return
	}

	// Source or scenario file changed — trigger reload callback.
	if pvOnReload != nil {
		pvOnReload()
	}
	if pvApp != nil {
		pvApp.Dirty = true
		pvApp.Wake()
	}
}

