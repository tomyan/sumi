# Sumi Design

A declarative TTY framework for Go. Inspired by Ink (terminal UI) and Svelte (reactivity model).

## Overview

Sumi lets you build terminal user interfaces using `.sumi` single-file components that compile to Go source code. It combines:

- **Svelte's approach to reactivity** — explicit reactive primitives (`$state`, `$derived`, `$effect`), compile-time transformation, fine-grained updates
- **Ink's idea** — declarative, component-based terminal UIs — but without React's virtual DOM or its rendering bugs
- **A curated subset of CSS** — scoped styling, responsive design via media queries, adapted for the terminal

## Architecture

```
.sumi files → sumi compiler → .go files → go build → binary
```

The sumi compiler parses `.sumi` files and generates Go source code. It integrates into the Go toolchain via `go generate`. The generated code uses a sumi runtime library for rendering, layout, and reactivity.

## .sumi File Format

Each `.sumi` file is a single component with three optional sections:

```html
<script>
// Go-like code with reactive primitives
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

The `<script>` block contains component logic. It uses Go syntax extended with reactive primitives (runes):

```html
<script>
name := $state("world")
upper := $derived(strings.ToUpper(name))

$effect(func() {
    log.Println("name changed:", name)
})

func reset() {
    name = "world"
}
</script>
```

**Runes:**

| Rune | Purpose | Example |
|------|---------|---------|
| `$state(initial)` | Declare reactive state | `count := $state(0)` |
| `$derived(expr)` | Computed value, updates when dependencies change | `doubled := $derived(count * 2)` |
| `$effect(fn)` | Side effect, runs when dependencies change | `$effect(func() { ... })` |
| `$prop(default)` | Component input from parent | `label := $prop("Click me")` |
| `$env(key)` | Reactive environment value (terminal width, height, theme) | `width := $env(width)` |

**Reactivity rule: reassignment triggers updates.** The compiler tracks variables declared with `$state` and `$prop`, and rewrites assignments to those variables to include invalidation. Mutation (e.g. struct field access) is not tracked — the compiler rejects untrackable mutations with a compile error.

```html
<script>
items := $state([]string{"a", "b"})

func addItem(s string) {
    items = append(items, s)  // reassignment — reactive, works
}
</script>
```

This keeps the model simple and predictable: if you want reactivity, reassign the variable.

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

The template section is the component's markup. It uses HTML-like syntax with Go-flavored control flow.

**Expressions:**
```html
<text>Hello, {name}</text>
<text>Count: {count + 1}</text>
```

**Conditionals:**
```html
{if count > 0}
    <text>Count: {count}</text>
{else}
    <text>No count yet</text>
{/if}
```

**Loops:**
```html
{for i, item := range items}
    <text>{i}: {item}</text>
{/for}
```

**Built-in elements:**
- `<text>` — styled text content
- `<box>` — container with layout, border, padding

Higher-level components (inputs, lists, tables, etc.) will be built as a separate component library on top of these primitives.

## Component Model

Each `.sumi` file is one component. Components compose naturally:

```html
<!-- counter.sumi -->
<script>
label := $prop("Count")
count := $state(0)

func increment() {
    count++
}
</script>

<style>
.label {
    bold: true;
    color: cyan;
}
</style>

<box direction="row" gap={1}>
    <text class="label">{label}:</text>
    <text>{count}</text>
</box>
```

```html
<!-- app.sumi -->
<script>
</script>

<box direction="column">
    <counter label="Clicks" />
    <counter label="Score" />
