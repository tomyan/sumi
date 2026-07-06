// Package dev implements the sumi dev rebuild pipeline: watch source,
// regenerate, rebuild, and hand a fresh binary to the supervisor.
package dev

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DevSocketPath returns the inspect socket path for an app directory:
// short and stable (sun_path caps Unix socket paths at ~104 bytes, so
// the app dir cannot be used directly).
func DevSocketPath(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(os.TempDir(), fmt.Sprintf("sumi-dev-%x.sock", sum[:6]))
}

// Result reports one generate+build attempt. Binary is empty when the
// attempt failed; Err then carries the human-readable tool output.
type Result struct {
	Binary   string
	Err      string
	Duration time.Duration
}

// Build regenerates the app (via the injected generate step) and
// compiles it to out. Errors from either stage come back structured —
// never as a dead terminal.
func Build(dir, out string, generate func(dir string) error) Result {
	start := time.Now()
	if err := generate(dir); err != nil {
		return Result{Err: err.Error(), Duration: time.Since(start)}
	}
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return Result{Err: msg, Duration: time.Since(start)}
	}
	return Result{Binary: out, Duration: time.Since(start)}
}

// Watcher polls directory trees for changes to matching files.
type Watcher struct {
	stop chan struct{}
	done chan struct{}
}

// WatchTree polls roots at the given interval and calls onChange
// (coalesced: once per tick at most) when any file with one of the
// extensions is added, modified, or removed. Generated *_sumi.go files
// are ignored — regeneration writes them, and watching them would loop.
func WatchTree(roots, exts []string, interval time.Duration, onChange func()) *Watcher {
	w := &Watcher{stop: make(chan struct{}), done: make(chan struct{})}
	prev := scanTree(roots, exts)
	go func() {
		defer close(w.done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-w.stop:
				return
			case <-ticker.C:
				next := scanTree(roots, exts)
				if changed(prev, next) {
					prev = next
					onChange()
				}
			}
		}
	}()
	return w
}

// Stop terminates the watcher goroutine.
func (w *Watcher) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	<-w.done
}

// scanTree collects mtimes of matching files under the roots.
func scanTree(roots, exts []string) map[string]time.Time {
	seen := make(map[string]time.Time)
	for _, root := range roots {
		filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_sumi.go") {
				return nil
			}
			for _, ext := range exts {
				if strings.HasSuffix(path, ext) {
					if info, err := d.Info(); err == nil {
						seen[path] = info.ModTime()
					}
					break
				}
			}
			return nil
		})
	}
	return seen
}

func changed(prev, next map[string]time.Time) bool {
	if len(prev) != len(next) {
		return true
	}
	for path, mt := range next {
		if p, ok := prev[path]; !ok || !mt.Equal(p) {
			return true
		}
	}
	return false
}
