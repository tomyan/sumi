package sumitest

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

func testScenario() Scenario {
	return Scenario{
		Name:   "test-serve",
		Width:  10,
		Height: 3,
		NewApp: func(w, h int) *tui.App {
			textNode := &layout.Input{Kind: layout.KindText, Content: "Hello"}
			root := &layout.Input{
				Kind:      layout.KindBox,
				Direction: "column",
				CursorCol: -1,
				CursorRow: -1,
				Children:  []*layout.Input{textNode},
			}
			var app *tui.App
			var prevTree *layout.Box
			app = &tui.App{
				OnRender: func() {
					tree := layout.Layout(root, app.TestWidth, app.TestHeight)
					buf := render.NewBuffer(app.TestWidth, app.TestHeight)
					layout.RenderTree(buf, tree, nil)
					app.TestBuffer = buf
					prevTree = tree
					_ = prevTree
				},
			}
			app.TestWidth = w
			app.TestHeight = h
			app.TestBuffer = render.NewBuffer(w, h)
			app.Render()
			return app
		},
		Steps: []Step{
			{Name: "initial"},
			{Name: "second"},
		},
		SourceFile: "test.sumi",
	}
}

func TestServeInfo(t *testing.T) {
	// Given — a serve listener with a test scenario
	client, stdout, cleanup := startServe(t, testScenario())
	defer cleanup()
	_ = stdout

	// When — connect and send info
	info, err := client.Info()
	if err != nil {
		t.Fatalf("info: %v", err)
	}

	// Then — metadata matches scenario
	if info.Name != "test-serve" {
		t.Errorf("name: got %q, want %q", info.Name, "test-serve")
	}
	if info.Width != 10 {
		t.Errorf("width: got %d, want 10", info.Width)
	}
	if info.Height != 3 {
		t.Errorf("height: got %d, want 3", info.Height)
	}
	if len(info.Steps) != 2 {
		t.Fatalf("steps: got %d, want 2", len(info.Steps))
	}
	if info.Steps[0] != "initial" {
		t.Errorf("step[0]: got %q, want %q", info.Steps[0], "initial")
	}
	if info.SourceFile != "test.sumi" {
		t.Errorf("source: got %q, want %q", info.SourceFile, "test.sumi")
	}
}

func TestServeStep(t *testing.T) {
	// Given — a serve listener with a test scenario
	client, stdout, cleanup := startServe(t, testScenario())
	defer cleanup()

	// When — send step 0
	resp, err := client.Step(0)
	if err != nil {
		t.Fatalf("step: %v", err)
	}

	// Then — response has step metadata
	if resp.Name != "initial" {
		t.Errorf("name: got %q, want %q", resp.Name, "initial")
	}
	if resp.StyledText == "" {
		t.Error("styled_text: expected non-empty")
	}

	// Verify ANSI was written to stdout
	stdout.Seek(0, 0)
	data := make([]byte, 4096)
	n, _ := stdout.Read(data)
	if n == 0 {
		t.Error("stdout: expected ANSI output")
	}
}

func TestServeInputCommand(t *testing.T) {
	// Given — a stateful scenario stepped to initial state
	client, _, cleanup := startServe(t, statefulScenario())
	defer cleanup()

	_, err := client.Step(0)
	if err != nil {
		t.Fatalf("step 0: %v", err)
	}

	// When — send an input event (type 'x')
	resp, err := client.Input(KeyEvent('x'))
	if err != nil {
		t.Fatalf("input: %v", err)
	}

	// Then — the app state changes to reflect the key press
	if resp.StyledText == "" || !strContains(resp.StyledText, "typed-x") {
		t.Errorf("styled: expected to contain 'typed-x', got %q", resp.StyledText)
	}
}

// startServe creates a listener, starts ServeOnListener, and connects a client.
// Returns the client, stdout file, and a cleanup function.
func startServe(t *testing.T, s Scenario) (*Client, *os.File, func()) {
	t.Helper()
	socketPath := filepath.Join(t.TempDir(), "test.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	stdout := createTempFile(t)
	go ServeOnListener(listener, s, stdout)

	client, err := Connect(socketPath)
	if err != nil {
		listener.Close()
		stdout.Close()
		t.Fatalf("connect: %v", err)
	}

	cleanup := func() {
		client.Quit()
		client.Close()
		listener.Close()
		stdout.Close()
	}
	return client, stdout, cleanup
}

func createTempFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	return f
}

// strContains checks if s contains substr.
func strContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