</box>
```

**Props** are declared with `$prop(default)`. Each prop is an independent reactive variable.

**Env values** are available via `$env()`:
```html
<script>
width := $env(width)
height := $env(height)
theme := $env(theme)
colorDepth := $env(colorDepth)
</script>
```

These are reactive — components re-render when the terminal resizes or theme changes.

## Layout Engine

Pure Go implementation of flexbox-like layout. Built iteratively, starting with:

1. Vertical and horizontal stacking (direction: row | column)
2. Width and height (fixed, percentage, auto)
3. Padding and margin
4. Border
5. Justify and align
6. Gap
7. Flex grow/shrink
8. Wrap
9. Min/max sizing

The layout engine maps the component tree to a grid of terminal cells, assigning each component a screen region (row, col, width, height).

## Rendering

### Cell-Addressed Updates

The renderer maintains a virtual screen buffer — a 2D grid of cells, where each cell holds a character and its style (color, bold, etc.). On each reactive update:

1. The reactivity system identifies which components are dirty
2. Those components re-layout and re-render into the buffer
3. The renderer diffs the new buffer against the previous one
4. Only changed cells are written to the terminal via cursor-addressed escape sequences

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
2. Update `$env(width)` and `$env(height)` (reactive — triggers dependent components)
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
2. Parses the script block as Go + runes, building a reactive dependency graph
3. Parses the style block as terminal CSS
4. Parses the template as a component tree
5. Generates Go code that wires up the reactivity, layout, and rendering

## Project Structure

```
sumi/
  cmd/sumi/          # compiler CLI
  runtime/            # runtime library used by generated code
    reactivity/       # $state, $derived, $effect implementation
    layout/           # flexbox layout engine
    render/           # cell buffer, terminal output, screen modes
    css/              # terminal CSS parser and resolver
  parser/             # .sumi file parser
    script/           # script block parser (Go + runes)
    style/            # style block parser (terminal CSS)
    template/         # template parser (markup + control flow)
  codegen/            # Go code generator
