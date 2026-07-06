package dev

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// G2a: the dev rebuild pipeline — generate → build with structured
// results, and a directory watcher that coalesces changes.

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func tinyModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/tiny\n\ngo 1.25\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	return dir
}

func TestBuildProducesRunnableBinary(t *testing.T) {
	// Given
	dir := tinyModule(t)
	generated := false

	// When
	res := Build(dir, filepath.Join(t.TempDir(), "bin"), func(string) error {
		generated = true
		return nil
	})

	// Then
	if !generated {
		t.Error("generate step not invoked")
	}
	if res.Err != "" {
		t.Fatalf("build failed: %s", res.Err)
	}
	if err := exec.Command(res.Binary).Run(); err != nil {
		t.Errorf("binary does not run: %v", err)
	}
	if res.Duration <= 0 {
		t.Error("duration not recorded")
	}
}

func TestBuildReportsGenerateErrors(t *testing.T) {
	// Given / When
	res := Build(tinyModule(t), "", func(string) error {
		return errors.New("app.sumi:3: unexpected token")
	})

	// Then
	if res.Binary != "" {
		t.Error("failed build must not report a binary")
	}
	if res.Err == "" || !contains(res.Err, "app.sumi:3") {
		t.Errorf("err = %q, want the generate error", res.Err)
	}
}

func TestBuildReportsCompileErrors(t *testing.T) {
	// Given
	dir := tinyModule(t)
	writeFile(t, dir, "main.go", "package main\n\nfunc main() { undefined() }\n")

	// When
	res := Build(dir, filepath.Join(t.TempDir(), "bin"), func(string) error { return nil })

	// Then
	if res.Err == "" || !contains(res.Err, "undefined") {
		t.Errorf("err = %q, want the compile error", res.Err)
	}
}

func TestWatchTreeFiresOnMatchingChange(t *testing.T) {
	// Given
	dir := t.TempDir()
	writeFile(t, dir, "app.sumi", "v1")
	changes := make(chan struct{}, 8)
	w := WatchTree([]string{dir}, []string{".sumi", ".go"}, 10*time.Millisecond, func() {
		changes <- struct{}{}
	})
	defer w.Stop()

	// When: modify after the initial scan settles.
	time.Sleep(30 * time.Millisecond)
	writeFile(t, dir, "app.sumi", "v2 with different length")

	// Then
	select {
	case <-changes:
	case <-time.After(2 * time.Second):
		t.Fatal("change not detected")
	}
}

func TestWatchTreeIgnoresOtherExtensions(t *testing.T) {
	// Given
	dir := t.TempDir()
	changes := make(chan struct{}, 8)
	w := WatchTree([]string{dir}, []string{".sumi"}, 10*time.Millisecond, func() {
		changes <- struct{}{}
	})
	defer w.Stop()

	// When
	time.Sleep(30 * time.Millisecond)
	writeFile(t, dir, "notes.txt", "irrelevant")

	// Then
	select {
	case <-changes:
		t.Fatal("txt change should not fire")
	case <-time.After(100 * time.Millisecond):
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
