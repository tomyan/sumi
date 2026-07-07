# Influences

Sumi's design comes from three places: Svelte's single-file
components, Solid's signals API, and Ink's terminal rendering. This
chapter says what was taken from each and what changed in the move
to Go.

## Svelte

The `.sumi` file layout follows Svelte's single-file components: a
`<script>` block, a `<style>` block, and markup. The template syntax
is modelled on Svelte's with some differences. Control-flow blocks
drop Svelte's sigils — `{if}` / `{else}` / `{/if}` where Svelte
writes `{#if}` / `{:else}` / `{/if}` — and the expressions inside
are Go. Loops keep Svelte's block form but not its clause: `{for i,
item := range items.Get() key=item}` is a Go range clause plus a
`key=` attribute, where Svelte writes `{#each items as item (key)}`.
`{expr}` interpolation and `bind:value` are spelled as in Svelte,
and event attributes are `onclick={handler}`, matching Svelte 5's
event syntax.

The compilation strategy is also Svelte's: `sumi generate` compiles
components to plain Go, the way Svelte compiles to plain JavaScript,
so the binary contains no template interpreter and no virtual DOM.

The idea of rendering HTML elements with a user-agent stylesheet on
a terminal cell grid comes from [svelterm](https://svelterm.dev),
which does this for Svelte components. Sumi tracks svelterm's CSS
support matrix so the two projects agree on what CSS means on
character cells.

## Solid (and Svelte 5's runes)

Reactivity is fine-grained signals with explicit reads and writes,
modelled on Solid. The shape differs with the language: Solid's
`createSignal` returns a getter/setter pair of functions, while a
sumi signal is one value with methods:

```sumi
<script>
count := sumi.New(0)
doubled := sumi.From(func() int { return count.Get() * 2 })
</script>
```

Svelte 5's runes are also fine-grained runtime signals — the
difference is only the API surface. Runes let you write `count++`
because the Svelte compiler rewrites plain JavaScript reads and
assignments into signal operations. There is no equivalent layer
available in Go: `count++` on an int can't be made reactive without
changing the language. So reads and writes are method calls —
`count.Get()`, `count.Set(n)`, `count.Update(fn)` — as in Solid.

One exception: `{count}` in template text auto-unwraps the signal,
since the compiler owns template syntax. Attribute expressions are
raw Go (`class={barClass.Get()}`), so nothing is rewritten in code
you write.

## Ink

[Ink](https://github.com/vadimdemedes/ink) demonstrated component
trees, flexbox, and declarative rendering for terminal UIs, in
React/Node. Sumi differs in vocabulary and packaging: HTML elements
and CSS instead of Ink's component set and style props, Go instead
of Node, and output is a single static binary with no runtime
dependency.

## Consequences of Go

- Components compile to constructors and a build-once tree with
  targeted updates — no reflection or eval at render time.
- Event handlers take plain structs (`sumi.Event`,
  `*sumi.DOMEvent`) with DOM-style bubbling.
- The framework ships as source with the CLI, so `sumi init` and
  offline builds work without a module proxy.
