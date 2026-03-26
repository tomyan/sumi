# Signal-Based Component Architecture

## End-state vision

Sumi is a declarative TTY framework where `.sumi` single-file components are the primary authoring model. Components compile to Go functions with runtime signal-based reactivity. The signal runtime is a standalone Go library; the compiler handles templates, styles, and component wiring.

## Core principles

1. **`.sumi` files are the product** — template syntax and terminal CSS are what make sumi valuable
2. **Script blocks are valid Go** — `go/ast` parses them, `gopls` works on them
3. **Runtime signals, not compile-time transformation** — one reactive model everywhere
4. **Components are Go functions** — closure-scoped state, no namespace renaming
5. **Go's type system enforces contracts** — prop types, slot signatures, signal vs plain values
6. **One `.sumi` file per package** — Go packages are the encapsulation boundary

## Component model

### File structure

One `.sumi` file per Go package. The filename determines the component name:

```
counter/counter.sumi        → package counter, func NewCounter(props CounterProps) *tui.Component
textinput/text-input.sumi   → package textinput, func NewTextInput(props TextInputProps) *tui.Component
myapp/app.sumi              → package main, func NewApp(props AppProps) *tui.Component
```

Package-level Go code (shared constants, helpers) lives in `.go` files alongside the `.sumi` file.

### Generated output

Each `.sumi` file generates:
- `NewFoo(props FooProps) *tui.Component` — the component constructor
- `FooProps` struct — typed props derived from `var` declarations in the script block

The root app's `main.go` wires things together:

```go
func main() {
    tui.Run(NewApp(AppProps{}))
}
```

Testing uses the runtime directly:

```go
func TestApp(t *testing.T) {
    app := tui.TestApp(NewApp(AppProps{}), 80, 24)
    app.Step(someEvent)
    // assert on app.Buffer()
}
```

No generated `Run()` or `CreateApp()` — the runtime provides these.

## Script block

The `<script>` block is **valid Go**. The compiler parses it with `go/ast` to understand declarations.

### Props — `var` declarations

```go
var value *signal.Signal[string]      // signal prop (bindable)
var label string = "Count"            // plain prop with default
var placeholder string                // plain prop, zero-value default
```

`var` declarations become fields on the generated `FooProps` struct and parameters to the constructor. The compiler extracts types from the Go AST.

### State — short declarations

```go
count := signal.New(0)                               // reactive state
doubled := signal.From(func() int { return count.Get() * 2 })  // derived
visible := true                                       // plain local variable
```

`:=` declarations with `signal.New` or `signal.From` are identified as signals. The compiler auto-unwraps them in template expressions (`{count}` → `count.Get()`).

### Functions — closures

```go
func increment() {
    count.Update(func(n int) int { return n + 1 })
}

func handleKey(evt input.Event) {
    if evt.Kind == input.EventSignal { app.Quit(); return }
    if evt.Rune == 'q' { app.Quit(); return }
}
```

Function bodies are passed through as valid Go — no assignment rewriting, no dirty injection. Signal `.Set()` / `.Update()` calls trigger reactivity via the runtime.

## Props system

### Props struct

The compiler generates a props struct from `var` declarations:

```go
// From text-input.sumi script:
//   var value *signal.Signal[string]
//   var placeholder string = "type here"
//   var readonly bool

// Generated:
type TextInputProps struct {
    Value       *signal.Signal[string]
    Placeholder string
    Readonly    bool
}
```

### Passing props

Template attributes map to struct fields. Go's type system validates at compile time:

```html
<TextInput bind:value={name} placeholder="Enter name" />
```

Generates:

```go
textinput.NewTextInput(textinput.TextInputProps{
    Value:       name,           // bind: passes the signal
    Placeholder: "Enter name",   // literal string
})
```

### Prop conventions

| Template syntax | Generated code | Use case |
|---|---|---|
| `prop="literal"` | `Prop: "literal"` | Static value |
| `prop={expr}` | `Prop: expr` or `Prop: expr.Get()` | Expression (compiler unwraps signals for non-signal props) |
| `bind:prop={signal}` | `Prop: signal` | Shared signal reference |
| `prop={callback}` | `Prop: callback` | Function prop |

## Template language

### Expressions

```html
<text>{count}</text>              <!-- signal auto-unwrapped to count.Get() -->
<text>{count * 2 + 1}</text>     <!-- Go expression, signals unwrapped -->
<text>Hello, {name}</text>        <!-- mixed literal and expression -->
```

### Control flow

```html
{if condition}
    <text>Yes</text>
{else}
    <text>No</text>
{/if}

{for i, item := range items.Get()}
    <text>{i}: {item.Name}</text>
{/for}
```

### Elements

```html
<box class="container" direction="row" border="single">
    <text class="title">Hello</text>
</box>
```

