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
	socketPath := filepath.Join(t.TempDir(), "test.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	stdout := createTempFile(t)
	defer stdout.Close()

	scenario := testScenario()
	go ServeOnListener(listener, scenario, stdout)

	// When — connect and send info
	client, err := Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

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

	client.Quit()
}

func TestServeStep(t *testing.T) {
	// Given — a serve listener with a test scenario
	socketPath := filepath.Join(t.TempDir(), "test.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	stdout := createTempFile(t)
	defer stdout.Close()

	scenario := testScenario()
	go ServeOnListener(listener, scenario, stdout)

	client, err := Connect(socketPath)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer client.Close()

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

	// Verify ANSI was written to stdout (should contain content)
	stdout.Seek(0, 0)
	data := make([]byte, 4096)
	n, _ := stdout.Read(data)
	if n == 0 {
		t.Error("stdout: expected ANSI output")
	}

	client.Quit()
}

func createTempFile(t *testing.T) *os.File {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	return f
}
