# scrollbar

A scroll position indicator with click-to-jump, drag-to-scroll, and animated transitions. Works in horizontal or vertical orientation. Used by `<sumi:TextInput>`, scroll containers, and any component that needs to visualise a viewport into larger content.

Replaces the old built-in scrollbar rendering (`runtime/render/scrollbar.go`, `runtime/layout/scrollbar_hittest.go`).

**Tier**: fundamental (lowercase, always available)

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `contentSize` | `int` | `0` | Total size of the scrollable content. |
| `viewSize` | `int` | `0` | Visible viewport size. |
| `offset` | `int` | `0` | Current scroll position. Use `bind:offset` for two-way binding. |
| `direction` | `string` | `"horizontal"` | `"horizontal"` or `"vertical"`. |
| `visible` | `bool` | `true` | Whether the scrollbar is rendered. When false, renders nothing. |

## Behavior

### Rendering

Renders a track with a proportionally-sized thumb:

```
< ----##---------- >
  ^    ^^           ^
  |    thumb        |
  arrows            arrows
```

- Track character: `-` (horizontal) or `|` (vertical)
- Thumb character: `#`
- Arrow characters: `<`/`>` (horizontal) or `^`/`v` (vertical)
- Thumb size: proportional to `viewSize / contentSize`, minimum 1 character
- Thumb position: proportional to `offset / (contentSize - viewSize)`
- All rendered with `dim` style

### Click

Clicking on the track jumps the scroll position to the clicked location with an animated ease-out-cubic transition (200ms).

### Drag

After clicking the track, holding and dragging moves the thumb to follow the mouse. The transition from click animation to drag is seamless — the animation completes, then dragging begins.

### Animation

Uses `app.RequestFrame()` for smooth ease-out-cubic scrolling on click. Each frame interpolates between the start and target offset.

### Hidden

When `visible` is false or `contentSize <= viewSize`, renders empty strings (zero-width). The component occupies no space.

## Usage

```html
<scrollbar
    content-size={len(items)}
    view-size={visibleCount}
    bind:offset={scrollTop}
    direction="vertical"
/>
```

```html
<scrollbar
    content-size={len(value)}
    view-size={contentW}
    bind:offset={viewOffset}
    visible={focused && len(value) > contentW}
/>
```
