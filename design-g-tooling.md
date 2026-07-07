# Phase G design — tooling (`sumi init` / `sumi dev` / `sumi inspect`)

Status: DRAFT for review with Tom. No Vite underneath us (svelterm gets
its dev loop for free): sumi compiles `.sumi → Go → binary`, so the dev
loop is ours to build. DX bar: Vite-quality. Existing machinery to
build on: `runtime/pty` (cross-platform), `runtime/vt100` (full
terminal model), `runtime/sumitest` control socket (Unix socket JSON
protocol), polling file watcher (`cmd/sumi/preview/watcher.go`),
`RunOptions` io injection.

## G1 — `sumi init`

`sumi init [dir]` scaffolds a runnable app:

- `app.sumi` — small canonical app (heading, counter button, quit
  hint) showing script/style/template + signals + an event handler.
- `main.go` — `//go:generate sumi generate .` + `Run` call.
- `go.mod` — `require github.com/tomyan/sumi`. Until the repo is
  published there is no fetchable module, so init adds
  `replace github.com/tomyan/sumi => <local checkout>` — located from
  the running `sumi` binary's module source (build info / `go env`),
  overridable with `--sumi-path`. Drop the replace at publish time
  (init flag flips default).
- Runs `sumi generate .` + `go mod tidy` so `go run .` works
  immediately; prints next steps (`sumi dev` to iterate).

## G2 — `sumi dev`

Watch → regenerate → rebuild → relaunch (HMR/state-preserving reload is
explicitly out of parity scope). The DX fork is *how the app runs
during the loop* — see Decision 1. Either way:

- Watch `.sumi` + `.go` (+ the sumi module itself when replace-directed,
  so framework hacking rebuilds apps too) with the existing polling
  watcher.
- Pipeline per change: `sumi generate .` → `go build` → swap in the
  new binary. Generate/compile errors NEVER leave a dead terminal.
- Keep-last-good: on error the running app stays up and interactive;
  the error shows without destroying it (Vite keeps serving the last
  good bundle). A successful build swaps the process.
- `sumi dev` always exports `SUMI_CONTROL_SOCKET` to the child so
  `sumi inspect` can attach (G3).

### Decision 1 (RESOLVED with Tom, 2026-07-06): PTY supervisor

**Option A — PTY supervisor (chosen).** `sumi dev` owns the real
terminal; the app runs in a PTY child mirrored to the screen
(byte-for-byte passthrough, input forwarded raw). Restarts are
seamless: the supervisor repaints the new child's frame with no
terminal-mode churn, no shell prompt flash. Build errors render as a
supervisor-drawn overlay/banner (dogfooding sumi for its own chrome —
optionally over a vt100 model of the dead app's last frame). Status
line (rebuild time, error count) available. This is the test-preview
architecture minus the editors, and it later hosts inspect overlays.
Risk: passthrough fidelity (mouse/kitty/paste sequences must forward
both ways — the PTY is transparent, so mostly free).

**Option B — bare exec loop.** `sumi dev` runs the app inheriting the
tty; on change: SIGTERM child (app restores terminal via its own
defers), rebuild, print errors plainly to the scrollback and wait for
the next save, relaunch on success. ~80 lines, zero passthrough risk;
but every rebuild flashes the shell, errors kill the running app, and
there is no chrome to build on for inspect/HMR later.

## G3 — `sumi inspect`

`sumi inspect [--socket path]` attaches to a running app (dev-launched
apps always listen) and dumps, svt-style:

- `tree` — element tree with tags/ids/classes/state flags.
- `styles <selector-or-path>` — computed style for a node (post-cascade
  Input fields + resolved render.Style).
- `boxes` — laid-out geometry (X/Y/W/H, fragments for inline runs).
- `watch` — stream re-render events (frame count, dirty causes).

Transport: extend the sumitest control protocol (JSON over the Unix
socket) with `inspect-tree` / `inspect-styles` / `inspect-boxes`
commands; the runtime side serializes from `comp.Tree` +
`comp.LayoutResult` on the event-loop goroutine via `app.Do` (no
races). The serve listener moves from sumitest-only into an opt-in
`tui` hook so real apps (not just scenario harnesses) can expose it.

## Slices

1. **G1** init scaffold + tests (golden scaffold, `go build` succeeds
   in a temp dir with replace).
2. **G2a** rebuild pipeline as a library (watch→generate→build with
   structured errors) + tests.
3. **G2b** the run/swap layer per Decision 1 + PTY integration test.
4. **G3a** control-socket inspect commands in the runtime + unit tests.
5. **G3b** `sumi inspect` CLI formatting + end-to-end test against a
   dev-launched app.
