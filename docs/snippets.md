# Snippets

A snippet is a chunk of template that a component renders on demand.
Snippets are how a component takes markup from its caller — the role
snippets play in Svelte 5 (which also replaced its slot mechanism with
them) and `children` plays in React. Sumi had a slot mechanism; it was
removed in favour of snippets, which are one idea covering both named
and default content. This chapter is the full
reference; [components](components.md) keeps a short overview that points
here.

A snippet is represented as a Go closure returning child nodes:

```go
func() []*sumi.Input
```

Everything below is that one type, produced and consumed in different
places.

## Local snippets

Inside a component, `{snippet name()}…{/snippet}` declares a reusable
fragment and `{render name()}` renders it. This factors repeated markup
without a separate component:

```sumi
<script>
rows := sumi.New([]string{"a", "b"})
</script>

{snippet divider()}<hr />{/snippet}

<div>
    {render divider()}
    <div>{rows}</div>
    {render divider()}
</div>
```

Each local snippet compiles to a closure hoisted to component scope, so
a single declaration can be rendered from several places. Parameters are
ordinary Go and let one snippet vary per call:

```sumi
{snippet cell(label string)}<td>{label}</td>{/snippet}
```

## Snippet props

A component accepts a snippet from its caller by declaring a prop whose
type is `func() []*sumi.Input` (optionally with parameters). The type is
what marks it as a snippet — a prop is a snippet prop exactly when its Go
type starts with `func` and ends with `[]*sumi.Input`. The component
renders it with `{render}` like a local one:

```sumi
<!-- card.sumi -->
<script>
var title    string
var footer   func() []*sumi.Input
var children func() []*sumi.Input
</script>
<div class="card">
    <div class="card-title">{title}</div>
    <div class="card-body">{render children()}</div>
    <div class="card-foot">{render footer()}</div>
</div>
```

## Passing snippets from the consumer

The caller fills those props from the component tag's body. A
`{snippet name()}…{/snippet}` block inside the tag becomes the matching
named prop; everything else in the body becomes the implicit `children`
snippet:

```sumi
<Card title="Hi">
    <p>this goes to children</p>
    {snippet footer()}<p>this goes to footer</p>{/snippet}
</Card>
```

Here `footer` fills the `footer` prop and the `<p>this goes to
children</p>` fills `children`. A component with no `children` prop
simply ignores loose body content.

Ordinary props are still passed as attributes (`title="Hi"` above);
snippet props come only from the body, never from an attribute.

## Resolution and defaults

`{render name}` resolves against local snippets first, then snippet
props, so a local `{snippet}` shadows a prop of the same name. A name
that matches neither is a generation error:

```
{render foo} names an unknown snippet: declare a {snippet foo()} or a
snippet prop
```

A snippet prop the caller does not pass renders nothing — the compiler
defaults an unpassed snippet prop to a closure returning `nil`, so
`{render footer()}` on a `<Card>` with no footer produces empty output
rather than a nil-pointer panic.

## Migrating from slots

Slots were removed. The old `<slot:name />` placeholder and `{slot
name}…{/slot}` block now raise a generation error pointing at the
replacement:

```
slots were removed; declare a {snippet name()} inside the component tag
and {render name()} in the component template
```

The translation is direct: a slot placeholder in the component becomes
`{render name()}`, and the consumer's slot-fill block becomes
`{snippet name()}…{/snippet}` — with default (unnamed) content becoming
loose body that lands in `children`.

## Limitations

Two restrictions follow from snippets compiling to component-scope
closures. Both surface as a Go compile error on the generated file, not
a friendly diagnostic:

- **A snippet cannot capture a `{for}` variable.** Because the closure
  is hoisted to component scope, a snippet declared inside a loop and
  referencing the loop variable emits code where that variable is out of
  scope (`undefined: item`). Render per-item markup inline in the loop
  instead.
- **A child component cannot appear inside a snippet body.** Component
  instances are constructed at component scope before the tree is built,
  and that construction does not descend into snippet bodies, so a
  component tag there compiles to a reference to an uninstantiated
  variable (`undefined: badge0`). Plain HTML elements and control flow in
  a snippet body are fine.
