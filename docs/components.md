# Components

A `.sumi` file is a single-file component: a `<script>` block of Go, a
`<style>` block of CSS, and a template of HTML elements. `sumi generate`
compiles each file to a Go constructor that your `main.go` — or another
component — mounts.

This chapter covers component files, props, composition, and scoped
styles. Reactivity has its own chapters — [signals](signals.md),
[templates](template-syntax.md), [control flow](control-flow.md),
[snippets](snippets.md) — events are in [events](events.md).

## Component files and naming

The component's name comes from the filename: the `.sumi` extension is
dropped, hyphens are removed, and the first letter is capitalized — and
only the first letter, so `split-panel.sumi` is `Splitpanel`, not
`SplitPanel`. `counter.sumi` → `Counter`.

Each file generates a props struct and a constructor beside it, in the
same package:

```go
func NewCounter(props CounterProps) *sumi.Component
```

`main.go` mounts the root component by calling its constructor and
passing it to the runtime:

```go
package main

import "github.com/tomyan/sumi/runtime/tui"

func main() {
	tui.Run(NewCounter(CounterProps{}))
}
```

Everything spelled `sumi.X` in a `<script>` — `sumi.New`, `sumi.Event`,
`sumi.DOMEvent`, `sumi.Quit`, the type aliases below — resolves to the
runtime prelude (`.../runtime/prelude`), which the generated file imports
under the alias `sumi`.

## Props

Declare props as plain package-level `var` declarations in `<script>`.
Each becomes a field on the props struct, with the first letter
capitalized:

```sumi
<!-- greeting.sumi -->
<script>
var name string
</script>
<p>Hello, {name}!</p>
```

```go
NewGreeting(GreetingProps{Name: "Tom"})
```

In a parent template, pass props as attributes — the lowercase prop name,
capitalized to the struct field (`name` → `Name`). A string literal is a
string value; a `{expr}` attribute is raw Go:

```sumi
<Greeting name="Tom" />
<Greeting name={who.Get()} />
```

### Callback props

A prop whose type is a function is a callback. Declare it, pass a
function in, and invoke it from the child's own handlers:

```sumi
<!-- child field.sumi -->
<script>
var onclear func()
</script>
<button onclick={onclear}>clear</button>
```

```sumi
<!-- parent -->
<script>
func clear() { count.Set(0) }
</script>
<Field onclear={clear} />
```

### bind: and cross-component signal flow

State is shared between components by passing the *signal itself*, not
its value. Declare the child prop as a signal type; both sides then hold
the same `*sumi.Signal`, so a write on either is seen by both:

```sumi
<!-- child field.sumi -->
<script>
var count *sumi.Signal[int]
var label string
</script>
<span>{label}: {count}</span>
```

```sumi
<!-- parent -->
<script>
clicks := sumi.New(0)
</script>
<Field label="Clicks" bind:count={clicks} />
```

`{count}` in the child auto-unwraps because the prop type contains
`Signal` (see [signals](signals.md)). For a signal-typed prop,
`bind:count={clicks}` and `count={clicks}` compile identically — the
signal is passed by reference either way — so `bind:` here is a
convention marking two-way intent. It does real work on the fundamental
elements (`textedit`, `scrollbar`), whose runtime wires it as binding.

### Snippet props

A prop typed `func() []*sumi.Input` is a *snippet* — a chunk of template
the parent supplies. The child declares it and renders it with
`{render name()}`:

```sumi
<!-- card.sumi -->
<script>
var children func() []*sumi.Input
var footer   func() []*sumi.Input
</script>
<div>{render children()}</div>
<div>{render footer()}</div>
```

The parent fills them from the tag body — a `{snippet name()}...{/snippet}`
block becomes the matching prop, and the remaining body becomes the
implicit `children` snippet. Full treatment in [snippets](snippets.md).

## The constructor model

Components are runtime constructors, not compile-time inlining: `NewFoo`
builds the component's `*sumi.Input` tree, wires one reactive effect over
its dynamic nodes, and returns a `*sumi.Component` holding that tree, its
`OnEvent` handler, and its stylesheet. Mounting a child calls its
constructor once during the parent's construction and splices the
returned `.Tree` into the parent's tree.

Updates are component-grained: a component has a single effect, so a
tracked signal change re-evaluates all of that component's dynamic
expressions and children. The per-cell minimalism on screen comes from
the tree diff that follows — only changed cells are written — not from
per-node effects (see [signals](signals.md)).

Every `.sumi` compiles to this constructor form — a file with no
`<script>`, or one whose script has only plain `:=` bindings and
functions, generates the same `NewFoo`/`FooProps` constructor as a
reactive one (with an empty effect when there is nothing dynamic to
track). So any parent can mount a child; there is no reactive-declaration
requirement and no separate static path.

## Scoped styles

Each component carries its own stylesheet, and a mounted child's subtree
is a style boundary. Each component's `<style>` rules are resolved against
its own subtree, and the boundary holds in both directions: the parent's
rules do not cross into a mounted child, and the child's rules do not
reach back out into the parent. Nesting works to any depth — a
grandchild's stylesheet resolves against the grandchild's subtree.

So a child styles itself with its own CSS classes, and the parent cannot
restyle the child's internals through a class selector; pass a prop if
the parent needs to influence the child's appearance.

## Referring to components

A local tag is the constructor name verbatim, so keep local names
single-word (`<Field>` → `NewField`); an imported component is
`<pkg.Name />` → `pkg.NewName`. Mounting gives you the child's structure,
props, callbacks, `bind:` signals, snippets, and its own scoped CSS.
