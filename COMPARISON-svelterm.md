# Sumi vs Svelte + svelterm — Feature Comparison

Date: 2026-07-05. Sumi side verified against implementation code (not DESIGN.md claims).
Svelterm side taken from `~/projects/svelterm/docs/reference.md` (authoritative matrix),
the chapter docs, `~/projects/svelterm-ui`, and `~/projects/svelterm-site`.

Both frameworks share a thesis: single-file components, real CSS, terminal cell grid,
LLM-native authoring. Svelterm gets Svelte's language and a browser-grade CSS engine
for free; sumi trades that for Go, compile-time flattening, and a single static binary.

## 1. Authoring model and reactivity

| Area | Svelte + svelterm | Sumi | Gap |
|---|---|---|---|
| Component format | `.svelte` SFC (script/style/template) | `.sumi` SFC (script/style/template) | parity |
| Reactivity | Svelte 5 runes: `$state`, `$derived`, `$effect`, `$props`, `$bindable` | Solid-style signals: `sumi.New`, `sumi.From`, `signal.Effect`, `signal.Batch` | conceptual parity |
| Conditionals | `{#if}` / `{:else if}` / `{:else}` | `{if}` / `{else}` — **no else-if** | small gap |
| Loops | `{#each}` with keys | `{for … range}` with `key=` | parity |
| Await / key blocks | `{#await}`, `{#key}` | none | gap (await is N/A-ish in sync Go event loop; key useful) |
| Snippets/slots | snippets + `{@render}` | `{slot}`, `{snippet}`/`{render}` parsed; codegen thin; **mixed slot + non-slot children is a TODO** | gap |
| Two-way binding | `bind:value`, `bind:this` unavailable (use `{@attach}`) | `bind:value` / `bind:<name>` | near parity |
| Component events | callback props (Svelte 5 style) | callback props | parity |
| Context/stores | Svelte context + stores | none (package-level signals fill the role) | acceptable difference |
| Self-closing tags | standard | `<box/>` **not** supported; components can self-close | paper cut |
| Compilation | custom renderer fork of Svelte compiler; runtime tree of TermNodes | `.sumi` → Go source; components fully inlined at compile time | different by design |

## 2. Element vocabulary

The single biggest gap. Svelterm renders the standard HTML element set with UA styles;
sumi has exactly `<text>`, `<box>`, `<title>`, `<slot:*>`, and component tags.

Svelterm supports: headings, `p`, `div`, `span`, lists (markers), `blockquote`, `pre`,
`code`, `hr`; inline text-level `strong/em/u/s/mark/kbd/abbr/samp/var`; `a` (focusable,
opens href); full `table` layout (colspan/rowspan, header groups, caption, colgroup);
form controls: `input` (text/checkbox/radio), `textarea`, `button`, `select/option/
optgroup`, `progress`, `meter`, `details/summary`, `label`, `dialog` (modal focus
trap, Escape); `img` (half-block cells + kitty graphics on capable terminals);
`<svt-ansi>` (raw ANSI passthrough); `<svt-region>` (consumer-fed cell area).

Sumi equivalents that exist: `border-title` boxes, `focusable`/`onkey` boxes,
`contenteditable` boxes, and library components (Button, TextInput/textedit,
scrollbar, SplitPanel) — i.e. behaviour is available but as bespoke attributes and
components, not as standard elements with UA semantics.

## 3. CSS: selectors

| Feature | svelterm | sumi |
|---|---|---|
| Type, `.class`, `#id`, `*`, lists | ✓ | `.class` + bare `text`/`box` only |
| Combinators (descendant, `>`, `+`, `~`) | ✓ | ✗ |
| Attribute selectors (all 7 operators) | ✓ | ✗ |
| Structural pseudo-classes (`:first-child`, full An+B `:nth-*`, `:empty`, …) | ✓ | ✗ |
| State (`:focus`, `:hover`, `:checked`, `:disabled`, `:enabled`) | ✓ | `:hover` partial (parsed + resolver, design in DESIGN-hover.md); `:focus`/`:active` parsed but unconsumed |
| Logical (`:not()`, `:is()`, `:where()`) | ✓ | ✗ |
| `::before` / `::after` with `content` | ✓ (not in table internals) | ✗ |
| Specificity + cascade + inline precedence | ✓ per spec | last-rule-wins-ish; inline attrs override CSS |

## 4. CSS: properties

