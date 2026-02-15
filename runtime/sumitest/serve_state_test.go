package sumitest

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

// statefulScenario creates a scenario where each step dispatches key events
// that change the text, allowing us to verify incremental vs fresh stepping.
func statefulScenario() Scenario {
	return Scenario{
		Name:   "stateful",
		Width:  20,
		Height: 3,
		NewApp: func(w, h int) *tui.App {
			text := "step0"
			textNode := &layout.Input{Kind: layout.KindText, Content: text}
			root := &layout.Input{
				Kind:      layout.KindBox,
				Direction: "column",
				CursorCol: -1,
				CursorRow: -1,
				Children:  []*layout.Input{textNode},
			}
			var app *tui.App
			sync := func() { textNode.Content = text }
			app = &tui.App{
				OnRender: func() {
					sync()
					tree := layout.Layout(root, app.TestWidth, app.TestHeight)
					buf := render.NewBuffer(app.TestWidth, app.TestHeight)
					layout.RenderTree(buf, tree, nil)
					app.TestBuffer = buf
				},
				OnEvent: func(evt input.Event) {
					if evt.Kind == input.EventKey {
						text = "typed-" + string(evt.Rune)
						app.Dirty = true
					}
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
			{Name: "type-a", Action: func(h *Harness) {
				h.Step(KeyEvent('a'))
			}},
			{Name: "type-b", Action: func(h *Harness) {
				h.Step(KeyEvent('b'))
			}},
		},
	}
}

// wrapNewAppCounter wraps a scenario's NewApp to count invocations.
func wrapNewAppCounter(s *Scenario) *int {
	count := 0
	orig := s.NewApp
	s.NewApp = func(w, h int) *tui.App {
		count++
		return orig(w, h)
	}
	return &count
}

func TestServeStateStepTo(t *testing.T) {
	// Given — a serveState with the test scenario
	stdout := createTempFile(t)
	defer stdout.Close()

	st := newServeState(testScenario(), stdout)

	// When — step to index 0
	resp := st.stepTo(0)

	// Then — response matches expectations
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

func TestServeStateIncrementalForward(t *testing.T) {
	// Given — a stateful scenario that counts app creations
	stdout := createTempFile(t)
	defer stdout.Close()

	s := statefulScenario()
	appCreated := wrapNewAppCounter(&s)
	st := newServeState(s, stdout)

	// When — step forward 0 → 1 → 2
	resp0 := st.stepTo(0)
	resp1 := st.stepTo(1)
	resp2 := st.stepTo(2)

	// Then — forward stepping reuses app (only 1 creation, not 3)
	if *appCreated != 1 {
		t.Errorf("app created: got %d, want 1 (incremental reuse)", *appCreated)
	}

	// Each step shows the correct name
	if resp0.Name != "initial" {
		t.Errorf("step 0 name: got %q, want %q", resp0.Name, "initial")
	}
	if resp1.Name != "type-a" {
		t.Errorf("step 1 name: got %q, want %q", resp1.Name, "type-a")
	}
	if resp2.Name != "type-b" {
		t.Errorf("step 2 name: got %q, want %q", resp2.Name, "type-b")
	}

	// Each step shows the correct content
	if !strContains(resp0.StyledText, "step0") {
		t.Errorf("step 0 styled: expected 'step0', got %q", resp0.StyledText)
	}
	if !strContains(resp1.StyledText, "typed-a") {
		t.Errorf("step 1 styled: expected 'typed-a', got %q", resp1.StyledText)
	}
	if !strContains(resp2.StyledText, "typed-b") {
		t.Errorf("step 2 styled: expected 'typed-b', got %q", resp2.StyledText)
	}
}

func TestServeStateBackwardResets(t *testing.T) {
	// Given — stepped forward to step 2
	stdout := createTempFile(t)
	defer stdout.Close()

	s := statefulScenario()
	appCreated := wrapNewAppCounter(&s)
	st := newServeState(s, stdout)
	st.stepTo(0)
	st.stepTo(1)
	st.stepTo(2)

	// When — step backward to 0
	*appCreated = 0
	resp := st.stepTo(0)

	// Then — content reflects step 0 (fresh reset)
	if resp.Name != "initial" {
		t.Errorf("name: got %q, want %q", resp.Name, "initial")
	}
	if !strContains(resp.StyledText, "step0") {
		t.Errorf("styled: expected 'step0', got %q", resp.StyledText)
	}
	if *appCreated != 1 {
		t.Errorf("app created on backward: got %d, want 1", *appCreated)
	}
}
