package preview

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
)

func TestPvMatchesNoSnapshot(t *testing.T) {
	// Given
	pvSnapshots = nil
	pvActualStyled = "something"

	// When
	result := pvMatches(0)

	// Then
	if result != 0 {
		t.Errorf("pvMatches = %d, want 0 (no snapshot)", result)
	}
}

func TestPvMatchesMatch(t *testing.T) {
	// Given
	pvSnapshots = []sumitest.Frame{
		{Name: "initial", StyledText: "hello"},
	}
	pvActualStyled = "hello"

	// When
	result := pvMatches(0)

	// Then
	if result != 1 {
		t.Errorf("pvMatches = %d, want 1 (match)", result)
	}
}

func TestPvMatchesDiff(t *testing.T) {
	// Given
	pvSnapshots = []sumitest.Frame{
		{Name: "initial", StyledText: "expected"},
	}
	pvActualStyled = "actual"

	// When
	result := pvMatches(0)

	// Then
	if result != 2 {
		t.Errorf("pvMatches = %d, want 2 (diff)", result)
	}
}

func TestPvStepCount(t *testing.T) {
	// Given
	pvInfo = &sumitest.InfoResponse{
		Steps: []string{"initial", "typed", "clicked"},
	}

	// When/Then
	if pvStepCount() != 3 {
		t.Errorf("pvStepCount = %d, want 3", pvStepCount())
	}
}

func TestPvStepName(t *testing.T) {
	// Given
	pvInfo = &sumitest.InfoResponse{
		Steps: []string{"initial", "typed"},
	}

	// When/Then
	if pvStepName(0) != "initial" {
		t.Errorf("pvStepName(0) = %q, want %q", pvStepName(0), "initial")
	}
	if pvStepName(1) != "typed" {
		t.Errorf("pvStepName(1) = %q, want %q", pvStepName(1), "typed")
	}
	if pvStepName(5) != "" {
		t.Errorf("pvStepName(5) = %q, want empty", pvStepName(5))
	}
}

func TestPvScenarioName(t *testing.T) {
	// Given
	pvInfo = &sumitest.InfoResponse{Name: "textinput-basics"}

	// When/Then
	if pvScenarioName() != "textinput-basics" {
		t.Errorf("pvScenarioName = %q, want %q", pvScenarioName(), "textinput-basics")
	}
}

func TestPvComponentHeight(t *testing.T) {
	// Given
	pvInfo = &sumitest.InfoResponse{Height: 5}

	// When/Then — height + 2 for borders
	if pvComponentHeight() != 7 {
		t.Errorf("pvComponentHeight = %d, want 7", pvComponentHeight())
	}
}

func TestPvUpdateSnapshot(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	testdataDir := filepath.Join(tmpDir, "testdata")
	os.MkdirAll(testdataDir, 0o755)

	pvSnapPath = filepath.Join(testdataDir, "test.snapshot")
	pvSnapshots = nil
	pvActualStyled = "<<bold>>hello<</>>"
	pvInfo = &sumitest.InfoResponse{
		Steps: []string{"initial"},
	}

	// When
	err := pvUpdateSnapshot(0)

	// Then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pvSnapshots) != 1 {
		t.Fatalf("snapshots len = %d, want 1", len(pvSnapshots))
	}
	if pvSnapshots[0].StyledText != "<<bold>>hello<</>>" {
		t.Errorf("snapshot text = %q, want %q", pvSnapshots[0].StyledText, "<<bold>>hello<</>>")
	}

	// Verify file was written
	frames, err := sumitest.ReadSnapshot(pvSnapPath)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if len(frames) != 1 || frames[0].StyledText != "<<bold>>hello<</>>" {
		t.Errorf("read-back mismatch: %v", frames)
	}
}
