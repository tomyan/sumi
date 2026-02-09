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
- Flexbox layout properties
- Box model: padding, margin, border (**border-box by default**)
- Cascade within component scope

**What we ditch:**
- `!important`
- Content-box model
- Float, clear, display: table, all legacy layout
- Complex specificity rules — flat and predictable
- Global scope — everything is scoped to the component

**Terminal-specific CSS:**
- `border-style: single | double | rounded | heavy | none`
- Colors: ANSI names (`red`, `cyan`), 256-color (`color-196`), hex for truecolor (`#ff0088`)
- Text: `bold`, `dim`, `italic`, `underline`, `strikethrough`, `inverse`

**Media queries:**
- `@media (width > N)` / `@media (height > N)` — terminal dimensions
- `@media (color-depth: monochrome | ansi | 256color | truecolor)` — graceful color degradation
- `@media (theme: dark | light)` — terminal theme detection

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
