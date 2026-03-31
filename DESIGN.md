# Sumi Design

A declarative TTY framework for Go. Inspired by Ink (terminal UI), Solid.js (runtime fine-grained reactivity), and Svelte (single-file components).

## Overview

Sumi lets you build terminal user interfaces using `.sumi` single-file components that compile to Go source code. It combines:

- **Solid-style runtime signals** — fine-grained reactivity via a Go-native API (`signal.New()`, `signal.From()`, `signal.Effect()`), composable by default, no compiler magic for reactivity
- **Ink's idea** — declarative, component-based terminal UIs — but without React's virtual DOM or its rendering bugs
- **A curated subset of CSS** — scoped styling, responsive design via media queries, adapted for the terminal
- **Marketplace-ready** — reactive utilities and components are just Go packages, importable with `go get`

## Architecture

```
.sumi files → sumi compiler → .go files → go build → binary
```

The sumi compiler parses `.sumi` files and generates Go source code. It integrates into the Go toolchain via `go generate`. The generated code uses a sumi runtime library for rendering, layout, and the signals runtime.

The **reactive signals runtime** (`runtime/signal/`) is a standalone Go library. It provides fine-grained dependency tracking at runtime — no compiler involvement. This means reactive logic works in plain `.go` files, not just `.sumi` files. Marketplace authors can publish reactive utilities as standard Go packages.

### Syntax Highlighting: Tree-sitter

Sumi uses **tree-sitter** for syntax highlighting and structural code awareness. Tree-sitter parses source code into a concrete syntax tree, enabling:

- **Syntax highlighting** — language-aware token classification for code blocks in markdown, diff rendering, and editor components
- **Incremental reparsing** — edit a character, reparse only the affected subtree (efficient for live editing)
- **Structural queries** — AST-level operations like smart selection, code folding, semantic scope identification

Tree-sitter grammars are available for 200+ languages. The Go bindings (`go-tree-sitter`) are proven in production — Sumi's sibling project (hubcap) already uses tree-sitter for JavaScript highlighting in a terminal DevTools interface.

This is preferred over regex-based highlighting (e.g. TextMate grammars / syntect) because:
- It stays in the Go ecosystem — no CGo/FFI complexity for a Rust dependency
- It enables structural features beyond highlighting (smart selection, folding, AST-aware diffs) that regex-based approaches cannot provide
- The integration cost is known and low

### Potential: Rust Core via FFI

If profiling reveals that the Go layout engine, screen diffing, or text measurement become bottlenecks at scale, these could be reimplemented in Rust and linked via CGo as a static library. This is a **performance escape hatch**, not a core architectural decision — the current all-Go implementation is the primary path.

Candidates for Rust acceleration if ever needed:
- Layout engine (flexbox constraint solving, runs every frame)
- Screen buffer diff (tight cell-comparison loop)
- Text measurement (unicode-width, grapheme clustering)
- Color quantization (perceptual distance calculations for truecolor→256→16 fallback)

## .sumi File Format

Each `.sumi` file is a single component with three optional sections:

```html
<script>
// Valid Go code with reactive primitives
</script>

<style>
/* Terminal CSS */
</style>

<!-- Template -->
<box class="container">
  <text class="title">Hello, {name}</text>
</box>
```

### Script Block

The `<script>` block contains component logic. It is **valid Go code** — the compiler parses it with `go/ast`. Props are `var` declarations. State is created with `signal.New()`. No `$state`/`$prop`/`$derived` runes.

```html
<script>
var label string = "Count"

count := signal.New(0)
doubled := signal.From(func() int { return count.Get() * 2 })

signal.Effect(func() {
    log.Println("count changed:", count.Get())
})

func increment() {
    count.Update(func(n int) int { return n + 1 })
}
</script>
```

**Signal API:**

