---
name: sumi
description: Orientation for sumi — .sumi single-file components (Go script, CSS, HTML template) compiled to Go TUI binaries. Use when writing or reviewing .sumi components, working in ~/projects/sumi or sumi-site, or answering "does sumi support X". Covers the sumi-specific surface and where the full support matrix lives; standard HTML/CSS behaviour is not repeated here.
---

# sumi

Declarative TTY framework for Go. A `.sumi` file has three sections —
`<script>` (Go), `<style>` (CSS), and an HTML template — compiled by
`sumi generate` to plain Go that renders to a terminal cell grid.
Standard HTML/CSS semantics apply wherever they make sense on cells;
consult this only for the deltas.

**Authoritative docs:** `~/projects/sumi/docs/` — `reference.md` is the
support matrix; chapters (getting-started, components, elements,
layout, selectors, terminal-css, motion, inline-mode, testing,
distribution, terminals) carry worked detail. Rendered at
https://gosumi.dev/docs. Only fetch for a specific supported/unsupported
or how-to question.

## The mental model

- 1 cell = the layout atom; lengths in `cell`/`ch` (aliases), `%`,
  unitless ints. Pixel-derived units (px/em/rem, fonts, shadows,
  transforms) parse and drop silently.
- Signals, not runes: `count := sumi.New(0)` in script;
  `{count}` in the template subscribes; `count.Set/Update` re-renders.
  `$derived`-style computed values are plain Go expressions over
  `.Get()`.
- Templates are HTML: real elements with UA semantics (button/input/
  select/dialog/details focus + activation), `onclick={handler}`
  event attrs (DOM-style bubbling, stopPropagation/preventDefault),
  `bind:value`, `{if expr}…{else}…{/if}`,
  `{for i, x := range xs.Get() key=x}…{/for}` (Go range syntax).
- `<box>`/`<text>` are long gone — div/span etc. Container tags need
  explicit close (`<div></div>`); `<input />` self-closes.
- CSS is runtime-resolved: full cascade, selector engine (structural +
  state pseudos, attribute selectors), `var()`, `calc()`,
  `light-dark()`, `@media` (incl. `display-mode: terminal`),
  `@container`, `@supports`, transitions/animations/`@keyframes`.
  Colors: full CSS color 4 set, degraded to terminal depth.
  `opacity: dim` maps to the dim attribute.

## sumi-specific surface

- **Events**: handlers take `*sumi.DOMEvent` (DOM-style) or
  `sumi.Event` (raw input, via `onkey` attr). `sumi.Quit()`,
  `sumi.EventSignal` for OS signals. Function order matters in
  script: callee before caller.
- **Terminal extras**: `border: single|double|rounded|heavy|ascii`,
  `border-title`, tmux-style `border-collapse` on boxes, sticky
  positioning, scrollable overflow with styled scrollbars.
- **Runtime modes**: altscreen default; inline mode
  (`RunOptions.Inline`) streams frames into scrollback; wasm target
  (`runtime/webterm` + xterm.js) for the browser.
- **CLI**: `sumi init` (scaffold; framework via replace → SUMI_PATH),
  `sumi dev` (hot reload, keep-last-good), `sumi generate .`,
  `sumi inspect tree|boxes` (live tree over the dev socket),
  `sumi lsp` (diagnostics/completion/hover in editors).
- **Testing**: `runtime/sumitest` scenario tests + vt100 assertions;
  `TestApp` renders deterministically (no animation stepping).

## Gotchas

- Reading a signal in script without `.Get()` won't compile; writing
  without `.Set/.Update` won't re-render.
- Whitespace collapses per CSS in templates (space-only rows need
  `white-space: pre`).
- Style attrs are CSS-owned: never expect runtime projections to
  override them — behavioural flags live on script/attrs instead.
- `go test ./pkg -flags`: package path BEFORE custom flags.

## Working in the repo

TDD (red → green → refactor → commit). Read
`~/projects/sumi/PLAN-svelterm-parity.md` first on any parity work —
it is the single source of truth for progress. Keep files ≤~200
lines; test files mirror source splits; Given/When/Then comments.
