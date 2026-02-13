package sumitest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssertSnapshotsMissingFileFails(t *testing.T) {
	// Given
	s := counterScenario()
	dir := t.TempDir()
	ft := &fakeTB{}

	// When
	AssertSnapshotsDir(ft, s, dir)

	// Then
	if !ft.failed {
		t.Fatal("expected test to fail when snapshot file is missing")
	}
	if !strings.Contains(ft.lastMsg, "snapshot file not found") {
		t.Errorf("expected 'snapshot file not found' message, got: %s", ft.lastMsg)
	}
}

func TestAssertSnapshotsUpdateCreatesFile(t *testing.T) {
	// Given
	s := counterScenario()
	dir := t.TempDir()
	oldVal := updateSnapshots
	updateSnapshots = true
	defer func() { updateSnapshots = oldVal }()
	ft := &fakeTB{}

	// When
	AssertSnapshotsDir(ft, s, dir)

	// Then
	if ft.failed {
		t.Fatalf("expected test to pass in update mode, got: %s", ft.lastMsg)
	}
	path := filepath.Join(dir, "counter-basics.snapshot")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}
}

func TestAssertSnapshotsMatchingPassses(t *testing.T) {
	// Given
	s := counterScenario()
	dir := t.TempDir()
	frames := RunScenario(s)
	path := filepath.Join(dir, "counter-basics.snapshot")
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	ft := &fakeTB{}

	// When
	AssertSnapshotsDir(ft, s, dir)

	// Then
	if ft.failed {
		t.Fatalf("expected test to pass when snapshots match, got: %s", ft.lastMsg)
	}
}

func TestAssertSnapshotsMismatchFails(t *testing.T) {
	// Given
	s := counterScenario()
	dir := t.TempDir()
	path := filepath.Join(dir, "counter-basics.snapshot")
	wrongFrames := []Frame{
		{Name: "initial", StyledText: "WRONG"},
		{Name: "after-first-key", StyledText: "WRONG"},
		{Name: "after-second-key", StyledText: "WRONG"},
	}
	if err := WriteSnapshot(path, wrongFrames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	ft := &fakeTB{}

	// When
	AssertSnapshotsDir(ft, s, dir)

	// Then
	if !ft.failed {
		t.Fatal("expected test to fail when snapshots mismatch")
	}
}

func TestAssertSnapshotsFrameCountMismatchFails(t *testing.T) {
	// Given
	s := counterScenario()
	dir := t.TempDir()
	path := filepath.Join(dir, "counter-basics.snapshot")
	frames := []Frame{
		{Name: "initial", StyledText: "Count: 0"},
	}
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	ft := &fakeTB{}

	// When
	AssertSnapshotsDir(ft, s, dir)

	// Then
	if !ft.failed {
		t.Fatal("expected test to fail when frame count mismatches")
	}
}

// fakeTB implements testing.TB enough for our assertion tests.
type fakeTB struct {
	testing.TB
	failed  bool
	lastMsg string
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatalf(format string, args ...interface{}) {
	f.failed = true
	f.lastMsg = format
	if len(args) > 0 {
		f.lastMsg = strings.TrimRight(strings.Repeat("%v ", len(args)), " ")
		f.lastMsg = format
	}
}

func (f *fakeTB) Errorf(format string, args ...interface{}) {
	f.failed = true
	f.lastMsg = format
}