| Function | Purpose | Example |
|----------|---------|---------|
| `signal.New[T](initial)` | Create reactive state | `count := signal.New(0)` |
| `signal.From[T](fn)` | Derived value, auto-tracks dependencies | `doubled := signal.From(func() int { return count.Get() * 2 })` |
| `signal.Effect(fn)` | Side effect, runs when dependencies change | `signal.Effect(func() { ... })` |
| `signal.Get()` | Read current value (tracks dependency) | `v := count.Get()` |
| `signal.Set(v)` | Write new value (notifies dependents) | `count.Set(42)` |
| `signal.Update(fn)` | Read-modify-write | `count.Update(func(n int) int { return n + 1 })` |

**Why Go-native, not compiler runes:**

The script block is valid Go. This means:
- `gopls` works — autocompletion, type checking, go-to-definition
- Reactive utilities can live in plain `.go` files, not just `.sumi` components
- Marketplace components are standard Go packages (`go get`)
- One mental model everywhere — same API in components, utilities, and libraries

**Reactivity rule: `.Set()` triggers updates.** The signals runtime tracks which `From`/`Effect` computations called `.Get()` on which signals, and re-runs them when those signals change via `.Set()`. No compiler involvement — dependency tracking is fully automatic at runtime.

```html
<script>
items := signal.New([]string{"a", "b"})

func addItem(s string) {
    items.Update(func(xs []string) []string {
        return append(xs, s)
    })
}
</script>
```

**Template expression sugar:** The template compiler automatically unwraps signals. `{count}` in a template generates `count.Get()` in the compiled code. This is the only compiler magic — and it only applies to templates, not Go code.

### Style Block

The `<style>` block uses CSS syntax, scoped to the component by default. It supports a curated subset of CSS designed for terminal UIs.

```html
<style>
.container {
    border: single;
    border-color: cyan;
    padding: 1 2;
    direction: row;
    justify: space-between;
    width: 100%;
}

.title {
    color: green;
    bold: true;
}

@media (width > 80) {
    .container {
        direction: row;
    }
}

@media (width <= 80) {
    .container {
        direction: column;
    }
}

@media (theme: dark) {
    .title {
        color: lime;
    }
}

@media (color-depth: truecolor) {
    .title {
        color: #00ff88;
    }
}
</style>
```

**What we keep from CSS:**
- Class selectors (`.class`), element selectors (`text`, `box`)
- Pseudo-classes: `:focus`, `:active`
- Custom properties: `--accent: cyan;`
- Media queries (adapted for terminal — see below)
- Container queries (`@container`)
- Flexbox layout properties
- Grid layout
- Box model: padding, margin, border (**border-box by default**)
- Cascade within component scope
- Transitions and animations (`transition`, `@keyframes`)
- `calc()`, `min()`, `max()`, `clamp()`

**What we ditch:**
- `!important`
- Content-box model
- Float, clear, display: table, all legacy layout
- Complex specificity rules — flat and predictable
- Global scope — everything is scoped to the component
- Font properties (terminal font is fixed)
- `transform: rotate/scale/skew` (no meaning in cell grid)
- `box-shadow` (no sub-cell rendering)

**Terminal-specific CSS:**
- `border-style: single | double | rounded | heavy | none`
- Colors: ANSI names (`red`, `cyan`), 256-color (`color-196`), hex for truecolor (`#ff0088`)
- Text: `bold`, `dim`, `italic`, `underline`, `strikethrough`, `inverse`, `blink`
- Scrollbar styling: custom characters for scrollbar track/thumb

**Media queries:**
- `@media (width > N)` / `@media (height > N)` — terminal dimensions
- `@media (color-depth: monochrome | ansi | 256color | truecolor)` — graceful color degradation
- `@media (theme: dark | light)` — terminal theme detection
- `@media (prefers-reduced-motion)` — skip animations
- `@media (prefers-contrast)` — high contrast mode

### Template

The template section is the component's markup. It uses HTML-like syntax with Go-flavored control flow. Control flow tags use `{if}`, `{for}`, `{slot}`, `{snippet}`, `{render}` — no `#` prefix.

**Expressions:**
```html
<text>Hello, {name}</text>
<text>Count: {count + 1}</text>
```

