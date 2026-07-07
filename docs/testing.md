# Testing

Sumi is built to be tested without a real terminal. Rendering is synchronous
and deterministic in test mode, events are plain values you construct and feed
in, and the output is a cell buffer you can read directly. On top of that sit
snapshot scenarios for whole-frame assertions, an in-repo terminal model for
verifying the actual bytes sumi emits, and PTY-driven subprocess tests for the
few behaviours that need a real terminal file descriptor.

## Component tests with TestApp

`tui.TestApp(comp, w, h)` builds an app that renders into an in-memory buffer of
the given size and does an initial render immediately. You then read cells or
step events — no goroutines, no stdout, no timing.

```go
comp := &tui.Component{
    Tree: &layout.Input{
        Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
        Children: []*layout.Input{
            {Kind: layout.KindText, Content: "Hello"},
        },
    },
}
app := tui.TestApp(comp, 20, 3)

if app.TestBuffer.Cell(0, 0).Ch != 'H' {
    t.Errorf("Cell(0,0) = %c, want 'H'", app.TestBuffer.Cell(0, 0).Ch)
}
```

`app.Step(evt)` dispatches one event and then converges the render (bounded to a
few passes so a signal change that triggers another render settles before Step
returns):

```go
app.Step(input.Event{Kind: input.EventKey, Rune: '+'})
```

Each `render.Cell` exposes its `Ch` rune and `Style`, so you can assert on both
content and styling. `app.Render()` forces a fresh frame after you mutate a
signal directly. In practice you build the component through generated
constructors rather than hand-writing the `layout.Input` tree — the tree above
is spelled out only to show what the buffer holds.

## Snapshot scenarios

For asserting whole frames across a sequence of interactions, define a
`sumitest.Scenario`: a name, a viewport size, a factory that returns a test app,
and a list of steps. A step with a nil `Action` captures the current frame
without dispatching anything; a step with an `Action` runs it (usually feeding
an event) and captures the result.

```go
func clickerScenario() sumitest.Scenario {
    return sumitest.Scenario{
        Name:   "clicker-basics",
        Width:  30,
        Height: 6,
        NewApp: func(w, h int) *tui.App {
            comp := NewApp(AppProps{})
            return tui.TestApp(comp, w, h)
        },
        Steps: []sumitest.Step{
            {Name: "initial"},
            {Name: "after-click", Action: func(h *sumitest.Harness) {
                h.Step(sumitest.ClickEvent(1, 3)) // row 1, col 3
            }},
        },
    }
}

func TestClickerSnapshots(t *testing.T) {
    sumitest.AssertSnapshots(t, clickerScenario())
}
```

`sumitest` provides event constructors so tests read clearly: `KeyEvent(r)`,
`CtrlEvent(r)`, `SpecialEvent(k)`, `PasteEvent(text)`, `EnterEvent()`,
`EscapeEvent()`, `BackspaceEvent()`, `TabEvent()`, `ShiftTabEvent()`,
`ClickEvent(row, col)`, `DragEvent(row, col)`, and the scroll variants. Note
that click and drag take **row first, then column**.

`AssertSnapshots` compares each captured frame's styled text against a file in
`testdata/<name>.snapshot`. The file stores each frame under a `=== Frame: name
===` header:

```
=== Frame: initial ===
┌──────────────┐
│[ Click me ]  │
└──────────────┘
Count: 0

=== Frame: after-click ===
…
```

Run the tests with the `-update` flag to (re)write the snapshot files instead
of comparing:

```sh
go test ./... -update
```

Review the resulting diff as you would any golden file — a snapshot change is a
visible change to what the user sees. The `Harness` also exposes `Text()`,
`StyledText()`, `Buffer()` and `Resize(w, h)` for ad-hoc assertions, and
`sumitest` has `AssertText`, `AssertStyledText`, `AssertContains` and
`AssertStyledContains` helpers for checking a single frame without a snapshot.

## Verifying emitted bytes with the vt100 model

The buffer tests above check what sumi *intends* to draw. To check the actual
escape sequences it *emits*, replay them through the in-repo terminal model.
Run a real app against in-memory streams, then feed the captured output to a
`vt100.Screen` and assert on the reconstructed cells:

```go
var out bytes.Buffer
tui.RunWithOptions(comp, tui.RunOptions{
    In: strings.NewReader("q"), Out: &out, ExitOn: []string{"q"},
})

screen := vt100.NewScreen(80, 24)
if _, err := screen.Write(out.Bytes()); err != nil {
    t.Fatalf("vt100 replay: %v", err)
}

cell := screen.Cell(0, 0)
if !cell.Style.Bold {
    t.Errorf("expected bold at (0,0)")
}
```

This is a genuine round trip: sumi's emitted bytes are parsed back into a screen
by an independent model, so a sequence that a real terminal would misread fails
the test. It also covers diffed updates — feeding a stream that drives several
frames and asserting the final reconstructed screen confirms the incremental
redraw logic, not just the first paint.

## PTY subprocess tests

A few behaviours need a real terminal file descriptor — job control being the
clearest example. Those tests build the app as a subprocess, start it on a
pseudo-terminal with `pty.Start(cmd, rows, cols)`, drive it by writing control
bytes to the master (for instance `0x1a` for Ctrl+Z), and poll the accumulated
output with a deadline. They are skipped under `go test -short`, since building
and spawning a subprocess is slower than an in-memory test. Reach for a PTY test
only when the in-memory streams above cannot reproduce the behaviour.

## Previewing scenarios

`sumi test-preview <component-dir>` runs a scenario interactively in a real
terminal for manual inspection. It generates a temporary program that serves the
scenario over a Unix socket, runs it on a PTY, and drives stepping from a
preview UI, showing the rendered frames alongside the source. This is a
development aid for eyeballing frames, not part of the automated suite — the
snapshot tests above are what run in CI.
