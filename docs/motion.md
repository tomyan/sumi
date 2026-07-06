# Motion

Sumi animates the cell grid with the same two CSS mechanisms as the web:
`transition` for state changes and `@keyframes` for scripted animations. Both
are driven by a frame loop that ticks roughly every 16ms while anything is
moving, and both resolve `var()` and `light-dark()` per node so animated values
follow the element's own custom-property scope and colour scheme.

## Transitions

A transition interpolates a property when its value changes. The shorthand is
`property duration [timing-function] [delay]`, and it needs at least a property
and a duration — a bare `transition: color` with no duration is dropped.

```css
.button {
  transition: color 200ms ease-out, background 500ms ease-in;
}
```

Durations accept `ms` (`200ms`), seconds (`0.5s`, `2s`), or a bare number read
as milliseconds. Multiple transitions are comma-separated in the shorthand. The
longhand properties (`transition-property`, `transition-duration`,
`transition-timing-function`, `transition-delay`) are also honoured, but the
longhand form produces only a single transition — use the shorthand for a list.
When no timing function is given, the default is `ease`.

Three property names interpolate as styles:

- `color` — interpolates the foreground colour.
- `background` — interpolates the background colour.
- `all` — interpolates the whole style. Colours blend continuously; boolean
  attributes (bold, italic, underline, and so on) snap at the halfway point.

Colour interpolation runs in plain sRGB, blending each channel linearly. There
is no `opacity` transition — width and height animate through a separate length
path described below.

## Easing

Named timing functions map to the standard cubic-bezier control points:

| Name | Control points |
|---|---|
| `linear` | 0, 0, 1, 1 |
| `ease` | 0.25, 0.1, 0.25, 1.0 |
| `ease-in` | 0.42, 0, 1.0, 1.0 |
| `ease-out` | 0, 0, 0.58, 1.0 |
| `ease-in-out` | 0.42, 0, 0.58, 1.0 |

`cubic-bezier(x1, y1, x2, y2)` takes exactly four numbers. Step functions are
supported as `step-start`, `step-end`, `steps(n)` and `steps(n, <position>)`,
where `n` must be at least 1 and `<position>` is one of `start`/`jump-start`
(step at the beginning) or `end`/`jump-end` (step at the end, the default):

```css
transition: color 1s steps(4, end);
```

Note that only these two step positions are implemented — the CSS
`jump-both` and `jump-none` keywords are not supported and will fail to parse.

## Length transitions

Width and height animate too, but as whole-cell integers rather than colours.
A `transition` covering `width`, `height` or `all` interpolates the fixed
(CSS-stamped) length toward its new target, rounding to whole cells each frame:

```css
.drawer { width: 20; transition: width 300ms ease-out; }
.drawer.open { width: 40; }
```

Only fixed lengths animate; percentage, `calc()` and flex-driven sizes do not,
because the resolver re-stamps their target every layout pass. Length
transitions support retargeting mid-flight — if the target changes while a
transition is running, the new transition starts from the currently displayed
value rather than snapping. Delay is honoured, holding the start value until it
elapses.

## Keyframe animations

`@keyframes` blocks define named animations with `from`/`to` or percentage
stops; the `animation` shorthand attaches one to an element:

```css
@keyframes pulse {
  from { color: cyan; }
  50%  { color: white; }
  to   { color: cyan; }
}

.status { animation: pulse 2s infinite ease-in-out; }
```

The shorthand is `name duration [rest…]` and needs at least a name and a
duration. Remaining tokens are classified by shape, so order among them is
flexible. Supported values:

- **iteration-count** — an integer, or `infinite` (runs forever).
- **direction** — `normal`, `reverse`, `alternate`, `alternate-reverse`.
- **fill-mode** — `none`, `forwards`, `backwards`, `both`.
- **play-state** — `running`, `paused`.
- **delay** — a duration (`500ms`, `1s`).

Defaults are `ease`, one iteration, `normal`, `none`, `running`. The
`animation-*` longhand properties set the same values individually. `forwards`
(or `both`) fill holds the final stop after the last iteration; `backwards` (or
`both`) holds the first stop during the delay. Pausing freezes the animation
and resuming continues from where it stopped rather than restarting.

Each keyframe stop is resolved against the element's own context: `var()`
references expand from that node's custom-property scope, and `light-dark()`
picks the arm for the active scheme. This means the same `@keyframes` block
produces different concrete colours on different elements, exactly as on the
web.

## Reduced motion

Sumi exposes a reduced-motion signal but does not silently disable animation —
you decide what to suppress. The signal is on when either the
`RunOptions.ReducedMotion` field is set or the `SUMI_REDUCED_MOTION`
environment variable is present (any non-empty value counts, including `0`).

The signal surfaces only through the `prefers-reduced-motion` media query, so
gate motion yourself:

```css
@media (prefers-reduced-motion: reduce) {
  .status { animation: none; }
  .drawer { transition: none; }
}
```

Nothing else in the engine consults reduced motion — if you do not write a
media query, animations still run.

## The frame loop

Animation is pull-based. When a transition or keyframe animation is active, the
runtime schedules the next frame with `RequestFrame()`, which waits about 16ms,
marks the app dirty, and wakes the event loop to dispatch a frame tick. The
loop re-renders, and as long as any transition or animation is still running it
schedules another frame. When everything settles, the ticking stops — sumi does
not spin a fixed-rate render loop.

Time is read through a small clock abstraction measured in milliseconds. In
production this is the wall clock; tests inject a controllable clock whose
`Advance(ms)` steps time deterministically, so animation behaviour can be
asserted without real delays. See [testing](testing.md) for how that plays out
in practice.
