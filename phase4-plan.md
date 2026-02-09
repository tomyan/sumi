# Phase 4: Style Block — Elephant Carpaccio Breakdown

Goal: `<style>` block parsing, class selectors, colors (fg/bg), text attributes (bold/dim/italic/underline), scoped to component. Styled text and borders rendered with ANSI escape codes.

Target demo:
```html
<style>
.container {
    border: single;
    padding: 1 2;
}
.title {
    color: green;
    bold: true;
}
.subtitle {
    color: cyan;
    dim: true;
}
.count {
    color: yellow;
    bold: true;
}
</style>

<script>
count := $state(0)
func increment() {
    count = count + 1
}
</script>

<box class="container" onkey="increment">
    <text class="title">Sumi Counter</text>
    <text class="subtitle">Press any key to increment, q to quit</text>
    <text class="count">Count: {count}</text>
</box>
```

## Slice 4.1: Style parser — CSS-like syntax for terminal properties
- New `parser/style` package
- Parse `<style>` block content into a stylesheet AST
- Selectors: `.class` and element names (`text`, `box`)
- Properties: `color`, `background`, `bold`, `dim`, `italic`, `underline`, `strikethrough`, `inverse`, `border`, `padding`, `direction`
- Color values: ANSI names (red, green, cyan, yellow, blue, magenta, white, black)
- Boolean values: `true`/`false` for bold, dim, etc.
- Output: `Stylesheet` containing list of `Rule{Selector, Properties map[string]string}`
- TDD

## Slice 4.2: Style type in render — Cell gets a Style field
- Extend `Cell` struct with a `Style` field
- `Style` struct: `FG, BG Color`, `Bold, Dim, Italic, Underline, Strikethrough, Inverse bool`
- `Color` type: ANSI color support (named colors → ANSI codes)
- Extend `SetCell` and `WriteText` to accept style
- `SetStyledCell(row, col int, ch rune, style Style)`
- `WriteStyledText(row, col int, text string, style Style)`
- Extend `RenderTo` to emit ANSI SGR sequences for styled cells
- TDD

## Slice 4.3: Style resolution — match class selectors to elements
- New `runtime/css` package (or extend style parser)
- `Resolve(stylesheet, element) → Style` — given an element's tag and classes, find matching rules and merge properties into a Style
- `class` attribute on template elements: `<text class="title">` / `<box class="container">`
- Template parser already handles attributes on `<box>`. Extend `<text>` to support `class` attribute too.
- Merge: later rules override earlier ones (flat cascade, no specificity)
- TDD

## Slice 4.4: Layout + codegen — wire style through layout to render
- Add `Style` field to `layout.Input` and `layout.Box`
- Codegen: parse stylesheet, resolve styles per element, pass to layout Input
- `renderTree` uses styled rendering: `WriteStyledText` instead of `WriteText`, border color from style
- Border color: `DrawBorder` gets optional style for border color
- TDD + update existing tests

## Slice 4.5: End-to-end — styled counter demo
- Integration test: .sumi with style+script+template → generates valid Go
- Build and verify styled output
- Update counter example with colors

## Dependencies
```
Slice 4.1 (style parser) ────────────────┐
Slice 4.2 (Cell + Style rendering) ──────┼→ Slice 4.4 (codegen wiring) → Slice 4.5 (e2e)
Slice 4.3 (style resolution + class) ────┘
```

4.1, 4.2, 4.3 can run in parallel.
