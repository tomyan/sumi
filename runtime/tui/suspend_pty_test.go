package tui_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tomyan/sumi/runtime/pty"
)

// D4b integration: a real app on a PTY stops on Ctrl+Z and repaints on
// SIGCONT.

func TestCtrlZStopsAndResumesOnPTY(t *testing.T) {
	if testing.Short() {
		t.Skip("PTY subprocess test")
	}

	// Given: the suspendapp built and running on a PTY.
	bin := filepath.Join(t.TempDir(), "suspendapp")
	build := exec.Command("go", "build", "-o", bin, "./testdata/suspendapp")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	cmd := exec.Command(bin)
	// The PTY child is a session leader (orphaned process group), so a
	// default SIGTSTP would be discarded per POSIX; the app substitutes
	// SIGSTOP under this env var — the rest of the path is identical.
	cmd.Env = append(os.Environ(), "SUSPEND_TEST_SIGSTOP=1")
	master, err := pty.Start(cmd, 24, 80)
	if err != nil {
		t.Fatalf("pty start: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		master.Close()
	}()
	output := drainPTY(master)
	waitFor(t, "initial paint", func() bool {
		return strings.Contains(output(), "suspend me")
	})

	// When: Ctrl+Z arrives.
	if _, err := master.Write([]byte{0x1a}); err != nil {
		t.Fatalf("write ctrl+z: %v", err)
	}

	// Then: the process stops. Raw status check: BSD reports SIGCONT as
	// stop-signal==SIGSTOP, so Go's WaitStatus.Stopped() is false for a
	// SIGSTOP-stopped child; 0x7f in the low byte means stopped.
	var ws syscall.WaitStatus
	waitFor(t, "process stopped", func() bool {
		wpid, err := syscall.Wait4(cmd.Process.Pid, &ws, syscall.WUNTRACED|syscall.WNOHANG, nil)
		return err == nil && wpid == cmd.Process.Pid && ws&0x7f == 0x7f
	})

	// When: resumed.
	before := len(output())
	if err := syscall.Kill(cmd.Process.Pid, syscall.SIGCONT); err != nil {
		t.Fatalf("sigcont: %v", err)
	}

	// Then: the app re-enters the terminal and repaints.
	waitFor(t, "repaint after resume", func() bool {
		resumed := output()[before:]
		return strings.Contains(resumed, "suspend me")
	})
}

// drainPTY reads master in the background, returning an accessor for
// everything read so far.
func drainPTY(master *os.File) func() string {
	var mu = make(chan struct{}, 1)
	mu <- struct{}{}
	collected := ""
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				<-mu
				collected += string(buf[:n])
				mu <- struct{}{}
			}
			if err != nil {
				return
			}
		}
	}()
	return func() string {
		<-mu
		s := collected
		mu <- struct{}{}
		return s
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}
