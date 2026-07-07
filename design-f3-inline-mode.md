# F3 design — inline mode + FrameLog

Status: DRAFT. Reference: svelterm `src/render/inline.ts` (InlineScreen
driver), `src/framelog.ts`, run-loop wiring in `src/index.ts`
(scout report 2026-07-06; shipped semantics, not the aspirational
DESIGN-inline-mode.md).

## Model (from the reference)

Two zones. **Archived** rows sit above the render origin and belong to
the terminal's native scrollback — never touched again. The **live
zone** is a content-sized region at the shell cursor, repainted by
cell diff. The zone's absolute screen position is deliberately
unknown: the single load-bearing invariant is **never emit an absolute
row coordinate** — rows move relatively (CUU/CUD), columns absolutely
(CHA `CSI n G`), growth appends LF newlines (LF scrolls where CUD
cannot), shrink erases with ED (`CSI 0J`), and archiving emits
*nothing* (the rows already look right; the driver just narrows its
comparison window and shifts its origin bookkeeping).

## Sumi mapping

### F3a — InlineScreen driver (`runtime/render/inline_screen.go`)

Port the state machine verbatim:

```go
type InlineScreen struct {
    prev          *Buffer // last painted, padded to physicalRows
    physicalRows  int     // lines realised on the terminal (shrink keeps blanks)
    contentHeight int     // rows the last frame actually used
    cursorRow     int     // relative to zone origin
    cursorCol     int     // -1 = wrap-pending
}
func (s *InlineScreen) Render(next *Buffer) []byte // diff → relative-move ANSI
func (s *InlineScreen) ReleaseTop(n int)           // archive: zero output
func (s *InlineScreen) Finish() []byte             // park cursor after content
func (s *InlineScreen) Reset()                     // suspend: forget the zone
```

Width change at Render top: erase live zone in place (`moveRow(0) \r
CSI 0J`), nil `prev`, full repaint; archived rows may mis-wrap
(accepted, per reference). Style runs reuse the SGR emission sumi's
absolute diff already uses. Unit tests mirror svelterm's
inline-screen tests: first render uses `\n` not CUP; no `CSI r;cH`
anywhere in output; grow emits LFs; shrink emits `CSI 0J`; ReleaseTop
emits nothing and later diffs align against shifted rows.

### F3b — run-mode wiring (`RunOptions.Inline bool`)

- Setup: raw mode, hide cursor, bracketed paste, kitty push — but **no
  alternate screen, no clear** (enterTerminal grows an inline variant).
- Render: inline always renders fully (the zone is content-sized; any
  change can move everything). Frame height = clamp(content extent,
  1, terminal height) — taller UIs truncate; archiving is how authors
  keep the zone short. Content extent = root box height from Layout.
- Exit: `Finish()` leaves the final frame in the scrollback (no clear).
- Suspend: `Reset()` before stopping; fresh zone wherever the shell
  leaves the cursor on resume. Resize: full repaint of the live zone.
- DEC 2026 sync wrapping stays for frames; mode-setup writes bypass it.
- Logs: RunOptions.OnLog (F4b) is the sanctioned channel; direct
  stdout writes desync the zone (documented; Go cannot intercept fmt).

### F3c — FrameLog (`runtime/tui/framelog.go`)

Frames are real sumi components (composition already works by
embedding `child.Tree`):

```go
type FrameLog struct { Host *layout.Input /* place in your tree */ }
func NewFrameLog() *FrameLog
func (l *FrameLog) Append(c *Component) int  // splice c.Tree under Host
func (l *FrameLog) Update — NOT provided: mutate the frame component's
    signals; sumi re-renders on dirty (svelterm's Update is just a
    props assign that relies on reactivity anyway)
func (l *FrameLog) Archive(id int)  // cumulative from top: sum laid-out
    heights (Input.LastH) of frames 0..id, InlineScreen.ReleaseTop(sum),
    Dispose each, drop from Host
func (l *FrameLog) Remove(id int)   // dispose one; zone reflows/shrinks
func (l *FrameLog) LiveFrames() []int
```

Fullscreen mode: Archive still disposes/unsplices but ReleaseTop is a
no-op (hooks only wired when inline).

## Decision (RESOLVED with Tom, 2026-07-06): CPR mouse included in F3

Full parity with current svelterm in one arc: CPR (`CSI 6n`) origin
discovery on start/resize/resume, origin shifted by ReleaseTop and by
zone growth, screen→zone mapping with the bottom-clamp rule
(`effectiveOrigin = min(originRow, termH - physicalRows + 1)`),
events above the zone dropped; terminals without CPR reply keep mouse
silently off in-zone (rendering never needs the origin).

Deferred: svelterm's embedded-preview combo mode (fullscreen render
without altscreen) — sumi's preview drives PTYs instead.

## Slices

1. **F3a** InlineScreen driver + unit tests (pure; no run-loop risk).
2. **F3b** RunOptions.Inline wiring + PTY integration test (run real
   app inline, assert no absolute CUP in output, final frame persists
   after exit, shell prompt lands below it).
3. **F3c** FrameLog + example (streaming frames archived into
   scrollback) + tests with a fake InlineScreen hook.
4. **F3d** CPR origin discovery + screen→zone mouse mapping
   (bottom-clamp; selection/hover/click work inside the live zone).
