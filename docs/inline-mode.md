# Inline mode

By default a sumi app takes over the whole terminal using the alternate screen,
like a full-screen TUI. Inline mode renders at the shell cursor instead,
drawing a live region in the normal buffer that grows and shrinks in place and
leaves its final frame behind in scrollback when the app exits. This is the
model behind installers, progress displays, prompts, and REPL-style tools that
should read as ordinary command output rather than take over the screen.

## Enabling it

Inline mode is a field on `RunOptions`:

```go
tui.RunWithOptions(comp, tui.RunOptions{
    Inline: true,
})
```

`tui.Run(comp)` is the full-screen shortcut and leaves `Inline` false. The
other `RunOptions` fields (`In`, `Out`, `ExitOn`, `ColorScheme`, `Mouse`, and
so on) work the same in both modes.

## The live-zone model

Inline mode maintains a *live zone*: a block of rows at the shell cursor that
sumi owns and repaints. Crucially, it never learns the zone's absolute position
on screen — it moves the cursor *relatively*. Rows move with cursor-up and
cursor-down; columns move with an absolute column command (`CHA`), which is
safe because column 1 is always column 1. It never emits an absolute cursor
position, and it never enters the alternate screen.

Growth and shrink are handled without scrolling surprises:

- **Growing** the zone appends real newlines (line feeds), because a line feed
  can scroll the viewport when the cursor is at the bottom whereas a cursor-down
  cannot.
- **Shrinking** erases from the cursor to the end of the display and leaves the
  now-blank physical rows realised, so re-growing back into them repaints
  without emitting fresh newlines.

Each render diffs the new frame against the previous one and rewrites only the
cells that changed, moving relatively between them. A width change erases the
zone in place and repaints it fully.

## Exit and scrollback

On exit sumi parks the cursor on a fresh line just below the rendered content
and shows it again. The final frame is not erased — it stays in the terminal's
scrollback as normal output, so the last state of the app remains visible in
the session history, exactly like the output of any other command.

## Suspend

When the user suspends the app with Ctrl+Z, inline mode finishes the current
frame (parking the cursor below it) and then forgets the zone entirely. While
the process is stopped the shell owns the screen, so on resume (SIGCONT) sumi
starts a fresh zone wherever the cursor now is and re-discovers its origin (see
mouse, below) rather than assuming the old position is still valid.

## The frame log

For append-only output — a stream of results, log lines, or completed
steps — inline mode offers a `FrameLog`. Each frame is a full sumi component
mounted into the live zone; frames stack in block flow under a host container
you place in your tree. The point of a frame log is that finished frames can be
handed to the terminal's native scrollback with no repaint.

```go
log := tui.NewFrameLog()
log.ReleaseTop = app.ReleaseTop        // wire archiving to the inline app

id := log.Append(resultComponent)      // mount a new frame, returns its id
// update a live frame by writing its signals; the next render reflects it
log.Archive(id)                        // hand its rows (and any above it) to scrollback
```

The three operations:

- **`Append(c *Component) int`** mounts a component as a new frame at the
  bottom of the log and returns its id.
- **`Archive(id int)`** releases the rows of that frame *and every frame above
  it* into native scrollback, then disposes those components. Archiving is
  cumulative from the top, which matches how output scrolls off: everything
  older than the archived frame goes with it.
- **`Remove(id int)`** disposes a single frame *without* archiving it — its rows
  are cleared and the zone reflows. Use this to redact a frame rather than
  commit it to history.

Wiring `ReleaseTop` to the app's `ReleaseTop` is what connects archiving to real
scrollback; the released rows already sit on the terminal, so handing them over
emits nothing and simply narrows the region sumi keeps diffing. Left nil (or in
full-screen mode) `Archive` becomes dispose-only.

## Mouse

Mouse events arrive from the terminal in absolute screen coordinates, but the
live zone only knows its own rows. To translate, inline mode discovers its
origin with a cursor-position report: it snapshots the cursor's zone row, sends
a Device Status Report (`ESC [ 6n`), and when the terminal replies with the
current cursor position it computes the screen row of zone row 0. It re-queries
after resize and after suspend, since either can move the zone.

With the origin known, incoming mouse rows are mapped back into zone
coordinates; a click outside the zone (or before the origin is known) is
dropped rather than misattributed. If the zone has grown and scrolled up so its
bottom is pinned to the screen bottom, the origin is clamped accordingly so the
mapping stays correct.
