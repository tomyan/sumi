# Philosophy

Sumi did not invent its shape. It borrows deliberately from three
projects, and the places where it diverges are the places where Go is
not JavaScript.

## The Svelte in sumi

A `.sumi` file is a Svelte single-file component, transplanted: a
`<script>` for behaviour, a `<style>` for scoped CSS, and a template
of real HTML elements. The template control flow — `{if}` / `{else}`
/ `{/if}`, `{for}` with keyed diffing — is Svelte's block syntax with
Go expressions inside. `bind:value`, `on*` event attributes, and
expression props all follow Svelte's attribute grammar. The compiler
strategy is Svelte's too: the component you write is not what runs —
`sumi generate` compiles it to plain Go the way Svelte compiles to
plain JavaScript, so there is no virtual DOM and no framework
interpreter in your binary.

The user-agent stylesheet idea — elements arriving with sensible
defaults, a cascade you can override — comes from the browser by way
of [svelterm](https://svelterm.dev), sumi's sibling project, which
renders Svelte components to a terminal cell grid. Sumi tracks
svelterm's CSS support matrix closely; the two projects share one
answer to "what does CSS mean on character cells?"

## The Solid in sumi (and why not runes)

Reactivity is fine-grained signals, read and written explicitly:

```sumi
<script>
count := sumi.New(0)
doubled := sumi.From(func() int { return count.Get() * 2 })
</script>
```

Svelte 5's runes are also runtime signals — `$state` and `$derived`
are fine-grained reactivity, not the old invalidate-and-rerun model.
The difference is the API surface. Runes work because Svelte's
compiler rewrites plain JavaScript: `count++` becomes a signal write,
reading `count` in a template becomes a tracked read. Go offers no
such layer — there is no compiler pass that could make `count++`
reactive without ceasing to be Go. So sumi takes Solid's position:
the signal is an ordinary value you call methods on. `count.Get()`
is a tracked read, `count.Set(n)` / `count.Update(fn)` is a write,
and both are exactly what they look like. In a language built on
explicitness, the explicit API is the natural one — the same
reasoning that makes `err != nil` idiomatic where other languages
throw.

Text interpolations are the one concession: `{count}` in template
text auto-unwraps the signal, because the compiler owns that syntax
anyway. Attribute expressions stay raw Go — `class={barClass.Get()}`
— so there is never doubt about what an expression means.

## The Ink in sumi

[Ink](https://github.com/vadimdemedes/ink) proved the premise:
component trees, flexbox, and declarative rendering belong in the
terminal. Sumi keeps Ink's conviction and swaps the foundations —
HTML elements instead of ad-hoc components, real CSS instead of
props, Go instead of Node, and a single static binary instead of a
runtime — which is also the honest summary of sumi versus Ink:
`go build` is the deployment story.

## What Go changes

- **Compile-time everything.** Components are Go constructors; the
  template becomes a build-once tree with surgical updates. There is
  no eval, no reflection-driven rendering, and dead simple stack
  traces.
- **Events are values.** Handlers take `sumi.Event` or
  `*sumi.DOMEvent` — plain structs, DOM-style bubbling — not closures
  over framework context.
- **One binary.** No node_modules, no runtime to install; cross-
  compile with the toolchain you already have. The framework ships
  as source with the CLI so `sumi init` works offline.
