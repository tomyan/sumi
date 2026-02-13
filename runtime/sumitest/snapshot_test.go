package sumitest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteSnapshotCreatesFile(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snapshot")
	frames := []Frame{
		{Name: "initial", StyledText: "Count: 0"},
	}

	// When
	err := WriteSnapshot(path, frames)

	// Then
	if err != nil {
		t.Fatalf("WriteSnapshot failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}
}

func TestSnapshotRoundtripSingleFrame(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snapshot")
	frames := []Frame{
		{Name: "initial", StyledText: "Count: 0"},
	}

	// When
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	got, err := ReadSnapshot(path)

	// Then
	if err != nil {
		t.Fatalf("ReadSnapshot: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(got))
	}
	if got[0].Name != "initial" {
		t.Errorf("name: expected %q, got %q", "initial", got[0].Name)
	}
	if got[0].StyledText != "Count: 0" {
		t.Errorf("styled text: expected %q, got %q", "Count: 0", got[0].StyledText)
	}
}

func TestSnapshotRoundtripMultipleFrames(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snapshot")
	frames := []Frame{
		{Name: "initial", StyledText: "Count: 0"},
		{Name: "after-key", StyledText: "Count: 1"},
		{Name: "final", StyledText: "Count: 2"},
	}

	// When
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	got, err := ReadSnapshot(path)

	// Then
	if err != nil {
		t.Fatalf("ReadSnapshot: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(got))
	}
	for i, f := range frames {
		if got[i].Name != f.Name {
			t.Errorf("frame %d name: expected %q, got %q", i, f.Name, got[i].Name)
		}
		if got[i].StyledText != f.StyledText {
			t.Errorf("frame %d text: expected %q, got %q", i, f.StyledText, got[i].StyledText)
		}
	}
}

func TestSnapshotRoundtripMultilineText(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snapshot")
	frames := []Frame{
		{Name: "multiline", StyledText: "Line 1\nLine 2\nLine 3"},
	}

	// When
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	got, err := ReadSnapshot(path)

	// Then
	if err != nil {
		t.Fatalf("ReadSnapshot: %v", err)
	}
	if got[0].StyledText != "Line 1\nLine 2\nLine 3" {
		t.Errorf("multiline text mismatch:\nexpected: %q\ngot:      %q", frames[0].StyledText, got[0].StyledText)
	}
}

func TestSnapshotRoundtripStyledMarkup(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snapshot")
	frames := []Frame{
		{Name: "styled", StyledText: "<<green>>hello<</>> world"},
	}

	// When
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	got, err := ReadSnapshot(path)

	// Then
	if err != nil {
		t.Fatalf("ReadSnapshot: %v", err)
	}
	if got[0].StyledText != "<<green>>hello<</>> world" {
		t.Errorf("styled text mismatch:\nexpected: %q\ngot:      %q", frames[0].StyledText, got[0].StyledText)
	}
}

func TestReadSnapshotFileNotFound(t *testing.T) {
	// When
	_, err := ReadSnapshot("/nonexistent/path.snapshot")

	// Then
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSnapshotFileFormat(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.snapshot")
	frames := []Frame{
		{Name: "first", StyledText: "Hello"},
		{Name: "second", StyledText: "World"},
	}

	// When
	if err := WriteSnapshot(path, frames); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	data, err := os.ReadFile(path)

	// Then
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	expected := "=== Frame: first ===\nHello\n\n=== Frame: second ===\nWorld\n\n"
	if string(data) != expected {
		t.Errorf("file format mismatch:\nexpected:\n%s\ngot:\n%s", expected, string(data))
	}
}
