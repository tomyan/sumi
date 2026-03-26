package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	modRoot := findModuleRoot(absDir)
	modPath, err := findModulePath(modRoot)
	if err != nil {
		return err
	}

	relDir, err := filepath.Rel(modRoot, absDir)
	if err != nil {
		return fmt.Errorf("rel path: %w", err)
	}
	importPath := modPath + "/" + filepath.ToSlash(relDir)

	// Generate a temp directory with main.go inside the module root
	// so that go run can resolve the local module.
	tmpDir, err := os.MkdirTemp(modRoot, ".sumi-preview-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	mainGo := fmt.Sprintf(`package main

import (
	scenario %q
	"github.com/tomyan/sumi/runtime/sumitest"
)

func main() {
	sumitest.Serve(scenario.Scenario())
}
`, importPath)

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("write main.go: %w", err)
	}

	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("sumi-preview-%d.sock", os.Getpid()))
	defer os.Remove(socketPath)

	// Build and run the temp main as a subprocess on a PTY.
	tmpRel, err := filepath.Rel(modRoot, tmpDir)
	if err != nil {
		return fmt.Errorf("rel temp path: %w", err)
	}

	cmd := exec.Command("go", "run", "./"+tmpRel)
	cmd.Dir = modRoot
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

	// Give subprocess time to exit gracefully, then force kill.
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		cmd.Process.Kill()
		<-done
	}

	return nil
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

// findModulePath reads the module path from go.mod.
func findModulePath(modRoot string) (string, error) {
	f, err := os.Open(filepath.Join(modRoot, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("module path not found in go.mod")
}
