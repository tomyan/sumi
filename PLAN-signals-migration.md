# Signals Migration Plan

Migrate sumi's reactivity from compile-time `$state` runes to runtime Solid-style signals.

## Current state

- `$state(v)` → plain variable + `dirty = true` on reassignment
- `$derived(expr)` → plain variable recomputed in `sync()`
- `$prop(v)` → inlined parent variable with namespace renaming
- `$env(key)` → `term.GetSize()` called in `sync()`
- Single `dirty` bool per component, full re-layout on any change
- Compiler transforms script block (rewrites assignments, generates sync/dirty)

## Target state

- `sumi.New[T](v)` → `*Signal[T]` with `.Get()`, `.Set()`, `.Update()`
- `sumi.From[T](fn)` → `*Computed[T]` with automatic dependency tracking
- `sumi.Effect(fn)` → runs when dependencies change
- Props are `*Signal[T]` passed between components
- Fine-grained: only recompute/re-render what actually depends on changed signal
- Script block is valid Go — compiler only transforms template expressions
- `runtime/signal/` is a standalone Go package with no sumi dependencies

## Iteration slices

### Slice 1: Signal runtime library (no integration yet)

Build `runtime/signal/` as a standalone package with tests.

- `Signal[T]` struct with `Get()`, `Set()`, `Update()`
- `Computed[T]` struct with `Get()`, auto-tracked dependencies
- `Effect(fn)` — runs fn, tracks `.Get()` calls, re-runs when those signals change
- Tracking context (goroutine-local or explicit scope)
- `Batch(fn)` — defer notifications until batch completes
- Tests: basic get/set, computed auto-tracks, effect fires, diamond dependency, batch

**Deliverable:** `go test ./runtime/signal/` passes, package is independently usable.

### Slice 2: Signal-driven rendering proof of concept

Wire signals into the app lifecycle for one hand-written example (no codegen changes yet).

- Write a counter example in pure Go using `signal.New()`, `signal.From()`, and the existing `layout.Input` tree
- Signal changes trigger `app.Wake()` → re-render
- Verify the render loop works with signal-driven state

**Deliverable:** A working counter app using signals, manually written (no .sumi file).

### Slice 3: Template compiler — signal unwrapping

Update the template compiler to auto-unwrap signals in expressions.

- Parser/codegen identifies signal variables (heuristic: assigned from `sumi.New()` or `sumi.From()`)
- Template `{count}` generates `count.Get()` instead of just `count`
- Template `{count + 1}` generates `count.Get() + 1`
- Existing non-signal expressions still work unchanged

**Deliverable:** A `.sumi` file using `sumi.New()` in script compiles and runs correctly.

### Slice 4: Migrate codegen — remove dirty/sync, emit signal-aware code

Update the code generator to use signals instead of dirty/sync.

- Remove `dirty` variable generation
- Remove `sync()` function generation
- Remove assignment rewriting (`count = x` → `count = x; dirty = true`)
- Generated render function subscribes to signals used in the layout tree
- Signal changes trigger re-layout of affected subtrees (initially: full re-layout via effect)

**Deliverable:** Generated code uses signals. Existing examples compile and work.

### Slice 5: Migrate $state → sumi.New()

Update all existing `.sumi` files and examples.

- `count := $state(0)` → `count := sumi.New(0)`
- `count = count + 1` → `count.Set(count.Get() + 1)` or `count.Update(...)`
- Update script parser to stop recognizing `$state`
- Update all test assertions

**Deliverable:** No `$state` references remain. All tests pass.

### Slice 6: Migrate $derived → sumi.From()

- `doubled := $derived(count * 2)` → `doubled := sumi.From(func() int { return count.Get() * 2 })`
- Remove `$derived` from parser
- Update codegen (no more derived variable extraction in sync)

**Deliverable:** No `$derived` references remain.

### Slice 7: Migrate $prop → signal-based props

- Props become `*Signal[T]` created by parent, passed to child
- Component inlining passes signal references instead of namespaced variables
- `bind:value` becomes sharing the same `*Signal[T]` between parent and child
- Remove `$prop` from parser

**Deliverable:** Component composition works with signal-based props.

### Slice 8: Migrate $env → sumi.Env()

- `width := $env(width)` → `width := sumi.Env[int]("width")`
- Framework creates env signals at app startup, updates them on SIGWINCH
- Components subscribe to env signals like any other signal

**Deliverable:** Responsive examples work with signal-based env.

### Slice 9: Fine-grained re-rendering

- Instead of full re-layout on any signal change, track which layout nodes depend on which signals
- Template compiler generates per-node signal subscriptions
- Only re-layout subtrees whose signals changed
- Measure performance improvement on complex examples

**Deliverable:** Partial re-layout working, measurable improvement on large UIs.

### Slice 10: Cleanup

- Remove old `$state`/`$derived`/`$prop`/`$env`/`$scroll` parsing from script parser
- Remove old codegen_reactive.go (dirty/sync generation)
- Update DESIGN.md iteration plan status
- Update all documentation and examples

**Deliverable:** Clean codebase with no legacy reactive code.

## Risk: this is a large migration

The current codebase has ~23 completed phases built on the compile-time model. The migration touches:
- Script parser (`parser/script/`)
- Code generator (`codegen/` — most files)
- All `.sumi` example files
- All scenario tests
- The preview tool

Mitigation: slices 1-2 are additive (no breaking changes). Slices 3-4 can coexist with old code temporarily. Slices 5-8 are mechanical find-and-replace within each primitive. The compiler can support both models briefly during transition.
