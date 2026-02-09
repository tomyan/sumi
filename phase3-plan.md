# Phase 3: Reactive State — Elephant Carpaccio Breakdown

Goal: `$state` reactive variables, template expressions `{count}`, keyboard input, and a counter that updates in place.

Target demo:
```html
<script>
count := $state(0)

func increment() {
    count = count + 1
}
</script>

<box border="single" padding="1" onkey="increment">
    <text>Count: {count}</text>
    <text>Press any key to increment, q to quit</text>
</box>
```

## Slice 3.1: Template expressions — `{expr}` in text content
- Extend template parser to handle `{variable}` and `{expr}` inside `<text>` elements
- TextElement changes from `Content string` to `Parts []Part`
- Part is either `StringPart{Value string}` or `ExprPart{Expr string}`
- `<text>Count: {count}</text>` → Parts: [StringPart{"Count: "}, ExprPart{"count"}]
- `<text>Hello</text>` (no expressions) → Parts: [StringPart{"Hello"}]
- Nested braces in expressions: `{items[0]}` — for now keep it simple, no nested braces
- TDD: static text (no expr), single expr, mixed text+expr, multiple exprs, expr at start/end

## Slice 3.2: Script block parser — `$state` + functions
- New `parser/script` package
- Parse the script block content (string from section splitter)
- Identify `name := $state(initialExpr)` → `StateDecl{Name, InitExpr string}`
- Identify `func name(...) { body }` → `FuncDecl{Name, Params, Body string}`
- Track which variables are state vars
- Find assignments to state vars in function bodies → mark as needing invalidation rewrite
- Approach: preprocess `$state(x)` → placeholder, then use `go/ast` or simple text parsing
- TDD: single state var, multiple state vars, function with state assignment, function without state, empty script

## Slice 3.3: Raw terminal keyboard input
- New `runtime/input` package
- `EnableRawMode(fd int) (restore func(), err error)` — set terminal to raw/cbreak mode
- `ReadKey(r io.Reader) (rune, error)` — read a single keypress
- Restore terminal on cleanup (important: even on panic)
- TDD: test with pipe (mock stdin), verify ReadKey returns correct runes

## Slice 3.4: Codegen — state, expressions, render loop, key handling
- Generated code structure:
  1. State variables declared as regular Go vars
  2. A `render()` closure that rebuilds layout tree using current state values, diffs buffers, and outputs changes
  3. Template expressions: `{count}` → `fmt.Sprintf("%v", count)` in the text content
  4. Event loop: enable raw mode → render → read key loop → call onkey handler → re-render → on 'q' exit
  5. Buffer diffing: keep previous buffer, only write changed cells
- `onkey` attribute on box → maps to function call in event loop
- Codegen uses script AST (state decls + func decls) + template AST (with expressions)
- TDD: generate from state+template, verify valid Go, verify state var present, verify render func, verify event loop

## Slice 3.5: End-to-end — counter demo
- Integration test: `.sumi` file with script + template → generates valid, compilable Go
- The counter demo works when run

## Dependencies
```
Slice 3.1 (template exprs) ──────┐
Slice 3.2 (script parser) ───────┼→ Slice 3.4 (codegen) → Slice 3.5 (e2e)
Slice 3.3 (keyboard input) ──────┘
```

All three initial slices can run in parallel.