Expressions auto-unwrap signals in template context. Conditions and clauses in `{if}`/`{for}` are written with explicit `.Get()` by the user:

```html
{if count.Get() > 0}
    <text>Count: {count}</text>
{else}
    <text>No count yet</text>
{/if}
```

**Loops:**
```html
{for i, item := range items.Get()}
    <text>{i}: {item}</text>
{/for}
```

**Snippets and render:**
```html
{snippet renderItem(item string)}
    <text>{item}</text>
{/snippet}

{render renderItem("hello")}
```

**Built-in elements:**
- `<text>` — styled text content
- `<box>` — container with layout, border, padding

Higher-level components (inputs, lists, tables, etc.) are built as a separate component library on top of these primitives.

### Slots

Components define slot placeholders with `<slot:name />`. Consumers fill them with `{slot name}...{/slot}`. A slot can access its default content via `<slot:default />`.

```html
<!-- card.sumi -->
<box border="single">
    <slot:header />
    <box padding="1">
        <slot:content />
    </box>
</box>
```

```html
<!-- usage -->
<card>
    {slot header}
        <text bold="true">My Title</text>
    {/slot}
    {slot content}
        <text>Body text here</text>
    {/slot}
</card>
```

Scoped slots support typed parameters for passing data from the component back to the consumer.

## Component Model

Each `.sumi` file is one component, living in its own package. The filename determines the component name. The compiler generates a `NewFoo(FooProps) *tui.Component` constructor and a `FooProps` struct.

**Props** are `var` declarations in the script block:

```html
<!-- counter.sumi (in package counter/) -->
<script>
var label string = "Count"

count := signal.New(0)

func handleKey(evt input.Event) {
    if evt.Kind == input.EventKey {
        count.Update(func(n int) int { return n + 1 })
    }
}
</script>

<style>
.label { color: cyan; bold: true; }
.count { color: yellow; bold: true; }
</style>

<box onkey="handleKey">
    <text class="label">{label}:</text>
    <text class="count">{count}</text>
</box>
```

This generates:

```go
package counter

type CounterProps struct {
    Label string
}

func NewCounter(props CounterProps) *tui.Component {
    label := props.Label
    if label == "" {
        label = "Count"
    }
    count := signal.New(0)
    // ... layout tree, effects, event handlers ...
    return &tui.Component{
        Tree:    root,
        OnEvent: handleKey,
    }
}
```

**Using components from Go:**

```go
package main

import (
    "github.com/example/myapp/counter"
    "github.com/tomyan/sumi/runtime/tui"
)

func main() {
    tui.Run(counter.NewCounter(counter.CounterProps{Label: "Clicks"}))
}
```

**Env values** are framework-provided reactive signals for terminal state:

```html
<script>
width := tui.Env[int]("width")
height := tui.Env[int]("height")
</script>
```

These are reactive — components re-render when the terminal resizes or other environment values change. The framework updates them on `SIGWINCH`.

**bind:value** passes signal references from parent to child, allowing two-way data binding between components.

### Composable Reactive Utilities

Because the signal API is plain Go, reactive logic can live outside `.sumi` files:

```go
// Published as github.com/someone/sumi-pagination
package pagination

import "github.com/tomyan/sumi/runtime/signal"

type Pagination[T any] struct {
    Page    *signal.Signal[int]
    Visible *signal.Computed[[]T]
}

func New[T any](items *signal.Signal[[]T], pageSize int) *Pagination[T] {
    page := signal.New(0)
    visible := signal.From(func() []T {
        all := items.Get()
        start := page.Get() * pageSize
        return all[start:min(start+pageSize, len(all))]
    })
    return &Pagination[T]{Page: page, Visible: visible}
}
```

A `.sumi` component uses it like any Go import:

```html
<script>
items := signal.New([]string{"a", "b", "c", "d", "e"})
pager := pagination.New(items, 2)
</script>

<box>
    {for i, item := range pager.Visible.Get()}
        <text>{item}</text>
    {/for}
</box>
```

No special sumi compiler support needed. The pagination package is a standard Go module.

