package preview

import (
	"os"
	"path/filepath"
	"time"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/term"
)

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