| Group | svelterm | sumi |
|---|---|---|
| Colour & text | `color`, `background(-color)`, `font-weight` (≥700 bold), `font-style`, `text-decoration`, `text-transform`, `text-align`, `text-overflow` (+`ellipsis-middle`), `white-space`, `word-break`, `opacity`≈dim, `visibility` | `color`, `background`, `border-color`; **non-standard booleans** `bold/dim/italic/underline/strikethrough/inverse: true`. No text-align/overflow/transform/white-space/visibility |
| Box model | width/height, min/max-*, padding, **margin** (auto centring, collapse), box-sizing, overflow incl. fading scrollbars | width/height (int only), `min-width`, padding, overflow + scrollbars. **No margin, no max-*, no min-height, no %, no box-sizing** |
| Display | block, inline, inline-block, flex, grid, none, contents, table types | flex (default) and `none` only |
| Flexbox | all 4 directions, wrap, grow/shrink/basis, gap, full justify-content incl. `space-*`, align-items, align-self, order | row/column, gap, flex-grow, justify (start/end/center/space-between), align (start/end/center/stretch). **No wrap/shrink/basis/align-self/order/reverse/space-around/evenly** |
| Grid | template cols/rows (`fr`, `repeat()`, `minmax()`), areas, span placement | ✗ |
| Tables | full CSS table layout | ✗ |
| Positioning | static/absolute/fixed + z-index; **relative offsets NOT applied** (their gap) | relative/absolute/fixed/**sticky** with offsets applied, z-index, z-aware hit test — **sumi ahead** |
| Borders | `single/double/rounded/heavy/ascii/eighth-*/half-*/full-cell`, per-side toggles, `border-corner` | parses several styles but **renders single-line only**; per-side top/bottom; **border-title and box border-collapse (tmux junctions) — sumi ahead** |
| Animation | `@keyframes`, animation longhands, iteration/infinite, timing funcs | ✓ comparable: @keyframes, shorthand+longhands, direction/fill-mode, cubic-bezier (no `steps()`; `play-state` ignored) |
| Transitions | shorthand, property list or `all`, duration, timing | ✓ but only `color`/`background`/`all` interpolate |

## 5. CSS: values, units, at-rules

| Feature | svelterm | sumi |
|---|---|---|
| Units | `cell`/`ch`, `%`, unitless 0, `fr`; px/em/rem parsed-and-dropped | bare ints only |
| Custom properties `var()` | ✓ inheritance + fallbacks | ✗ |
| `calc()`/`min()`/`max()`/`clamp()` | ✓ | ✗ |
| Colours | hex 3/6/8, rgb/hsl/hwb/lab/lch/oklab/oklch, 148 named, `transparent`, `currentColor`, `light-dark()`; alpha composites at paint time | 8 ANSI names + `#rrggbb` only |
| Colour depth | detect + degrade truecolor→256→16→mono, `NO_COLOR`; scheme via OSC 11 → `prefers-color-scheme` | truecolor + named out; **no detection, no degradation, no scheme detection** |
| `@media` | display-mode, prefers-color-scheme, min/max-width/height (cells) | **✗ parse error**; responsive = `tui.Env` signals + `{if}` |
| `@container`, `@supports` | ✓ | ✗ |
| `inherit`/`initial`/`unset` | ✓ | ✗ |
| Graceful-drop rule | anything pixel-derived parses and drops silently | unknown properties are not handled by a stated policy |

## 6. Events, input, focus

| Feature | svelterm | sumi |
|---|---|---|
| Event model | W3C capture/bubble, `stopPropagation`/`preventDefault`, default actions, payloads on `event.data` | `onkey`/`onclick` handlers, focus dispatch + `stopPropagation()`; no capture phase, no generic event types (`input`/`change`/`toggle`/`paste` as events) |
| Focus | Tab/Shift+Tab over control types, disabled skipped, `:focus` styling, click focuses | `focusable="true"` + Tab cycling, blur/focus dispatch; no `:focus` styling |
| Keyboard | Enter/Space activation, kitty keyboard protocol, Ctrl+Z suspend/restore | arrows/nav/Enter/Esc/Backspace/Del/Tab, Ctrl+A–Z, Alt; **no F1–F12, no kitty protocol, no suspend** |
| Mouse | click/wheel/hover, `:hover`, cell coords | SGR mouse press/release/motion/scroll, scrollbar drag, shift+scroll horizontal; hover partial |
| Selection & clipboard | drag/double/triple selection painted, copies via OSC 52 + platform tools | selection inside textedit component only; OSC 52 write exists |
| Paste | bracketed paste → `paste` event | ✓ bracketed paste → EventPaste |

## 7. Runtime, API, terminal handling