### Components

```html
<Counter label="Clicks" />

<TextInput bind:value={name} placeholder="Enter name" />

<Card title="Stats">
    <text>Content here</text>
</Card>
```

## Slots

Slots let a component accept template content from its consumer.

### Defining slots (component side)

Use `<slot:name>` elements as placeholders in the component's template:

```html
<!-- card/card.sumi -->
<box class="card" border="single">
    <box class="header">
        <slot:header>
            <text>Untitled</text>
        </slot:header>
    </box>
    <slot:children />
</box>
```

- `<slot:children />` — default slot, receives unnamed content
- `<slot:header>...</slot:header>` — named slot with default content
- Self-closing `<slot:name />` — no default (empty if not provided)

### Filling slots (consumer side)

Use `{slot name}...{/slot}` blocks:

```html
<Card>
    {slot header}
        <text bold="true">Dashboard</text>
    {/slot}
    <text>Body content</text>
</Card>
```

- Unnamed content between component tags fills `<slot:children />`
- `{slot name}...{/slot}` fills the named slot
- Omitted slots render their default content (or nothing if no default)

### Accessing default content from override

Inside a `{slot}` block, `<slot:default />` renders what the component's default would have been:

```html
<Button label="Submit">
    <text>★ </text>
    <slot:default />
</Button>
```

Renders `★ Submit` — the consumer wraps the component's default content. `<slot:default />` is only valid inside a `{slot}` block.

### Scoped slots

Components can pass data to slot content. The `<slot:name>` element passes attributes, the `{slot}` block receives typed parameters:

**Component (list/list.sumi):**
```html
<box>
    {for i, item := range items.Get()}
        <slot:children {item} {i} />
    {/for}
</box>
```

**Consumer:**
```html
<List items={data}>
    {slot children(item Item, i int)}
        <text>{i}: {item.Name}</text>
    {/slot}
</List>
```

In Go, scoped slots compile to function props:

```go
type ListProps struct {
    Items        *signal.Signal[[]Item]
    ChildrenSlot func(item Item, i int) []*layout.Input
}
```

### Snippets

Template-level reusable fragments within a single component:

```html
{snippet listItem(item Item, selected bool)}
    <box class={selected ? "selected" : ""}>
        <text>{item.Name}</text>
    </box>
{/snippet}

{for i, item := range items.Get()}
    {render listItem(item, i == selected.Get())}
{/for}
```

Snippets are local to the component — not exported, not passed between components.

## Reactivity

### Signal runtime (`runtime/signal/`)

Standalone Go library with no sumi dependencies:

```go
signal.New[T](initial T) *Signal[T]         // reactive state
signal.From[T](fn func() T) *Computed[T]    // derived value, auto-tracks
signal.Effect(fn func()) func()              // side effect, returns dispose
signal.Batch(fn func())                      // defer notifications
```

### How rendering works

1. Signal change (`.Set()` / `.Update()`) propagates through dependency graph
2. `signal.Effect` updates affected layout tree nodes + calls `app.Wake()`
3. App event loop wakes, calls render
4. Render: `Layout()` → `DiffTrees()` → `ApplyChanges()` (or full redraw)

Fine-grained: only effects whose dependencies changed re-run. Only changed cells are written to terminal.

### Component lifecycle

- **Mount**: component function runs — creates signals, effects, layout tree
- **Update**: signal changes propagate through effects automatically
- **Unmount**: `Component.Dispose()` cleans up all effects

### `{if}` and component lifecycle

When `{if}` toggles a component off then on, the component **resets** — fresh signals, fresh effects. To preserve state across visibility toggles, lift the signal to the parent and pass via `bind:`.

### Keyed reconciliation in `{for}`

```html
{for _, item := range items.Get() key=item.Id}
    <TodoItem bind:done={item.Done} label={item.Text} />
{/for}
```

- Components in `{for}` are cached by key
- List reorder → existing component instances move (state preserved)
- Key removed → component disposed
- New key → fresh component created
- No key → components reset on every rebuild (current behavior)

## Event dispatch

Events bubble through the layout tree, crossing component boundaries.

1. Keypress arrives at the app
2. Dispatched to the focused element's handler
3. If not stopped (`stopPropagation()`), bubbles to parent box's handler
4. Continues up through component boundaries to the root

Components don't need to know about parent shortcuts. A TextInput handles character keys and calls `stopPropagation()`. Unhandled keys (like `q` for quit) bubble to the root handler.

## Env signals

Framework-provided signals for terminal state:

```go
// In .sumi script:
width := tui.Env[int]("width")
height := tui.Env[int]("height")
```

These are `*Signal[int]` updated by the framework on SIGWINCH. Components subscribe by calling `.Get()` — standard signal dependency tracking.