```

## Iteration Plan

Thin slices, each delivering something you can see working in the terminal.

### Iteration 1: Static text rendering
- Compiler parses a minimal `.sumi` file with just a `<text>` element (no script, no style)
- Generates Go code that renders static text to alternate screen
- `sumi generate` CLI exists and produces a working `.go` file
- **You see:** text on screen, program exits cleanly

### Iteration 2: Box layout basics
- `<box>` element with `direction="column"` (vertical stacking)
- Fixed `width`, `height`, `padding` attributes
- Border rendering (`border="single"`)
- **You see:** bordered boxes with text inside, stacked vertically

### Iteration 3: Reactive state
- `<script>` block parsing with `$state` rune
- Compiler rewrites assignments to trigger invalidation
- Template expressions `{variable}` bound to state
- Keyboard input (basic — read stdin)
- **You see:** a counter you can increment with a keypress, updating in place

### Iteration 4: Style block
- `<style>` block parsing (basic terminal CSS)
- Class selectors, colors, bold/dim/italic
- Scoped to component
- **You see:** styled, colored text and borders

### Iteration 5: Components
- Multiple `.sumi` files composing together
- `$prop` rune for component inputs
- **You see:** parent-child component composition working

### Iteration 6: Flexbox layout
- `direction: row`, `justify`, `align`, `gap`
- Percentage-based widths
- Flex grow/shrink
- **You see:** components laid out in flexible rows and columns

### Iteration 7: Responsive design
- `$env(width)`, `$env(height)` reactive environment
- CSS `@media` queries for terminal dimensions
- `SIGWINCH` handling, re-layout on resize
- **You see:** layout adapting as you resize the terminal

### Iteration 8: Derived state and effects
- `$derived` rune
- `$effect` rune
- **You see:** computed values updating automatically, side effects running

### Iteration 9: Inline rendering mode
- Render within terminal scrollback (not alternate screen)
- Cursor-addressed updates in inline region
- Resize handling in inline mode
- **You see:** sumi UI rendered inline, other terminal output above/below

### Iteration 10: Color depth and theme
- `@media (color-depth: ...)` queries
- `@media (theme: dark | light)` detection
- Graceful color degradation
- **You see:** UI adapting to terminal capabilities and theme

### Iteration 11: Layering and positioning
- `z-index` property — controls paint order for overlapping elements
- `position: absolute` — position relative to nearest positioned ancestor (or screen)
- `position: relative` — establishes positioning context for children
- `top`, `left`, `right`, `bottom` — offset from positioned parent
- Layout engine maintains a layer stack, paints back-to-front by z-index
- **You see:** modals, overlays, dropdowns rendered on top of other content

### Iteration 12: Compositing and transparency
- `opacity: 0.5` — semi-transparent elements (blends with content below)
- `background: transparent` — no background, content below shows through
- `background: dim` — dims content below (like a modal backdrop)
- Compositing model: when rendering a cell, blend foreground element's style with the cell already in the buffer
  - Transparent background → keep the character and bg color from below, apply fg styling on top
  - Dim → keep character from below, apply dim attribute
  - Opacity → interpolate colors between layers (truecolor), or use dim/bold approximation (ANSI)
- `backdrop-filter: dim | blur` — apply effect to content behind the element (dim is practical for terminals, blur is approximated)
- **You see:** modal dialogs with dimmed backgrounds, overlapping panels where content shows through

### Iteration 13: Overflow and scrolling
- `overflow: hidden` — clip content at box boundary
- `overflow: scroll` — scrollable region with keyboard navigation
- `overflow: auto` — scroll only when content exceeds bounds
- `overflow-x` / `overflow-y` — independent axis control
- Custom scrollbar rendering with box-drawing characters (`▓░▒█`)
- `text-overflow: ellipsis` — truncate text with `…`
- **You see:** scrollable lists, log panels, content that clips cleanly at borders

### Iteration 14: Text layout
- `text-align: left | center | right` — horizontal text alignment within box
- `white-space: nowrap | pre-wrap` — text wrapping control
- `word-break` — break long words at box edge
- `line-height` — extra rows between text lines
- `text-transform: uppercase | lowercase | capitalize`
- `letter-spacing` — extra cells between characters
- `text-indent` — indent first line
- **You see:** properly aligned, wrapped, and formatted text

### Iteration 15: Transitions and animations
- `transition: property duration timing-function` — animate property changes
  - Color transitions: fade between colors (truecolor interpolation, 256-color stepping)
  - Position transitions: slide elements across the screen
  - Size transitions: grow/shrink boxes
- `transition-timing-function: linear | ease | ease-in | ease-out | ease-in-out`
- `transition-delay`
- `@keyframes name { from { ... } to { ... } }` — multi-step animations
- `animation: name duration timing-function iteration-count direction`
  - Loading spinners: `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
  - Pulsing highlights, blinking cursors
  - Marquee text
  - `animation-iteration-count: infinite`
  - `animation-play-state: paused | running`
- Timer/tick system in the event loop for frame-based updates
- **You see:** smooth color changes, sliding panels, loading spinners

### Iteration 16: Grid layout
- `display: grid` on box elements
- `grid-template-columns` / `grid-template-rows` — define cell-based tracks
- `grid-gap` — cell spacing between tracks
- `grid-area` / named grid areas — place elements in named regions
- `grid-auto-flow: row | column` — auto-placement direction
- **You see:** dashboard-style layouts with named grid regions

### Iteration 17: Advanced selectors and cascade
- `.parent .child` — descendant combinator
- `.parent > .child` — direct child combinator
- `.a + .b` — adjacent sibling
- `:first-child` / `:last-child` / `:nth-child()`
- `:not()` — negation pseudo-class
- `[attr]` / `[attr=value]` — attribute selectors
- `:disabled` / `:enabled` — input state pseudo-classes
- `inherit` / `initial` / `revert` — cascade keywords
- `all: unset` — reset all properties
- **You see:** precise styling without extra class names

### Iteration 18: Custom properties and functions
- `--custom-prop: value` — CSS custom properties (variables)
- `var(--prop)` / `var(--prop, fallback)` — reference with optional fallback
- `calc(100% - 10)` — cell arithmetic in property values
- `min()` / `max()` / `clamp()` — size clamping functions
- Custom properties cascade through component tree
- **You see:** themeable components with shared design tokens

