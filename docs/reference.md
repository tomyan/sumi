# Reference

The authoritative support matrix. Three buckets throughout: **✓**
supported, **~** supported with stated deviations, **✗** not supported
(dropped gracefully unless noted). Anything not listed is unsupported.

## Elements

| Element | Status | Notes |
|---|---|---|
| `div`, `span`, `p`, `h1`–`h6` | ✓ | block/inline per the UA stylesheet |
| `ul`, `ol`, `li` | ~ | `•` markers via `li::before`; no `ol` counters yet |
| `blockquote`, `pre`, `code`, `hr` | ✓ | |
| `strong/b`, `em/i`, `u`, `s/del`, `mark`, `kbd`, `abbr`, `samp`, `var` | ✓ | SGR-attribute styling |
| `a` | ✓ | underlined, focusable; Enter/click opens `href` |
| `button` | ✓ | Enter/Space/click activate |
| `input` (text) | ✓ | readline keymap, `value`, `maxlength`, `readonly`, `password` |
| `input` checkbox / radio | ✓ | `[x]` / `(•)`, Space toggles, radio `name` groups |
| `select` / `option` / `optgroup` | ~ | popup-less cycling control (no dropdown overlay) |
| `textarea` | ✓ | multi-line editing |
| `progress`, `meter` | ✓ | eighth-block bars |
| `details` / `summary` | ✓ | ▶/▼, `open`, `toggle` event |
| `dialog` | ✓ | modal focus trap, Escape closes |
| `label` | ✓ | wrapping and `for=` association |
| `region` | ✓ | consumer-fed cell area + `resize` event |
| `ansi` | ✓ | raw SGR passthrough, `pre`-like |
| `img` | ✗ | planned (half-block cells, then kitty graphics) |
| `table`, `tr`, `td`, `th`, `thead/tbody/tfoot`, `caption`, `colgroup` | ✓ | see Layout |

## CSS: layout

| Feature | Status | Notes |
|---|---|---|
| `display: block / inline / inline-block` | ✓ | real inline formatting context: runs wrap across styled elements |
| `display: flex` | ✓ | direction, wrap, gap, grow/shrink/basis, justify (incl. space-around/evenly), align, align-self, order, reverse |
| `display: grid` | ~ | `cell`/`ch`/`%`/`fr`, `repeat()`, `minmax()`, areas, span; **row auto-flow only** |
| `display: table` + internals | ✓ | colspan/rowspan, caption, header/footer groups, `border-collapse`, `border-spacing`, `table-layout: fixed`, `empty-cells`, colgroup width hints |
| `display: contents / none` | ✓ | |
| width/height, min/max-\*, `%`, `calc()` | ✓ | `box-sizing` both models |
| padding, margin | ✓ | margin `auto` centring; vertical margin collapse between adjacent block siblings (positive margins only) |
| `position: relative / absolute / fixed / sticky` | ✓ | `z-index` paint order |
| `overflow: hidden / scroll / auto` | ✓ | scrollbars, mouse wheel, scroll state |
| Units | ~ | `cell` and `ch` (1 cell) only; pixel-derived units drop to 0 |

## CSS: selectors and cascade

| Feature | Status | Notes |
|---|---|---|
| tag / class / id / attribute selectors, combinators | ✓ | |
| structural pseudo-classes (`:first-child`, `:nth-child()`, …) | ✓ | true runtime sibling context, including `{for}` lists |
| `:hover`, `:focus`, `:checked`, `:disabled`, `:enabled` | ✓ | |
| `:not()` / `:is()` | ~ | compound-only arguments |
| `::before` / `::after` | ~ | `content` strings, `attr()`, concatenation; subject-only |
| specificity + source order + inline-attribute precedence | ✓ | inline template attributes always win |
| `var()` custom properties | ✓ | inheritance and fallbacks |
| `calc()` / `min()` / `max()` / `clamp()` | ✓ | `%` resolves at layout |
| `@media` | ✓ | viewport, `prefers-color-scheme`, `prefers-reduced-motion` |
| `@container` | ✓ | size queries against laid-out ancestors |
| `@supports` | ~ | property-name checks |

## CSS: visual

| Feature | Status | Notes |
|---|---|---|
| colours | ✓ | named, hex, `rgb()`, `light-dark()`; Lab-space interpolation; quantized to detected depth (16/256/truecolor) |
| `opacity` / alpha | ~ | composites over RGB backdrops; dim fallback otherwise |
| borders | ✓ | single/double/rounded/heavy/ascii box drawing; `border-collapse` junction merging; `border-title` (sumi extra) |
| `text-align`, `text-overflow` (`ellipsis`, `ellipsis-middle`) | ✓ | |
| `white-space: normal / nowrap / pre`, `text-transform`, `word-break`, `visibility` | ✓ | |
| transitions | ✓ | colours + lengths, easing incl. `steps()` |
| `@keyframes` animations | ✓ | iteration/direction/fill-mode/play-state; per-node `var()` and `light-dark()` |

## Events and input

| Feature | Status | Notes |
|---|---|---|
| capture/bubble DOM dispatch, `stopPropagation`, `preventDefault` | ✓ | `click`, `keydown`, `input`, `change`, `paste`, `toggle`, `focus`, `blur`, `resize` |
| focus management | ✓ | Tab/Shift-Tab cycling, click-to-focus, dialog focus trap, `disabled` skipped |
| keyboard | ✓ | full modifier decoding, F1–F12, kitty protocol (disambiguate mode: Ctrl+Enter, clean Escape) with legacy fallback |
| mouse | ✓ | click, wheel scroll, hover, drag |
| global text selection | ✓ | drag / double-click word / triple-click line; inverse painting; OSC 52 + platform-tool clipboard on release |
| bracketed paste | ✓ | |
| Ctrl+Z suspend / resume | ✓ | clean terminal restore + repaint |

## Runtime

| Feature | Status | Notes |
|---|---|---|
| fullscreen mode | ✓ | alternate screen, synchronized output (DEC 2026), diffed frames |
| inline mode | ✓ | renders at the shell cursor; final frame stays in scrollback; FrameLog streams and archives frames |
| io injection | ✓ | `RunOptions.In/Out/Size`; runs in the browser via `runtime/webterm` (wasm) |
| colour scheme | ✓ | OSC 11 detection, `light-dark()`, forced via option |
| reduced motion | ✓ | option or `SUMI_REDUCED_MOTION` |
| log capture | ✓ | `RunOptions.OnLog` routes stdlib `log` |
| tooling | ✓ | `sumi init`, `sumi dev` (hot reload, keep-last-good), `sumi inspect` |

## Templates

| Feature | Status | Notes |
|---|---|---|
| `{expression}` text, `name={expr}` / `{name}` attributes | ✓ | |
| mixed content (text interleaved with elements) | ✓ | whitespace-only gaps follow the JSX newline rule |
| `{if}/{else}/{/if}`, `{for}` with `key=` diffing | ✓ | |
| snippets and `{render}` | ✓ | |
| components with props, callback props, `bind:value`, slots | ✓ | |
| `{#await}` / `{#key}` equivalents | ✗ | |

Deviations from web behaviour are listed in each chapter. The terminal
capability matrix (which terminals support which escape sequences) is
in [terminals](terminals.md).
