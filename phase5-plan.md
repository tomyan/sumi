# Phase 5: Components — Elephant Carpaccio Breakdown

Goal: Multiple `.sumi` files composing together, `$prop` rune for component inputs, parent-child component composition.

Target demo:
```html
<!-- counter.sumi -->
<script>
label := $prop("Count")
count := $state(0)

func increment() {
    count = count + 1
}
</script>

<style>
.label { bold: true; color: cyan; }
</style>

<box onkey="increment">
    <text class="label">{label}:</text>
    <text>{count}</text>
</box>
```

```html
<!-- app.sumi -->
<box direction="column">
    <counter label="Clicks" />
    <counter label="Score" />
</box>
```

## Architecture Decision: Struct-based Components

Child components generate a struct with state + a `Layout()` method returning `*layout.Input`. Parent instantiates child components once and calls `Layout()` on each render cycle.

```go
// counter_sumi.go (generated)
type CounterComponent struct {
    label string
    count int
    dirty bool
}

func NewCounterComponent(label string) *CounterComponent { ... }
func (c *CounterComponent) Layout() *layout.Input { ... }
func (c *CounterComponent) HandleKey(key byte) { ... }
```

```go
// app_sumi.go (generated)
func Run() {
    counter0 := NewCounterComponent("Clicks")
    counter1 := NewCounterComponent("Score")
    // ... event loop dispatches to counter0.HandleKey, counter1.HandleKey ...
    // ... doRender calls counter0.Layout(), counter1.Layout() ...
}
```

Root component (no $prop, has event loop) still generates `func Run()`. Child component (has $prop) generates struct-based code.

## Slice 5.1: $prop parsing in script block
- Add `PropDecl` type: `{Name string, DefaultExpr string}`
- Add `PropDecls []PropDecl` to `Script`
- Parse `name := $prop(default)` — same mechanics as `$state`
- Props are also reactive — assignments to props tracked in functions
- Update `resolveStateAssignments` to include prop names
- TDD: empty, single prop, multiple props, string default, int default, mixed with $state

## Slice 5.2: ComponentElement + self-closing tags in template parser
- Add `ComponentElement` AST node: `{Name string, Attributes map[string]string}`
- Unknown tag names (not "text"/"box") → parse as ComponentElement
- Support self-closing syntax: `<counter />` and `<counter label="Clicks" />`
- Parse attributes as before
- Also support `<counter>...</counter>` (children ignored for Phase 5)
- TDD: basic component tag, with attributes, self-closing, inside box, error cases

## Slice 5.3: Child component codegen (struct-based)
- New codegen mode: when script has PropDecls → generate struct-based component
- Generate: `type XxxComponent struct { propFields; stateFields; dirty bool }`
- Generate: `func NewXxx(props...) *XxxComponent` — props as params, state from init exprs
- Generate: `func (c *XxxComponent) Layout() *layout.Input` — returns layout tree
- Generate: `func (c *XxxComponent) HandleKey(key byte)` — dispatches to onkey handler
- Reactive functions become methods, dirty flag is `c.dirty`
- For Phase 5: all prop types are `string`
- TDD: stateless child (props only), stateful child (props + state), with style

## Slice 5.4: Parent codegen references child components
- Add `ComponentInfo` to codegen: `{Name string, ExportedName string, Props []string}`
- `writeInputNode()` handles `*template.ComponentElement` → calls `childVar.Layout()`
- Parent initialization: `counter0 := NewCounterComponent("Clicks")`
- Event dispatch: parent event loop calls `childVar.HandleKey(key)` for each child
- Dirty check: `counter0.Dirty() || counter1.Dirty()`
- TDD: single child, multiple children, props passed through

## Slice 5.5: CLI multi-file compilation + component registry
- `generateDir()` now does two passes:
  1. Parse all .sumi files, build component registry
  2. Generate all, passing registry to codegen for component resolution
- `ComponentRegistry` maps component name → `{Props []string, HasState bool, HasOnkey bool}`
- Component name derived from filename: `counter.sumi` → `counter`
- Validate: all `<componentName>` references resolve to known components
- Integration test: two .sumi files in temp dir, verify both generate valid Go

## Slice 5.6: E2E — counter + app demo
- Create `examples/components/counter.sumi` and `examples/components/app.sumi`
- `app.sumi` uses `<counter label="Clicks" />` and `<counter label="Score" />`
- Run `sumi generate .` → generates both `counter_sumi.go` and `app_sumi.go`
- Verify: compiles, runs, two independent counters with different labels
- Update counter example or create new components example

## Dependencies
```
Slice 5.1 ($prop parsing) ──────────┐
                                    ├→ Slice 5.3 (child codegen) → Slice 5.4 (parent codegen) → Slice 5.5 (CLI) → Slice 5.6 (E2E)
Slice 5.2 (ComponentElement) ───────┘
```

5.1 and 5.2 can run in parallel. 5.3+ are sequential.