## Layout Engine

Pure Go implementation of flexbox-like layout:

1. Vertical and horizontal stacking (direction: row | column)
2. Width and height (fixed, percentage, auto)
3. Padding and margin
4. Border (single, double, rounded, heavy)
5. Justify and align
6. Gap
7. Flex grow/shrink
8. Wrap
9. Min/max sizing

The layout engine maps the component tree to a grid of terminal cells, assigning each component a screen region (row, col, width, height). Available width/height are threaded through `layoutNode(input, availW, availH)`.

## Rendering

### Cell-Addressed Updates

The renderer maintains a virtual screen buffer — a 2D grid of cells, where each cell holds a character and its style (color, bold, etc.). On each reactive update:

1. Signal changes propagate through the dependency graph (runtime)
2. Effects subscribed to changed signals trigger re-layout of affected subtrees
3. The renderer diffs the new buffer against the previous one
4. Only changed cells are written to the terminal via cursor-addressed escape sequences

The signals runtime enables **fine-grained updates**: when a single signal changes, only the template nodes that depend on it (via `.Get()`) need to re-render. This scales to large UIs — appending a line to a log doesn't re-layout the entire screen.

This avoids Ink's "clear N lines and rewrite everything" approach, which is the root cause of its scrollback bugs.

### Render Modes

**Alternate screen** (`\x1b[?1049h`):
- Takes over the full terminal
- No scrollback interaction
- Simplest and safest mode
- Full repaint on resize (positions shift, but no scrollback to corrupt)

**Inline mode:**
- Renders within the terminal's normal scrollback flow
- Tracks its own screen region
- On resize: recalculates layout, re-renders affected cells
- Handles terminal reflow carefully

Components can switch between modes. Both modes use the same cell-addressed rendering underneath.

### Resize Handling

Terminal resize (`SIGWINCH`) triggers:

1. Query new terminal dimensions
2. Update `tui.Env[int]("width")` and `tui.Env[int]("height")` (reactive — triggers dependent components)
3. Re-run layout for the full tree with new dimensions
4. CSS media queries re-evaluate
5. Diff and render changed cells

In alternate screen mode, this is a clean full repaint. In inline mode, the renderer re-anchors to its region before updating.

## Compiler

The `sumi` CLI tool compiles `.sumi` files to Go source:

```
sumi generate              # compile all .sumi files in current directory
sumi generate ./components # compile .sumi files in specific directory
```

For each `foo.sumi`, it produces `foo_sumi.go` in the same package. Intended to be used via `go generate`:

```go
//go:generate sumi generate
```

The compiler:
1. Parses the `.sumi` file into script, style, and template sections
2. Parses the script block with `go/ast` — identifies `var` declarations (props), `signal.New()` calls (state), and function declarations
3. Parses the style block as terminal CSS
4. Parses the template as a component tree (including `{if}`, `{for}`, `{slot}`, `{snippet}`, `{render}`)
5. Generates Go code that wires up the layout tree, signal effects, and event handlers

The compiler does **not** transform the script block's Go code. Its only role with signals is identifying which variables are `*signal.Signal` or `*signal.Computed` so that template expressions like `{count}` can generate `count.Get()`. All reactive dependency tracking happens at runtime via the signals library.

## Project Structure

```
sumi/
  cmd/sumi/          # compiler CLI
    preview/          # interactive preview tool (bridge, editors, rendering)
  runtime/            # runtime library used by generated code
    signal/           # reactive signals runtime (New, From, Effect)
    layout/           # flexbox layout engine (Go, or thin wrapper over Rust core)
    render/           # cell buffer, terminal output, screen modes
    css/              # terminal CSS parser and resolver
    tui/              # app lifecycle, event loop, terminal setup, Env signals
    input/            # keyboard/mouse input parsing, event encoding
    term/             # terminal size, capabilities
    vt100/            # VT100 terminal emulator (for embedded editors)
    pty/              # PTY wrapper (macOS /dev/ptmx)
    sumitest/         # control protocol for scenario testing (serve mode)
  parser/             # .sumi file parser
    script/           # script block parser (valid Go via go/ast, extracts signal info)
    style/            # style block parser (terminal CSS)
    template/         # template parser (markup + control flow + slots + snippets)
  codegen/            # Go code generator
  components/         # first-party component library
    sumi/             # built-in components (TextInput, SplitPanel, Button)
  examples/           # example applications
```

