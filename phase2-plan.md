# Phase 2: Box Layout Basics — Elephant Carpaccio Breakdown

Goal: `<box>` elements with direction, padding, border → rendered as bordered boxes with text inside, stacked vertically.

## Slice 2.1: Template parser — `<box>` element with attributes + nesting
- Add `BoxElement` to AST: `Attributes map[string]string`, `Children []Node`
- Parse `<box direction="column"><text>Hello</text></box>`
- Parse attributes: `direction`, `width`, `height`, `padding`, `border`
- Support nesting (box containing text elements, box containing boxes)
- TDD: box with text child, box with attributes, nested boxes, self-closing box (empty), missing close tag error

## Slice 2.2: Layout engine — tree layout with column stacking
- New `runtime/layout` package
- `LayoutNode` struct: `X, Y, Width, Height int`, `Children []*LayoutNode`, `Tag string`, `Content string`
- `Layout(tree, availableWidth, availableHeight) *LayoutNode` — computes positions
- Column direction (default): children stack vertically, each child gets full width, height = content height or fixed
- Text nodes: height=1, width=len(content)
- Box with fixed width/height: use those values
- TDD: single text node, column of text nodes, box with fixed dimensions, nested boxes

## Slice 2.3: Layout engine — padding
- Padding offsets children within a box
- `padding="1"` → 1 cell all sides. `padding="1 2"` → 1 top/bottom, 2 left/right
- Box's content area = box size minus padding
- Children positioned relative to content area
- TDD: box with padding, children offset correctly, padding reduces available space

## Slice 2.4: Border rendering
- Extend `Buffer` with `DrawBorder(row, col, width, height int, style string)`
- Single border style: `─ │ ┌ ┐ └ ┘`
- Border occupies 1 cell on each side (inset from box bounds, like border-box)
- TDD: draw border on buffer, verify corner and edge characters

## Slice 2.5: Codegen — box support + layout integration
- Codegen walks AST, builds layout tree, generates code that:
  1. Constructs layout input from AST
  2. Runs layout to get positions
  3. Renders borders and text at computed positions
- Generated code imports `runtime/layout` in addition to `runtime/render`
- TDD: generate from box+text AST, verify valid Go, verify layout/render imports

## Slice 2.6: End-to-end — bordered boxes with text
- Integration test: `.sumi` file with boxes → `sumi generate` → `.go` file → compiles
- Verify generated code is correct and handles the full pipeline

## Dependencies
```
Slice 2.1 (parser) ─────────────────────────────────────┐
Slice 2.2 (layout column) → Slice 2.3 (padding) ────────┼→ Slice 2.5 (codegen) → Slice 2.6 (e2e)
Slice 2.4 (border render) ──────────────────────────────┘
```

Parallelizable: 2.1 + 2.2 + 2.4 can start simultaneously.
2.3 depends on 2.2. 2.5 depends on 2.1 + 2.3 + 2.4.
