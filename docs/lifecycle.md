# Application lifecycle

This chapter is about the runtime around a component: how an app starts,
how the render loop turns signal writes into frames, how animation and
resize hook in, and how it stops.

## Starting up

`tui.Run(component)` runs a component full-screen. It puts the terminal
into raw mode, switches to the alternate screen, enables bracketed paste,
the kitty key protocol, and (by default) mouse tracking, probes the
terminal background colour for `light-dark()`, then enters the event
loop. On exit it restores all of that.

`tui.RunWithOptions(component, opts)` is the same with configuration.
`Run` is exactly `RunWithOptions(component, RunOptions{})`. The options
that matter to component authors:

```go
tui.RunWithOptions(NewApp(AppProps{}), tui.RunOptions{
	Inline:      true,               // render at the shell cursor, no alt screen
	Mouse:       &enabled,           // *bool; overrides mouse auto-detection
	ColorScheme: "dark",             // "light"/"dark" forces the scheme
	ExitOn:      []string{"ctrl+c", "q"},
	In:          os.Stdin,           // input stream  (nil = os.Stdin)
	Out:         os.Stdout,          // output stream (nil = os.Stdout)
})
```

- **`Inline`** renders a live zone at the shell cursor instead of taking
  the alternate screen; the final frame is left in scrollback. See
  [inline mode](inline-mode.md).
- **`Mouse`** is a `*bool`: leave it nil to auto-detect, or point it at a
  value to force mouse mode on or off. It defaults on (sumi's own
  selection replaces the terminal's).
- **`ColorScheme`** forces `"light"` or `"dark"` and skips the background
  probe; `""` detects. **`ColorDepth`** (`render.ColorDepth`,
  `DepthAuto` by default) sets the emission depth.
- **`In` / `Out`** inject the terminal streams — useful for tests and for
  hosts without real file descriptors, paired with `Size func() (w, h int)`
  to supply the viewport (this is how the wasm host runs).

`ExitOn`, `OnResize`, `OnPostRender`, `SetApp`, and `OnLog` are covered
below where they're relevant.

## The render loop

There is no polling. A single dirty flag drives rendering: a signal
write, a handled event, or a resize marks the app dirty, and the loop
renders once things settle. The loop blocks waiting for an input event, a
resize, an OS signal, an animation wake, or a queued function; it drains
everything pending, then *converges* — it renders up to three passes
while the dirty flag is still set, so a render that itself dirties state
(a self-measuring layout, say) resolves in the same frame.

Each render pass resolves the root component's CSS, steps any running
length transitions, lays the tree out, projects the form controls'
state, paints a buffer, and diffs it against the previous frame so only
changed cells are written. If you set `RunOptions.OnPostRender`, it runs
once after each converged frame — the hook used to paint custom content
into laid-out regions.

## Animation

`RequestFrame()` on the `*App` schedules one more frame about 16 ms
later (delivered as an `EventFrame`); call it from a frame handler to
keep animating. The runtime also calls it for you while CSS transitions
or animations are active, so declarative motion needs no manual driving
(see [motion](motion.md)).

To reach the `*App` from outside a handler — to call `RequestFrame`,
`Wake` (render now), or `Do` (run a function on the render goroutine) —
capture it with `SetApp`:

```go
var app *tui.App
tui.RunWithOptions(NewApp(AppProps{}), tui.RunOptions{
	SetApp: func(a *tui.App) { app = a },
})
```

`app.Wake()` triggers an immediate re-render; background goroutines use
it to push new content in. `app.Do(fn)` runs `fn` on the render
goroutine, where mutating signals and the tree is safe.

## Resize and suspend

A terminal resize marks the app dirty and re-lays-out; the `$env` width
and height signals update every render, so responsive layout follows the
new size. If you pass `RunOptions.OnResize`, it runs on each resize.

Ctrl+Z suspends the process (SIGTSTP; it resumes on SIGCONT). This is a
keydown default action: the runtime restores the terminal, stops, and on
resume re-enters the terminal and forces a clean full repaint. A keydown
handler can `PreventDefault()` to keep Ctrl+Z for itself.

## Stopping

`sumi.Quit()` ends the app from any handler — it signals the running app
to leave its loop. Two things also stop it without your writing the call:

- **Exit chords**: after your `handleKey` runs, an event matching
  `RunOptions.ExitOn` quits. The default is `["ctrl+c"]`; an explicit
  list replaces it. A chord is `"ctrl+<letter>"`, a special-key name
  (`"escape"`, `"enter"`), or a single character.
- **OS signals**: `SIGINT`/`SIGTERM` arrive as an `EventSignal` to
  `handleKey`. They are *not* auto-handled for a component — quit on them
  yourself: `if evt.Kind == sumi.EventSignal { sumi.Quit() }`.

### Dispose

`*sumi.Component` has a `Dispose func()` field. When set, the runtime
calls it once when the app exits, and when a [FrameLog](inline-mode.md)
frame is archived or removed. It is a teardown hook for embedding and
hand-written code; generated `.sumi` components do not populate it today,
and signals and effects have no separate per-component teardown — they
live as long as the app does.
