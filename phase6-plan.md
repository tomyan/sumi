# Phase 6: Flexbox Layout — Elephant Carpaccio Breakdown

Goal: Horizontal layout (`direction: row`), spacing (`gap`), flexible sizing (`flex-grow`/`flex-shrink`), and alignment (`justify`/`align`). Available width/height must flow through the layout tree.

Target demo:
```html
<style>
.row { direction: row; gap: 2; }
.stretch { flex-grow: 1; }
.centered { justify: center; align: center; }
</style>

<box class="row">
    <box class="stretch" border="single">
        <text>Left panel</text>
    </box>
    <box class="stretch" border="single">
        <text>Right panel</text>
    </box>
</box>
```

## Current State

- Layout engine only does column (vertical stacking)
- `Direction` field exists on Input but "row" is not implemented
- `availWidth`/`availHeight` params to `Layout()` are passed but **unused**
- No fields for: gap, justify, align, flex-grow, flex-shrink

## Slice 6.1: Row direction (horizontal stacking)

The foundation — children placed side by side instead of stacked vertically.

- Update `layoutNode()` to branch on `Direction == "row"`
- Row layout: children placed left-to-right, `cursorX` advances by child width
- Parent auto-width = sum(child widths) + padding + border
- Parent auto-height = max(child heights) + padding + border
- Codegen already emits `Direction: "row"` — no codegen changes needed
- TDD: single row child, two row children, nested row in column, row with padding+border

## Slice 6.2: Available width/height propagation

Layout() must pass available dimensions through the tree so flex-grow, percentage widths, and justify can work.

- `layoutNode()` → `layoutNode(input, availW, availH)`
- Column children get `availW - padding - border` as their available width
- Row children get remaining width after fixed-size siblings
- Auto-sized boxes constrained to available dimensions
- TDD: box fills available width, nested box propagation, text truncation to available width

## Slice 6.3: Gap (space between children)

- Add `Gap int` field to `Input`
- Column: add gap between children vertically (not before first or after last)
- Row: add gap between children horizontally
- Gap included in auto-size calculation
- Codegen: emit `Gap: N` from `gap` attribute/CSS property
- TDD: column gap, row gap, gap with single child (no gap), gap in auto-size

## Slice 6.4: Flex-grow (distribute extra space)

- Add `FlexGrow int` field to `Input` (default 0 = no grow)
- After placing fixed-size children, distribute remaining space proportionally among flex-grow children
- Works for both row and column directions
- Codegen: emit `FlexGrow: N` from `flex-grow` attribute/CSS property
- TDD: one grow child fills space, two equal grow, unequal grow ratios, grow in row, grow in column

## Slice 6.5: Justify-content (main axis alignment)

- Add `Justify string` field to `Input`
- Values: `start` (default), `end`, `center`, `space-between`, `space-around`, `space-evenly`
- Column: distributes vertical space among children
- Row: distributes horizontal space among children
- Codegen: emit `Justify: "center"` from `justify` attribute/CSS property
- TDD: each justify value in column, each in row, justify with one child

## Slice 6.6: Align-items (cross axis alignment)

- Add `Align string` field to `Input`
- Values: `start` (default), `end`, `center`, `stretch`
- Column: aligns children horizontally within the box
- Row: aligns children vertically within the box
- `stretch`: child expands to fill cross-axis dimension
- Codegen: emit `Align: "center"` from `align` attribute/CSS property
- TDD: each align value in column, each in row, stretch behavior

## Slice 6.7: E2E — flexbox demo

- Create `examples/flexbox/` with a dashboard-style layout
- Two-column layout with flex-grow panels
- Centered header, spaced-between footer
- Verify: compiles, runs, looks correct in terminal

## Dependencies
```
Slice 6.1 (row) → Slice 6.2 (avail dims) → Slice 6.3 (gap)
                                           → Slice 6.4 (flex-grow)
                                           → Slice 6.5 (justify)
                                           → Slice 6.6 (align)
                                                                  → Slice 6.7 (E2E)
```

6.1 first, then 6.2, then 6.3-6.6 can be done in any order (all depend on available width). 6.7 last.
