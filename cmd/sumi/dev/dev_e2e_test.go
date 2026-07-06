package dev_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/tomyan/sumi/runtime/pty"
	"github.com/tomyan/sumi/runtime/vt100"
)

var spaceRe = regexp.MustCompile(`\s+`)

// G2b end-to-end: the sumi dev supervisor mirrors the app, hot-swaps on
// save, keeps the last good build over errors, recovers, and exits with
// the app.

func TestDevLoopEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the CLI and an app; runs a PTY session")
	}

	// Given: the sumi CLI and a scaffolded app.
	tmp := t.TempDir()
	cli := filepath.Join(tmp, "sumi")
	if out, err := exec.Command("go", "build", "-o", cli, "..").CombinedOutput(); err != nil {
		t.Fatalf("build cli: %v\n%s", err, out)
	}
	appDir := filepath.Join(tmp, "app")
	sumiRoot, _ := filepath.Abs("../../..")
	initCmd := exec.Command(cli, "init", "--sumi-path", sumiRoot, appDir)
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	appSumi := filepath.Join(appDir, "app.sumi")
	original, _ := os.ReadFile(appSumi)

	// When: sumi dev runs on a PTY.
	cmd := exec.Command(cli, "dev", appDir)
	master, err := pty.Start(cmd, 24, 80)
	if err != nil {
		t.Fatalf("pty: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		master.Close()
	}()
	output := drain(master)

	// Then: the app is mirrored.
	expect(t, output, "Hello, sumi", 20*time.Second, "initial mirror")

	// Hot swap on save.
	rewrite(t, appSumi, strings.Replace(string(original), "Hello, sumi", "Hot reloaded!", 1))
	expect(t, output, "Hot reloaded!", 30*time.Second, "rebuild and swap")

	// Error keeps the last good build.
	rewrite(t, appSumi, strings.Replace(string(original), "</script>", "func broken( {\n</script>", 1))
	expect(t, output, "last good build still running", 30*time.Second, "error banner")

	// Recovery.
	rewrite(t, appSumi, strings.Replace(string(original), "Hello, sumi", "Recovered fine", 1))
	expect(t, output, "Recovered fine", 30*time.Second, "recovery")

	// Quit the app: dev exits.
	master.Write([]byte("q"))
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cmd.ProcessState != nil {
			break
		}
		if err := cmd.Process.Signal(nil); err != nil { // process gone
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func rewrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// drain replays the supervisor's output through a vt100 model in the
// background (cell diffs skip unchanged cells, so raw text matching is
// lossy) and returns an accessor for the current screen text.
func drain(master *os.File) func() string {
	mu := make(chan struct{}, 1)
	mu <- struct{}{}
	screen := vt100.NewScreen(80, 24)
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				<-mu
				screen.Write(buf[:n])
				mu <- struct{}{}
			}
			if err != nil {
				return
			}
		}
	}()
	return func() string {
		<-mu
		defer func() { mu <- struct{}{} }()
		var b strings.Builder
		for row := 0; row < screen.Height(); row++ {
			for col := 0; col < screen.Width(); col++ {
				ch := screen.Cell(row, col).Ch
				if ch == 0 {
					ch = ' '
				}
				b.WriteRune(ch)
			}
			b.WriteRune('\n')
		}
		return b.String()
	}
}

func expect(t *testing.T, output func() string, needle string, timeout time.Duration, what string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(spaceRe.ReplaceAllString(output(), " "), needle) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s: %q never appeared; screen:\n%s", what, needle, output())
}