| Feature | svelterm | sumi |
|---|---|---|
| Entry API | `run(component, {css, fullscreen, mouse, props, colorScheme, io, onConsole, mode, exitOn, colorDepth, debug})` | `tui.Run` / `RunWithOptions{OnPostRender, OnResize, SetApp}`; App.Quit/Wake/Do/RequestFrame |
| Inline (non-fullscreen) mode | ✓ + FrameLog streaming into scrollback | ✗ fullscreen only |
| Log capture | `onConsole` (console.log throws otherwise) | ✗ |
| Headless testing | `@svelterm/core/headless` cell buffers | ✓ TestApp + TestClock + vt100 + sumitest socket protocol — parity, arguably ahead on determinism |
| Synchronized output (DEC 2026) | ✓ | ✗ |
| Terminal matrix / CI | documented matrix, ANSI round-trip CI tests | ✗ |
| Debug tooling | `svt` CLI (tree/query/style/box/console) over WebSocket + terminal DevTools app | test-preview TUI (VT100 + PTY + embedded nvim editors + file watcher) — different strengths |
| Distribution | Node runtime; curl-pipe demos, npm, bundles | **single static Go binary — sumi ahead** |

## 8. CLI / DX

| | svelterm | sumi |
|---|---|---|
| Commands | `init`, `dev`, `build`, `inspect`, `devtools` | `generate`, `test-preview` |
| Watch/dev loop | ✓ | design only (`sumi dev` in design-ui-support.md, unbuilt) |
| Editor tooling | — | nvim ftdetect/syntax/tree-sitter for `.sumi` — sumi ahead |

## 9. Component library

@svelterm/ui (11): Dialog, List, Tabs, FuzzyPicker, Toaster (+`toast()` queue),
ColorSwatch/ColorPalette/ColorSlider/ColorPanel/ColorInput/ColorPicker; plus fuzzy
matcher and a full OKLCH/OKLab colour engine as plain TS.

sumi components (6): Button, TextInput (deep: password/maxlength/readonly/selection/
undo/kill-buffer — **deeper than anything in @svelterm/ui**), textedit, scrollbar
(animated, draggable), SplitPanel, placeholder.

Missing vs @svelterm/ui: Dialog/modal, List, Tabs, FuzzyPicker, Toaster, colour suite.

## 10. Docs and site

svelterm: svelterm.dev — playground-first landing (live editor, dual terminal/browser
previews via xterm.js + untrusted-origin iframe), 12 docs chapters authored in the
core repo and rendered by the site, one authoritative reference matrix with MDN links,
distinctive terminal-hardware visual identity, honest known-gaps voice.

sumi: no site, no rendered docs. DESIGN.md is stale in places (claims `@media` and all
justify values done; both false). Blog has a stale draft `introducing-sumi` post
(pre-signals API) and a draft project page on tomyandell.dev.

## 11. Where sumi is ahead

- Single static binary; no Node, no compiler fork to maintain.
- Compile-time component flattening — zero runtime component overhead.
- `border-title` and inter-box `border-collapse` with junction characters (tmux-style
  panels) — svelterm has neither.
- `position: sticky`, and relative offsets actually applied (svelterm gap).
- TextInput/textedit depth (undo, selection, kill buffer, word ops).
- Deterministic testing story (TestClock, scenario socket, vt100, PTY all in-repo).
- test-preview with embedded nvim editors.
- Written multi-target (WebView) design for mobile/desktop/web.

## 12. Ranked gap summary (largest first)

1. **CSS engine**: selectors/cascade/specificity, standard property names, units
   (`cell`/`ch`/`%`), `var()`/`calc()`, colour level 4 + degradation, `@media`/
   `@container`/`@supports`, `::before`/`::after`, graceful-drop policy.
2. **HTML element vocabulary + UA styles**, including all form controls, `a`,
   `details`, `dialog`, `img`, tables.
3. **Layout**: margin, min/max/%, flex wrap/shrink/basis/align-self/order, block/
   inline flow, grid, tables.
4. **Event model**: capture/bubble, standard event types, `event.data` payloads,
   state pseudo-classes wired to input.
5. **Terminal capability handling**: colour depth detection/degradation, OSC 11
   scheme, kitty keyboard, DEC 2026, images, suspend.
6. **Rendering**: border styles beyond single, alpha compositing, opacity≈dim.
7. **Runtime**: inline mode + frame log, run options, log capture.
8. **Tooling**: `sumi dev`, `init`, inspector.
9. **Component library**: Dialog, List, Tabs, FuzzyPicker, Toaster, colour suite.
10. **Site + docs + announcement**: everything.
