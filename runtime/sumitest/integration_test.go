package sumitest

import (
	"fmt"
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

// createCounterApp builds, by hand, the test app a generated counter
// component would produce via tui.TestApp(NewCounter(...)):
//
//	<script>
//	  count := sumi.New(0)
//	  func increment() { count.Update(func(n int) int { return n + 1 }) }
//	</script>
//	<div onkey="increment">Count: {count}</div>
func createCounterApp(w, h int) *tui.App {
	count := 0

	var app *tui.App

	node0 := &layout.Input{
		Kind:    layout.KindText,
		Content: fmt.Sprintf("Count: %d", count),
	}
	root := &layout.Input{
		Kind:      layout.KindBox,
		Direction: "column",
		CursorCol: -1,
		CursorRow: -1,
		Children:  []*layout.Input{node0},
	}

	sync := func() {
		node0.Content = fmt.Sprintf("Count: %d", count)
	}

	var prevTree *layout.Box
	var prevW, prevH int

	doRender := func() {
		sync()
		var termW, termH int
		if app.TestWidth > 0 {
			termW, termH = app.TestWidth, app.TestHeight
		} else {
			termW, termH = 80, 24
		}
		tree := layout.Layout(root, termW, termH)
		changes, scrollChanged := layout.DiffTrees(prevTree, tree)
		if prevTree == nil || termW != prevW || termH != prevH || scrollChanged || tree.HasOverlap {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			if app.TestBuffer != nil {
				app.TestBuffer = buf
			}
		} else if app.TestBuffer != nil {
			buf := render.NewBuffer(termW, termH)
			layout.RenderTree(buf, tree, nil)
			app.TestBuffer = buf
		}
		prevTree = tree
		prevW = termW
		prevH = termH
		_ = changes
	}

	increment := func() {
		count++
		app.Dirty = true
	}

	app = &tui.App{
		OnRender: doRender,
		OnEvent: func(evt input.Event) {
			if evt.Kind == input.EventKey {
				increment()
			}
		},
	}
	app.TestWidth = w
	app.TestHeight = h
	app.TestBuffer = render.NewBuffer(w, h)
	app.Render()
	return app
}

func TestIntegrationCounterInitialRender(t *testing.T) {
	// Given
	app := createCounterApp(20, 3)
	h := New(app)

	// Then — initial render shows count 0
	AssertContains(t, h, "Count: 0")
}

func TestIntegrationCounterAfterKeyPress(t *testing.T) {
	// Given
	app := createCounterApp(20, 3)
	h := New(app)

	// When — press a key to increment
	h.Step(KeyEvent('x'))

	// Then — count is now 1
	AssertContains(t, h, "Count: 1")
}

func TestIntegrationCounterMultipleKeyPresses(t *testing.T) {
	// Given
	app := createCounterApp(20, 3)
	h := New(app)

	// When — press 3 keys
	h.Step(KeyEvent('a'))
	h.Step(KeyEvent('b'))
	h.Step(KeyEvent('c'))

	// Then — count is now 3
	AssertContains(t, h, "Count: 3")
}

func TestIntegrationCounterTextOutput(t *testing.T) {
	// Given
	app := createCounterApp(20, 3)
	h := New(app)
	h.Step(KeyEvent('x'))

	// When
	text := h.Text()

	// Then — plain text contains the counter
	got := text
	if got == "" {
		t.Fatal("Text() returned empty string")
	}
	AssertContains(t, h, "Count: 1")
}

func TestCounterSnapshots(t *testing.T) {
	AssertSnapshots(t, counterScenario())
}

func TestCounterPreview(t *testing.T) {
	if !PreviewMode() {
		t.Skip("run with -preview to see interactive preview")
	}
	Preview(counterScenario())
}
