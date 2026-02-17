package preview

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcherDetectsFileChange(t *testing.T) {
	// Given — a file being watched
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("original"), 0644)

	var callCount int32
	w := NewWatcher([]string{path}, 50*time.Millisecond, func(changed string) {
		atomic.AddInt32(&callCount, 1)
	})
	defer w.Stop()

	// When — modify the file after a brief delay
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(path, []byte("modified"), 0644)

	// Then — callback fires within a reasonable time
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&callCount) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Error("watcher callback was not called after file modification")
}

func TestWatcherIgnoresUnchangedFiles(t *testing.T) {
	// Given — a file being watched
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("content"), 0644)

	var callCount int32
	w := NewWatcher([]string{path}, 50*time.Millisecond, func(changed string) {
		atomic.AddInt32(&callCount, 1)
	})
	defer w.Stop()

	// When — wait without modifying
	time.Sleep(200 * time.Millisecond)

	// Then — no callback
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("callback called %d times, want 0", atomic.LoadInt32(&callCount))
	}
}

func TestWatcherStops(t *testing.T) {
	// Given
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("content"), 0644)

	w := NewWatcher([]string{path}, 50*time.Millisecond, func(changed string) {})

	// When
	w.Stop()

	// Then — no hang, test completes
}

func TestWatcherPassesChangedPath(t *testing.T) {
	// Given — two files being watched
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.txt")
	path2 := filepath.Join(dir, "b.txt")
	os.WriteFile(path1, []byte("a"), 0644)
	os.WriteFile(path2, []byte("b"), 0644)

	var changedPath string
	w := NewWatcher([]string{path1, path2}, 50*time.Millisecond, func(changed string) {
		changedPath = changed
	})
	defer w.Stop()

	// When — modify only the second file
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(path2, []byte("b-modified"), 0644)

	// Then
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if changedPath != "" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if changedPath != path2 {
		t.Errorf("changedPath = %q, want %q", changedPath, path2)
	}
}