## What this eliminates from the current codegen

| Current (compile-time inlining) | New (signal components) |
|---|---|
| `codegen_inline.go` (~350 lines) | Removed |
| `codegen_inline_stateful.go` (~225 lines) | Removed |
| Namespace renaming (`counter0_count`) | Not needed — closure scope |
| `buildStateNameMap`, `writeNamespacedFuncBody` | Not needed |
| `resolveCallbackProps`, `replaceIdentifier` for inlining | Not needed |
| Two-pass compilation (parse all → build registry → inline) | Single pass — each `.sumi` generates independently |
| `codegen_reactive.go` dirty injection | Not needed — signals handle reactivity |
| `sync()` function generation | Replaced by `signal.Effect` |
| `dirty` flag | Replaced by signal propagation + `app.Wake()` |
| Script block transformation | Not needed — script is valid Go, passed through |

## What the new compiler does

1. Parse `.sumi` file into script, style, and template sections
2. Parse script with `go/ast` — identify `var` declarations (props), signal declarations, functions
3. Parse style as terminal CSS, resolve to `render.Style` structs
4. Parse template — elements, control flow, slots, component usage
5. Generate:
   - `FooProps` struct from `var` declarations
   - `NewFoo(props FooProps) *tui.Component` function containing:
     - Signal declarations (pass-through from script)
     - Function closures (pass-through from script)
     - Layout tree construction (from template + styles)
     - `signal.Effect` for expression nodes (auto-generated)
     - Child component instantiation (from template component tags)
     - Slot wiring
   - Props struct field → constructor parameter mapping

## Project structure

```
sumi/
  cmd/sumi/              # compiler CLI
  runtime/
    signal/              # reactive signals (standalone, no sumi deps)
    tui/                 # app lifecycle, Component type, Run, TestApp
    layout/              # flexbox layout engine
    render/              # cell buffer, terminal output
    css/                 # terminal CSS parser
    input/               # keyboard/mouse input parsing
    term/                # terminal size, capabilities
    vt100/               # VT100 terminal emulator
  parser/
    script/              # Go AST analysis for script blocks
    style/               # terminal CSS parser
    template/            # template parser (elements, control flow, slots)
  codegen/               # Go code generator
  components/            # first-party component library
    textinput/
    button/
    scrollbar/
    ...
```

## Iteration plan

### Slice 1: Component type + root component generation ✓
- Add `tui.Component` struct with `Tree`, `Dispose`
- Add `tui.Run(comp)` and `tui.TestApp(comp, w, h)`
- New codegen generates `NewFoo(FooProps) *Component` from a simple `.sumi` file
- Parse script with `go/ast` for signal identification
- Convert counter example
- **You see**: counter works as a signal-based component function

### Slice 2: Props ✓ struct generation
- `var` declarations in script → `FooProps` struct fields
- Template attributes → struct literal fields
- Go type system validates at compile time
- Convert examples with literal props
- **You see**: typed props passed between components

### Slice 3: Child ✓ component composition
- `<Counter label="Clicks" />` generates `counter.NewCounter(counter.CounterProps{Label: "Clicks"})`
- Child's tree embedded in parent's layout tree
- Event bubbling across component boundaries
- **You see**: parent-child composition works

### Slice 4: Bound ✓ props (bind:value)
- `bind:value={name}` passes parent's `*Signal[T]` to child
- Shared signal reference — no copying, no renaming
- Convert text input example
- **You see**: two-way binding works

### Slice 5: Slots ✓
- `<slot:name />` placeholders in component templates
- `{slot name}...{/slot}` content in consumer templates
- Default slot content with `<slot:default />` override access
- Scoped slots with typed parameters
- **You see**: flexible component composition with slots

### Slice 6: Keyed ✓ reconciliation
- `{for ... key=expr}` caches component instances by key
- List reorder preserves component state
- Key removal disposes component
- New key creates fresh component
- **You see**: efficient list rendering with stable identity

### Slice 7: Snippets ✓
- `{snippet name(params)}...{/snippet}` defines template functions
- `{render name(args)}` invokes them
- Local to the component
- **You see**: template reuse within components

### Slice 8: Env ✓ signals and scroll
- `tui.Env[int]("width")` → framework-provided `*Signal[int]`
- Scroll state as signals
- **You see**: responsive layout and scrolling work

### Slice 9: Remove old codegen
- Delete compile-time inlining code
- Delete dirty/sync generation
- Delete namespace renaming
- Delete assignment rewriting
- Remove `$state`/`$derived`/`$prop`/`$env` from parser
- **You see**: cleaner, smaller codebase

### Slice 10: Migrate all examples and components
- Convert every `.sumi` file to new syntax
- Update all tests
- **You see**: full framework running on new architecture