### Iteration 19: Container queries
- `container-type: size | inline-size` — mark element as query container
- `@container (width > N)` — style based on container size, not terminal size
- Enables truly reusable components that adapt to their available space
- **You see:** components that rearrange based on their parent, not the terminal

### Iteration 20: Mouse and interaction
- Mouse event support (click, hover, scroll wheel)
- `:hover` pseudo-class — highlight on mouse over
- `cursor` property — terminal cursor style (`block | bar | underline` via `\x1b[N q`)
- `pointer-events: none` — pass-through for overlays
- `user-select: none` — prevent text selection in interactive regions
- Click handlers: `onclick="handler"` attribute
- **You see:** clickable buttons, hover highlights, mouse-aware interfaces

## CSS Feature Roadmap

Comprehensive mapping of CSS features to terminal UI, organized by priority tier.

### Tier 1 — Core (essential for real applications)

**Box Model:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `padding` | Cell insets | Done |
| `margin` | Cell spacing between elements | Planned |
| `border` | Box-drawing characters | Done |
| `border-style` | `single \| double \| rounded \| heavy \| none` | Partial (single) |
| `border-color` | ANSI color on border chars | Planned |
| `width` / `height` | Fixed cell counts | Done |
| `min-width` / `min-height` | Minimum cell counts | Planned |
| `max-width` / `max-height` | Maximum cell counts | Planned |
| `box-sizing` | Always `border-box` | Done |

**Layout:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `direction: column` | Vertical stacking | Done |
| `direction: row` | Horizontal stacking | Planned (iter 6) |
| `justify-content` | `start \| end \| center \| space-between \| space-around \| space-evenly` | Planned (iter 6) |
| `align-items` | `start \| end \| center \| stretch` | Planned (iter 6) |
| `align-self` | Per-child override | Planned |
| `gap` | Cell spacing between children | Planned (iter 6) |
| `flex-grow` / `flex-shrink` | Distribute extra space | Planned (iter 6) |
| `flex-basis` | Initial size before grow/shrink | Planned |
| `flex-wrap` | Wrap children to next line | Planned |

**Display & Visibility:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `display: none` | Remove from layout entirely | Planned |
| `display: flex` | Default for `<box>` | Implicit |
| `display: grid` | Grid layout mode | Planned (iter 16) |
| `display: contents` | Unwrap container | Planned |
| `visibility: hidden` | Takes space but renders blank | Planned |

**Overflow:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `overflow: hidden` | Clip at box boundary | Planned (iter 13) |
| `overflow: scroll` | Scrollable region | Planned (iter 13) |
| `overflow: auto` | Scroll only if needed | Planned (iter 13) |
| `overflow-x` / `overflow-y` | Per-axis control | Planned |
| `text-overflow: ellipsis` | Truncate with `…` | Planned (iter 13) |

**Text:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `color` | Foreground ANSI color | Done |
| `background` | Background ANSI color | Done |
| `bold` / `dim` / `italic` / `underline` / `strikethrough` / `inverse` | SGR attributes | Done |
| `text-align` | `left \| center \| right` within box | Planned (iter 14) |
| `white-space` | `nowrap \| pre-wrap` | Planned (iter 14) |

**Pseudo-classes:**
| Selector | Terminal Mapping | Status |
|---|---|---|
| `:focus` | Currently focused element | Planned |
| `:active` | Being activated / pressed | Planned |

**Custom properties:**
| Feature | Terminal Mapping | Status |
|---|---|---|
| `--custom: value` | CSS variables | Planned (iter 18) |
| `var(--custom)` | Variable reference | Planned (iter 18) |
| `var(--custom, fallback)` | With fallback | Planned (iter 18) |

### Tier 2 — Power features (real framework feel)