The `runtime/signal/` package is independently usable — it has no dependency on the rest of sumi. Third-party packages can import it directly to build reactive utilities.

## What Has Been Built

The following capabilities are implemented and working:

**Core rendering and layout:**
- Static text rendering with alternate screen mode
- Box layout with direction (row/column), padding, border (single/double/rounded/heavy), fixed size
- Flexbox: justify, align, gap, flex-grow, percentage widths
- Text wrapping within boxes
- Border-color on box-drawing characters
- Border-title (tmux-style) and border-collapse (shared borders with junction characters)
- Min-width constraints
- Display:none (removes from layout, nil placeholders in children)

**Reactivity and state:**
- Signal runtime: `signal.New()`, `signal.From()`, `signal.Effect()` with automatic dependency tracking
- Build-once tree with expression node extraction — tree built once, `sync()` patches mutable fields before re-layout
- Derived declarations (`signal.From`)
- Template auto-unwrapping of signals in expressions

**Styling:**
- CSS-like scoped style blocks with class and element selectors
- Colors (ANSI names), bold/dim/italic/underline/strikethrough/inverse
- Responsive design: `tui.Env[int]("width")`/`tui.Env[int]("height")` signals, `SIGWINCH` resize handling, dynamic title

**Components:**
- Single-file `.sumi` components with script/style/template sections
- Props as `var` declarations, generating `FooProps` struct and `NewFoo(FooProps) *tui.Component` constructor
- Component composition via Go imports
- `bind:value` for two-way data binding (passes signal references)
- Callback props (function-typed props)
- Slots (`<slot:name />` placeholders, `{slot name}...{/slot}` definitions)
- Snippets (`{snippet name(params)}...{/snippet}`) and render (`{render name(args)}`)

**Template control flow:**
- `{if}`/`{else}`/`{/if}` conditionals
- `{for}`/`{/for}` loops with keyed diffing (`key=expr` syntax, identity-based diff matching)
- IIFE codegen pattern for dynamic children

**Positioning and layering:**
- `position: relative`, `absolute`, `fixed`, `sticky`
- `z-index` with z-aware paint order and hit testing
- Top/left/right/bottom offsets

**Scrolling and overflow:**
- `overflow: scroll`, `auto`, `hidden` with clipping
- Scroll state tracking, keyboard and mouse scroll input
- Animated scrollbar with click, ease-out-cubic, drag
- `$scroll` for scroll state access

**Input and interaction:**
- Keyboard input: basic keys, extended special keys (Enter, Escape, Backspace, Delete, Ctrl+key)
- Mouse events: click, scroll wheel, hit-test against layout tree
- Focus management: Tab/Shift-Tab cycling, `stopPropagation()`, focus dispatch
- Cursor rendering primitive: CursorCol/CursorRow, FindCursor, ShowCursor/HideCursor
- Event-aware handlers with parameterized closures
- `app.Quit()` via quit channel, signal dispatch as EventSignal

**Built-in components:**
- TextInput: bind:value, cursor, view offset, password/maxlength/readonly, text selection, copy/cut (OSC 52 clipboard), double-click word select, bracketed paste, strip on blur, scroll indicator, animated scrollbar
- SplitPanel: side-by-side panels with border-collapse
- Button: simple labeled button

**Animation:**
- `EventFrame` + `app.RequestFrame()` for frame-based animation
- Ease-out-cubic timing for scrollbar animation

**Developer tooling:**
- `sumi generate` CLI with `go generate` integration
- Interactive preview tool with VT100 parser, PTY wrapper, embedded nvim editors
- Scenario testing: serve mode with Unix socket control protocol (info/step/quit)
- File watcher for hot-reload during development

