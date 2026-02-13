package sumitest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// AssertSnapshots runs a scenario and compares frames against saved snapshots
// in the "testdata" directory relative to the test file.
// Use -update flag to create or update snapshot files.
func AssertSnapshots(t testing.TB, s Scenario) {
	t.Helper()
	AssertSnapshotsDir(t, s, "testdata")
}

// AssertSnapshotsDir runs a scenario and compares frames against saved snapshots
// in the given directory.
func AssertSnapshotsDir(t testing.TB, s Scenario, dir string) {
	t.Helper()

	frames := RunScenario(s)
	path := filepath.Join(dir, s.Name+".snapshot")

	if updateSnapshots {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create snapshot dir: %v", err)
			return
		}
		if err := WriteSnapshot(path, frames); err != nil {
			t.Fatalf("failed to write snapshot: %v", err)
			return
		}
		return
	}

	saved, err := ReadSnapshot(path)
	if err != nil {
		t.Fatalf("snapshot file not found: %s\nRun with -update to create it.", path)
		return
	}

	if len(frames) != len(saved) {
		t.Fatalf("frame count mismatch: got %d frames, snapshot has %d", len(frames), len(saved))
		return
	}

	for i, got := range frames {
		want := saved[i]
		if got.StyledText != want.StyledText {
			t.Errorf("frame %q (#%d) mismatch:\n%s",
				got.Name, i, formatDiff(want.StyledText, got.StyledText))
		}
	}
}

// formatDiff produces a simple side-by-side diff for snapshot mismatches.
func formatDiff(want, got string) string {
	return fmt.Sprintf("--- want ---\n%s\n--- got ---\n%s", want, got)
}
