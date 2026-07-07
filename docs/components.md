# Components and reactivity

A `.sumi` file is a single-file component: a `<script>` block of Go, a
`<style>` block of CSS, and a template of HTML elements. `sumi generate`
compiles each file to a Go constructor (`counter.sumi` →
`NewCounter(CounterProps) *sumi.Component`) that your `main.go` — or
another component — mounts.

## Signals

State lives in signals — fine-grained reactive values in the Solid.js
style. Reading a signal inside a template expression subscribes that
expression; writing the signal re-renders exactly what depends on it.

```go
count := sumi.New(0)                 // *Signal[int]
count.Get()                          // read
count.Set(5)                         // write
count.Update(func(n int) int { return n + 1 })

doubled := sumi.From(func() int {    // derived: recomputes when count changes
	return count.Get() * 2
})
```

In the template, `{count}` renders the current value and stays live —
signals are auto-unwrapped in text expressions. Attribute expressions
are raw Go: write `class={barClass.Get()}` explicitly.

## Template expressions

- `{expr}` in text — any Go expression; signals auto-unwrap.
- `name="literal"` — string attribute.
- `name={expr}` — expression attribute (raw Go, no auto-unwrap).
- `{name}` — shorthand for `name={name}`.

## Control flow

```sumi
{if count.Get() > 0}
	<p>Non-empty</p>
{else}
	<p>Empty</p>
{/if}

{for _, item := range items.Get() key=item.ID}
	<div>{item.Title}</div>
{/for}
```

The `{for}` clause is a real Go range clause. `key=` enables identity-based
diffing so reordered items keep their boxes (and focus) instead of being
rebuilt positionally.

## Event handlers

Declare functions in `<script>` and reference them from `on<type>`
attributes:

```sumi
<script>
func save(evt *sumi.DOMEvent) {
	evt.PreventDefault()
}

func handleKey(evt sumi.Event) {
	if evt.Rune == 'q' { sumi.Quit() }
}
</script>

<div onkey="handleKey">
	<button onclick={save}>Save</button>
</div>
```

Handlers taking `*sumi.DOMEvent` participate in capture/bubble dispatch
with `StopPropagation` / `PreventDefault` (see the events sections in
[elements](elements.md)). A function named `handleKey` taking
`sumi.Event` becomes the component's raw event handler — it sees every
input event after DOM dispatch. Deviation: that wiring is name-based;
only `handleKey` is picked up.

Zero-argument handlers get automatic Ctrl+C/quit-signal handling;
declaring any event-aware handler hands you full control (call
`sumi.Quit()` yourself).

## Props

Declare props as plain `var` declarations; consumers pass them via the
generated props struct:

```sumi
<!-- greeting.sumi -->
<script>
var name string
</script>
<p>Hello, {name}!</p>
```

```go
greeting.NewGreeting(greeting.GreetingProps{Name: "Tom"})
```

Callback props are function-typed vars, passed the same way. `bind:value`
on an input-like component binds a parent signal to the child's value in
both directions.

## Composition

Components compose two ways:

- **In a template** — `<counter label="Clicks" />` mounts a sibling
  component (its constructor must be importable by the generate CLI).
- **In Go** — embed a child's tree directly:

```go
c1 := counter.NewCounter(counter.CounterProps{Label: "Clicks"})
root := parent.Tree // any *layout.Input
root.Children = append(root.Children, c1.Tree)
```

Each component carries its own stylesheet; styles are scoped by the
component's own cascade and do not leak into embedded children.

## Slots and snippets

A component template can declare `<slot:name />` placeholders with
optional fallback content; consumers fill them with
`{slot name}...{/slot}` blocks. `{snippet name(params)}...{/snippet}`
defines a local template function invoked with `{render name(args)}`.

## Lifecycle

`sumi.Quit()` ends the app from any handler. A component's `Dispose`
runs when the app exits (or when a [FrameLog](inline-mode.md) frame is
archived or removed). There is no polling: a single dirty flag set by
signal writes schedules the next render pass.
