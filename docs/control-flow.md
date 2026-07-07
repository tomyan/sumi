# Control flow

Templates branch and repeat with brace blocks. Sumi drops Svelte's
sigils — `{if}` where Svelte writes `{#if}` — and the expressions inside
are Go. This chapter covers conditionals, loops, and what `key=` buys.

## Conditionals

```sumi
{if count.Get() > 0}
    <p>Non-empty</p>
{else}
    <p>Empty</p>
{/if}
```

The condition is a Go boolean expression. It is not template text, so a
signal needs an explicit `.Get()` — the auto-unwrap that applies to
`{count}` in body text does not apply here. The `{else}` branch is
optional. A branch may hold any number of elements, text, and nested
blocks.

### No `{else if}`

There is no `{else if}` form. Writing one produces invalid Go and fails
at generation. Chain conditions by nesting an `{if}` inside `{else}`:

```sumi
{if n.Get() == 0}
    <p>none</p>
{else}
    {if n.Get() == 1}
        <p>one</p>
    {else}
        <p>many</p>
    {/if}
{/if}
```

## Loops

```sumi
{for i, item := range items.Get() key=item.ID}
    <li>{i}: {item.Title}</li>
{/for}
```

Everything between `{for` and the trailing `key=` is a real Go range
clause, emitted verbatim into a `for … range` loop. Both the index and
value bindings are in scope in the body, and — being ordinary text
expressions — `{i}` and `{item.Title}` interpolate directly. `items` is
a signal, so it reads as `items.Get()`; because the clause is Go, not
template text, you write that `.Get()` yourself.

The `key=` attribute is split off the end of the clause on the last
occurrence of ` key=`, so a range variable whose name contains `key`
does not confuse the parser. The key is a Go expression evaluated per
item.

### What `key=` buys

Without a key, list items diff positionally: the box at index 0 is
compared with the new index 0, and so on. Reordering or inserting then
rebuilds every box from the first change onward.

With a key, items are matched by identity across renders. When the list
reorders, each item's existing box is found by its key and moved rather
than rebuilt, so per-box state that lives in the tree — focus, an
input's edit cursor, scroll position — travels with the item instead of
staying pinned to a slot. Use a key that is stable and unique per item
(a record ID, not the array index).

## Nesting

Blocks nest freely and mix with elements. A loop can contain
conditionals, a conditional can contain loops, and either can sit inside
an element or hold elements:

```sumi
<ul>
    {for _, task := range tasks.Get() key=task.ID}
        <li>
            {if task.Done}
                <span class="done">{task.Title}</span>
            {else}
                <span>{task.Title}</span>
            {/if}
        </li>
    {/for}
</ul>
```

Each block must be closed by its own terminator — `{/if}` for `{if}`,
`{/for}` for `{for}` — before the enclosing block closes. An unterminated
block is a generation error.

## How blocks re-render

A container holding control flow compiles so that the block's contents
are produced by a function, and that rebuild runs inside the component's
`sumi.Effect`. The effect subscribes to whatever signals the condition or
range clause reads, so changing one of them re-evaluates the block in
place — picking the live branch, or ranging the current list — as part of
the component's dynamic-node sync described in [signals](signals.md). The
single dirty flag then schedules one repaint, and the tree diff writes
only the cells that actually changed. Keyed loops reuse the matched boxes
across that rebuild rather than reconstructing them.