## CSS Feature Roadmap

Comprehensive mapping of CSS features to terminal UI, organized by priority tier.

### Tier 1 — Core (essential for real applications)

**Box Model:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `padding` | Cell insets | Done |
| `margin` | Cell spacing between elements | Planned |
| `border` | Box-drawing characters | Done |
| `border-style` | `single \| double \| rounded \| heavy \| none` | Done |
| `border-color` | ANSI color on border chars | Done |
| `border-title` | Tmux-style title in border | Done |
| `border-collapse` | Shared borders with junction characters | Done |
| `width` / `height` | Fixed cell counts | Done |
| `min-width` / `min-height` | Minimum cell counts | Done (min-width) |
| `max-width` / `max-height` | Maximum cell counts | Planned |
| `box-sizing` | Always `border-box` | Done |

**Layout:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `direction: column` | Vertical stacking | Done |
| `direction: row` | Horizontal stacking | Done |
| `justify-content` | `start \| end \| center \| space-between \| space-around \| space-evenly` | Done |
| `align-items` | `start \| end \| center \| stretch` | Done |
| `align-self` | Per-child override | Planned |
| `gap` | Cell spacing between children | Done |
| `flex-grow` / `flex-shrink` | Distribute extra space | Done (flex-grow) |
| `flex-basis` | Initial size before grow/shrink | Planned |
| `flex-wrap` | Wrap children to next line | Planned |

**Display & Visibility:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `display: none` | Remove from layout entirely | Done |
| `display: flex` | Default for `<box>` | Implicit |
| `display: grid` | Grid layout mode | Planned |
| `display: contents` | Unwrap container | Planned |
| `visibility: hidden` | Takes space but renders blank | Planned |

**Overflow:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `overflow: hidden` | Clip at box boundary | Done |
| `overflow: scroll` | Scrollable region | Done |
| `overflow: auto` | Scroll only if needed | Done |
| `overflow-x` / `overflow-y` | Per-axis control | Planned |
| `text-overflow: ellipsis` | Truncate with `...` | Planned |

**Text:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `color` | Foreground ANSI color | Done |
| `background` | Background ANSI color | Done |
| `bold` / `dim` / `italic` / `underline` / `strikethrough` / `inverse` | SGR attributes | Done |
| `text-align` | `left \| center \| right` within box | Planned |
| `white-space` | `nowrap \| pre-wrap` | Done (text wrapping) |

**Pseudo-classes:**
| Selector | Terminal Mapping | Status |
|---|---|---|
| `:focus` | Currently focused element | Done |
| `:active` | Being activated / pressed | Planned |

**Custom properties:**
| Feature | Terminal Mapping | Status |
|---|---|---|
| `--custom: value` | CSS variables | Planned |
| `var(--custom)` | Variable reference | Planned |
| `var(--custom, fallback)` | With fallback | Planned |

### Tier 2 — Power features (real framework feel)

**Colors:**
| Property | Terminal Mapping | Status |
|---|---|---|
| 8 basic colors | `black red green yellow blue magenta cyan white` | Done |
| 16 colors (bright) | `bright-red`, `bright-cyan`, etc. | Planned |
| 256-color | `color-196` syntax | Planned |
| Truecolor (24-bit) | `#ff0088` hex values | Planned |
| `opacity` | Layer blending | Planned |
| `background: transparent` | Show content below | Planned |

**Positioning:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `position: relative` | Positioning context | Done |
| `position: absolute` | Position relative to ancestor | Done |
| `position: fixed` | Position relative to screen | Done |
| `position: sticky` | Stick to edge on scroll | Done |
| `top` / `right` / `bottom` / `left` | Cell offsets | Done |
| `z-index` | Paint order / layer stack | Done |

**Transitions & Animations:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `transition` | Animate property changes over time | Planned |
| `transition-duration` | Timing in ms | Planned |
| `transition-timing-function` | `ease \| linear \| ease-in-out` | Planned |
| `@keyframes` | Multi-step animations | Planned |
| `animation` | Shorthand for keyframe animations | Planned |
| `animation-iteration-count` | `infinite` for loops | Planned |
| `animation-play-state` | `paused \| running` | Planned |

