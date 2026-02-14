package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tomyan/sumi/cmd/sumi/preview"
	"github.com/tomyan/sumi/runtime/pty"
	"github.com/tomyan/sumi/runtime/sumitest"
)

func testPreview(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}

	testFunc, err := findServeTest(absDir)
	if err != nil {
		return err
	}

	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("sumi-preview-%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	// Start the test subprocess on a PTY.
	cmd := exec.Command("go", "test",
		"./"+relPath(absDir),
		"-run", "^"+testFunc+"$",
		"-serve", "-count=1", "-timeout=0",
	)
	cmd.Dir = findModuleRoot(absDir)
	cmd.Env = append(os.Environ(), "SUMI_CONTROL_SOCKET="+socketPath)

	master, err := pty.Start(cmd, 24, 80)
	if err != nil {
		return fmt.Errorf("start subprocess: %w", err)
	}
	defer master.Close()
	defer cmd.Process.Kill()

	// Wait for the socket to appear.
	if err := waitForSocket(socketPath); err != nil {
		return fmt.Errorf("subprocess did not create socket: %w", err)
	}

	client, err := sumitest.Connect(socketPath)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	info, err := client.Info()
	if err != nil {
		return fmt.Errorf("info: %w", err)
	}

	// Resize the PTY to match the component's dimensions.
	pty.SetSize(master, info.Height, info.Width)

	preview.Setup(client, master, info, absDir)
	preview.RunPreview()

	client.Quit()
	cmd.Wait()
	return nil
}

// serveTestPattern matches test functions that call ServeMode().
var serveTestPattern = regexp.MustCompile(`func\s+(Test\w+Serve\w*)\s*\(`)

// findServeTest scans *_test.go files in dir for a function containing ServeMode().
func findServeTest(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		if !strings.Contains(content, "ServeMode()") {
			continue
		}
		if m := serveTestPattern.FindStringSubmatch(content); m != nil {
			return m[1], nil
		}
	}
	return "", fmt.Errorf("no serve test found in %s (look for ServeMode() in *_test.go)", dir)
}

// waitForSocket polls for the socket file to appear.
func waitForSocket(path string) error {
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(path); err == nil {
			// Socket exists — give the server a moment to start accepting.
			time.Sleep(10 * time.Millisecond)
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s", path)
}

// findModuleRoot walks up from dir looking for go.mod.
func findModuleRoot(dir string) string {
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		if d == filepath.Dir(d) {
			return dir // fallback
		}
	}
}

// relPath returns the directory relative to the module root, suitable for go test.
func relPath(absDir string) string {
	modRoot := findModuleRoot(absDir)
	rel, err := filepath.Rel(modRoot, absDir)
	if err != nil {
		return absDir
	}
	return rel
}
