# Signals

State in a sumi component is held in signals: reactive values with
explicit reads and writes. This is the Solid.js model, adapted to Go
methods rather than the getter/setter function pair Solid returns (see
[influences](influences.md) for how this relates to Solid and Svelte 5's
runes). This chapter covers the whole signal surface and the two-layer
update model behind it.

## Creating and using a signal

`sumi.New` constructs a signal from an initial value. Go infers the type
parameter from that value, so `sumi.New(0)` is a `*Signal[int]` and
`sumi.New("")` is a `*Signal[string]`.

```go
count := sumi.New(0)                          // *Signal[int]
count.Get()                                   // read  → 0
count.Set(5)                                   // write
count.Update(func(n int) int { return n + 1 }) // read-modify-write
```

| Method            | Signature                | Effect                              |
| ----------------- | ------------------------ | ----------------------------------- |
| `New[T](initial)` | `T → *Signal[T]`         | Construct a signal.                 |
| `Get()`           | `→ T`                    | Read; subscribes the caller if any. |
| `Set(v)`          | `T →`                    | Replace the value, notify readers.  |
| `Update(fn)`      | `func(T) T →`            | `Set(fn(current))`.                 |

`Set` and `Update` are interchangeable; `Update` exists so a
read-modify-write does not need a separate `Get`. There is no equality
check on `Set` — assigning the current value still notifies. Signal
values are held by the signal, not copied out to component fields, so a
`*Signal[T]` is the thing you pass around and close over in handlers.

## Deriveds

`sumi.From` builds a read-only signal computed from other signals. The
function runs once immediately to establish its dependencies and produce
the first value, and re-runs when any signal it read changes.

```go
count   := sumi.New(2)
doubled := sumi.From(func() int { return count.Get() * 2 }) // *Computed[int]
doubled.Get()                                               // 4
count.Set(10)
doubled.Get()                                               // 20
```

Dependencies are tracked by execution, not declared. Each time the
function runs, its previous subscriptions are dropped and re-established
from the reads that actually happen this time, so a derived that reads
`a` only when `b` is true stops depending on `a` when `b` goes false.
Recomputation is eager: changing a source runs the derived's function
right away, not on the next `Get`. A `*Computed[T]` exposes `Get()` only
— it has no `Set`.

## Auto-unwrap in template text

In template text you reference a signal by its bare name and the
compiler inserts the read:

```sumi
<div>Count: {count}</div>          <!-- compiles to count.Get() -->
<div>Doubled: {count + 1}</div>    <!-- count.Get() + 1 -->
```

The rewrite replaces the signal identifier with `name.Get()` wherever it
appears in the expression. Two consequences follow, and both are easy to
trip over:

- Do **not** write `.Get()` yourself in text. `{count.Get()}` becomes
  `count.Get().Get()` and fails to compile.
- A plain (non-signal) variable is left alone: `{label}` stays `label`.

Attribute expressions are the opposite — raw Go, never rewritten. Write
the read explicitly:

```sumi
<div class={barClass.Get()}>…</div>
```

This split is deliberate: the compiler owns template text, so it can
unwrap there, but an attribute value is code you wrote and it is emitted
unchanged. Dynamic state attributes (`class`, `checked`, `disabled`, …)
still take a raw expression; the value is re-read and re-applied when its
signals change.

## The two-layer update model

A signal write does two separate things, and keeping them apart explains
the runtime's behaviour.

**Dynamic-node patching.** The generated component collects its dynamic
text and its control-flow blocks into one `sumi.Effect`, which subscribes
to every signal those expressions read. A write to any of them re-runs
that effect, recomputing each dynamic node's `Content` and rebuilding
each block's `Children`. The effect is per component, not per node, so it
does not isolate one changed signal to one node — but static markup is
built once and never revisited, and nested components each have their own
effect. `sumi.Effect(fn)` is available directly too; it runs `fn`
immediately, re-runs it on dependency changes, and returns a dispose
function.

**Coarse render scheduling.** Painting the terminal is gated by one
boolean, `App.Dirty`. The event loop sets it after dispatching any input
event, then runs `converge()`, which re-resolves styles, re-lays-out, and
diffs the tree against the last frame — writing only the cells that
changed — repeating up to three times while `Dirty` stays set. There is
no per-signal render queue: an input event yields at most one repaint,
however many signals it touched.

So signals decide *which nodes* recompute and the diff decides *which
cells* paint; the dirty flag decides *when* a frame is produced. Inside
an event handler you do not manage any of it — write your signals and the
frame follows. State changed from outside a handler (a background
goroutine) should be applied through `App.Do`, which runs the closure and
marks the frame dirty.

## Batching

`sumi.Batch(fn)` defers notifications until `fn` returns, so several
writes collapse into one recomputation of each dependent:

```go
sumi.Batch(func() {
    first.Set("Ada")
    last.Set("Lovelace")
}) // a derived reading both recomputes once, not twice
```

Batches nest; deferred work flushes when the outermost one ends.

## Writing without reading

A signal with no readers is inert. `Set` walks the subscriber list and
notifies each one; if nothing ever called `Get` in a tracked context —
no template expression, no derived, no effect reads it — that list is
empty and the write does nothing observable. This is not an error, but a
signal only written and displayed nowhere will not, on its own, cause a
render. Note that Go still requires the variable to be used somewhere, or
the component will not compile.