Note: Frame-based animation is supported via `app.RequestFrame()` and `EventFrame` dispatch at the runtime level, but CSS transition/animation syntax is not yet implemented.

**Grid:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `grid-template-columns` / `rows` | Cell-based tracks | Planned |
| `grid-gap` | Cell spacing between tracks | Planned |
| `grid-area` | Named grid regions | Planned |
| `grid-auto-flow` | Auto-placement | Planned |

**Selectors:**
| Selector | Terminal Mapping | Status |
|---|---|---|
| `.class` | Class selector | Done |
| `element` | Element type selector | Done |
| `.parent .child` | Descendant combinator | Planned |
| `.parent > .child` | Direct child | Planned |
| `:first-child` / `:last-child` | Positional | Planned |
| `:nth-child()` | Nth element | Planned |
| `:not()` | Negation | Planned |

**Media & Container Queries:**
| Feature | Terminal Mapping | Status |
|---|---|---|
| `@media (width > N)` | Terminal width | Done |
| `@media (height > N)` | Terminal height | Done |
| `@media (color-depth)` | Capability detection | Planned |
| `@media (theme)` | Dark/light detection | Planned |
| `@container` | Component-relative queries | Planned |
| `@media (prefers-reduced-motion)` | Skip animations | Planned |

**Functions:**
| Function | Terminal Mapping | Status |
|---|---|---|
| `calc()` | `calc(100% - 10)` → cell math | Planned |
| `min()` / `max()` | Size bounds | Planned |
| `clamp()` | `clamp(10, 50%, 40)` | Planned |

### Tier 3 — Polish

**Text refinements:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `text-transform` | `uppercase \| lowercase \| capitalize` | Planned |
| `word-break` | Break long words | Planned |
| `line-height` | Extra rows between lines | Planned |
| `letter-spacing` | Extra cells between characters | Planned |
| `text-indent` | Indent first line | Planned |
| `blink` | SGR 5 (most terminals support) | Planned |

**Layout refinements:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `order` | Reorder children visually | Planned |
| `aspect-ratio` | Approximate in cells | Planned |
| `outline` | 1 cell outside border | Planned |

**Interaction:**
| Feature | Terminal Mapping | Status |
|---|---|---|
| Mouse events | Click, scroll, hit-test | Done |
| `:hover` | Mouse hover detection | Planned |
| `cursor` | Terminal cursor style | Done (CursorCol/CursorRow rendering) |
| `pointer-events` | Mouse event handling | Planned |
| `user-select` | Copyable text regions | Planned |
| Focus management | Tab cycling, stopPropagation | Done |
| Scrollbar | Custom animated scrollbar | Done |

**Compositing:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `backdrop-filter: dim` | Dim content behind element | Planned |
| `clip-path` | Rectangular clipping | Planned |
| `contain` | Layout/paint containment (perf) | Planned |

**Scrolling refinements:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `scroll-behavior: smooth` | Animated scrolling | Done (ease-out-cubic) |
| Scrollbar styling | Custom track/thumb characters | Done |

### Explicitly excluded

These CSS features have no meaningful mapping to terminal cells:

| Feature | Reason |
|---|---|
| `font-family` / `font-size` / `font-weight` (numeric) | Terminal font is fixed |
| `transform: rotate/scale/skew` | No sub-cell rendering |
| `box-shadow` | No sub-cell rendering |
| `border-radius` (pixel-level) | Covered by `border-style: rounded` |
| `float` / `clear` | Legacy layout, use flexbox/grid |
| `display: table` | Legacy layout |
| `!important` | Keeps cascade simple |
| Content-box model | Border-box only |
| Complex specificity | Flat and predictable |
| Global scope | Everything scoped to component |
| `touch-action` | No touch in terminal |
| `scroll-snap` | Overkill for terminal |