**Colors:**
| Property | Terminal Mapping | Status |
|---|---|---|
| 8 basic colors | `black red green yellow blue magenta cyan white` | Done |
| 16 colors (bright) | `bright-red`, `bright-cyan`, etc. | Planned |
| 256-color | `color-196` syntax | Planned (iter 10) |
| Truecolor (24-bit) | `#ff0088` hex values | Planned (iter 10) |
| `opacity` | Layer blending | Planned (iter 12) |
| `background: transparent` | Show content below | Planned (iter 12) |

**Positioning:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `position: relative` | Positioning context | Planned (iter 11) |
| `position: absolute` | Position relative to ancestor | Planned (iter 11) |
| `position: fixed` | Position relative to screen | Planned |
| `position: sticky` | Stick to edge on scroll | Planned |
| `top` / `right` / `bottom` / `left` | Cell offsets | Planned (iter 11) |
| `z-index` | Paint order / layer stack | Planned (iter 11) |

**Transitions & Animations:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `transition` | Animate property changes over time | Planned (iter 15) |
| `transition-duration` | Timing in ms | Planned (iter 15) |
| `transition-timing-function` | `ease \| linear \| ease-in-out` | Planned (iter 15) |
| `@keyframes` | Multi-step animations | Planned (iter 15) |
| `animation` | Shorthand for keyframe animations | Planned (iter 15) |
| `animation-iteration-count` | `infinite` for loops | Planned (iter 15) |
| `animation-play-state` | `paused \| running` | Planned (iter 15) |

**Grid:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `grid-template-columns` / `rows` | Cell-based tracks | Planned (iter 16) |
| `grid-gap` | Cell spacing between tracks | Planned (iter 16) |
| `grid-area` | Named grid regions | Planned (iter 16) |
| `grid-auto-flow` | Auto-placement | Planned (iter 16) |

**Selectors:**
| Selector | Terminal Mapping | Status |
|---|---|---|
| `.class` | Class selector | Done |
| `element` | Element type selector | Done |
| `.parent .child` | Descendant combinator | Planned (iter 17) |
| `.parent > .child` | Direct child | Planned (iter 17) |
| `:first-child` / `:last-child` | Positional | Planned (iter 17) |
| `:nth-child()` | Nth element | Planned (iter 17) |
| `:not()` | Negation | Planned (iter 17) |

**Media & Container Queries:**
| Feature | Terminal Mapping | Status |
|---|---|---|
| `@media (width > N)` | Terminal width | Planned (iter 7) |
| `@media (height > N)` | Terminal height | Planned (iter 7) |
| `@media (color-depth)` | Capability detection | Planned (iter 10) |
| `@media (theme)` | Dark/light detection | Planned (iter 10) |
| `@container` | Component-relative queries | Planned (iter 19) |
| `@media (prefers-reduced-motion)` | Skip animations | Planned |

**Functions:**
| Function | Terminal Mapping | Status |
|---|---|---|
| `calc()` | `calc(100% - 10)` → cell math | Planned (iter 18) |
| `min()` / `max()` | Size bounds | Planned (iter 18) |
| `clamp()` | `clamp(10, 50%, 40)` | Planned (iter 18) |

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
| `:hover` | Mouse hover detection | Planned (iter 20) |
| `cursor` | Terminal cursor style | Planned (iter 20) |
| `pointer-events` | Mouse event handling | Planned (iter 20) |
| `user-select` | Copyable text regions | Planned |

**Compositing:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `backdrop-filter: dim` | Dim content behind element | Planned (iter 12) |
| `clip-path` | Rectangular clipping | Planned |
| `contain` | Layout/paint containment (perf) | Planned |

**Scrolling refinements:**
| Property | Terminal Mapping | Status |
|---|---|---|
| `scroll-behavior: smooth` | Animated scrolling | Planned |
| Scrollbar styling | Custom track/thumb characters | Planned |

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
