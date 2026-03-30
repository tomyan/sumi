# :hover Pseudo-Class Design

## Goal

Elements with `:hover` CSS rules change appearance when the mouse is over them. This enables visual feedback for interactive elements (tabs, buttons, links) without any script code.

## Example

```css
.tab { dim: true; }
.tab:hover { dim: false; color: white; }
```

```sumi
<box class="tab"><text>Console</text></box>
```

When the mouse moves over the tab box, it undims and turns white. When the mouse leaves, it returns to dim.

## Architecture

### CSS Parser

The style parser already handles `.class { props }`. Extend it to handle `.class:hover { props }`.

A rule with `:hover` produces two entries:
- `.tab` → base properties
- `.tab:hover` → hover properties

The `:hover` suffix is stored as metadata on the rule, not mixed into the selector string.

```go
type Rule struct {
    Selector   string
    Pseudo     string            // "hover", "" for normal
    Properties map[string]string
}
```

### CSS Resolver

`css.Resolve()` currently returns one merged property map. It needs to return both base and hover properties:

```go
type ResolvedStyle struct {
    Base  map[string]string
    Hover map[string]string  // nil if no :hover rules match
}
```

Or simpler: two separate calls. The codegen resolves base and hover independently.

### Layout Input

Add a `HoverStyle` field to `Input`:

```go
type Input struct {
    ...
    Style      render.Style
    HoverStyle render.Style // applied when mouse is over this node
    Hovered    bool         // set by the framework before render
}
```

### Mouse Tracking

The app tracks the current mouse position. Before each render:

1. Read mouse position from the latest `EventMouse`
2. Walk the box tree from the previous layout result
3. Hit-test to find which boxes contain the mouse position
4. Set `Hovered = true` on matching Input nodes (and their ancestors up to the first hover-styled box)
5. Render — `renderContent` uses `HoverStyle` when `Hovered` is true

### Render

In `renderTreeWithInherit`, when `box.Hovered` is true and `box.HoverStyle` is non-zero, use `HoverStyle` instead of `Style` for that node's rendering and inheritance.

### Codegen

`writeStyleLiteral` emits `HoverStyle` when hover properties exist:

```go
if hoverProps != nil {
    writeStyleLiteral(buf, tabs, "HoverStyle", hoverProps)
}
```

### Hit Testing

Sumi already has `HitTestScroll` for scroll containers. Mouse hover hit-testing is simpler: walk the box tree, find all boxes whose bounds contain the mouse position. Set `Hovered` on them.

The hit test runs against the *previous* layout result (which is the box tree from the last render). This is fine — hover is visual feedback, one frame of latency is imperceptible.

## Implementation Slices

1. **CSS parser**: parse `:hover` pseudo-class on selectors, store in `Rule.Pseudo`
2. **CSS resolver**: resolve hover properties separately from base
3. **Layout Input**: add `HoverStyle` and `Hovered` fields
4. **Codegen**: emit `HoverStyle` from hover CSS rules
5. **Mouse tracking**: track position in app, hit-test before render, set `Hovered`
6. **Render**: use `HoverStyle` when `Hovered` is true
7. **Cursor shape**: emit OSC 22 for pointer cursor on hoverable elements (stretch)

## What Inherits on Hover

When a box is hovered, its `HoverStyle` REPLACES its `Style` for rendering (not merges). This means `:hover` is a complete style override for that node. Children inherit from the hover style when the parent is hovered.

This matches CSS: `.tab:hover { dim: false; color: white; }` means the tab is NOT dim and IS white when hovered, regardless of what `.tab` says.

## Mouse Enable

Mouse tracking requires enabling SGR mouse mode. The app already has `HasMouse bool`. For hover support, mouse mode needs to be enabled. The `RunWithOptions` could auto-enable mouse when any component has hover styles, or it could be a global setting.

Simplest: enable mouse mode when the component tree contains any `HoverStyle`.
