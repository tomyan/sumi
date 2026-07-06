# CSS on a cell grid

Sumi styles components with a subset of standard CSS, but the render target is
a grid of terminal cells, not a plane of pixels. This chapter describes how
each family of CSS properties maps onto that grid, and what happens to
declarations that have no meaning on a terminal. The guiding rule is graceful
degradation: anything sumi cannot honour drops silently, leaving the rest of
the rule intact.

## The cell is the unit

The only length unit with meaning is the cell. `1cell` is one terminal cell,
and `1ch` is an exact alias for it (`1ch = 1cell`) — not "the width of the 0
glyph" as on the web, simply one cell. A bare integer means cells too, so
`width: 20`, `width: 20cell` and `width: 20ch` are equivalent.

Lengths are whole numbers. A cell cannot be subdivided, so `1.5cell` is not a
valid length and drops to zero. Any pixel-derived unit — `px`, `pt`, `em`,
`rem`, `vh`, `vw` — has no cell equivalent and also resolves to zero: an
unrecognised unit yields nothing rather than an error.

Percentages are resolved per property against the containing block, and
properties that have no sensible percentage meaning ignore them. `calc()`,
`min()`, `max()` and `clamp()` are supported and evaluate to whole cells
(rounded to nearest), mixing percentages and cell counts:

```css
width: calc(100% - 10);
min-width: min(50%, 30);
```

Only `%`, `cell` and `ch` are recognised inside `calc()`; any other unit fails
the expression.

## Colours

Colour values accept the common CSS forms:

- The eight ANSI keywords `black red green yellow blue magenta cyan white`.
  These stay as *palette names*, not fixed RGB, so they follow the terminal's
  own theme.
- The full CSS named-colour set (`tomato`, `rebeccapurple`, `gray`/`grey`, and
  so on), which resolve to fixed RGB.
- Hex: `#rgb`, `#rgba`, `#rrggbb`, `#rrggbbaa`.
- Functional forms: `rgb() rgba() hsl() hsla() hwb() lab() lch() oklab()
  oklch()`, in both legacy comma syntax and modern space-plus-`/alpha` syntax.
- `transparent`, meaning "no colour set".
- `light-dark(<light>, <dark>)`, which keeps both arms and picks one at emission
  time according to the active colour scheme.

```css
color: cyan;                        /* themeable palette name */
color: rgb(255 85 0 / 0.5);         /* modern syntax with alpha */
color: lab(54.29 80.81 69.89);      /* parsed via CIELAB D50 */
color: light-dark(#fff, rgb(0 0 0));
```

`lab()`, `lch()`, `oklab()` and `oklch()` are converted to sRGB when parsed
(CIELAB uses the D50 reference white, per the CSS spec). This Lab maths is used
only to *parse* those literals — colour interpolation for transitions happens
in plain sRGB, not in Lab space.

## Colour depth

RGB colours are quantised at emission time to the terminal's detected depth:

| Depth | Meaning |
|---|---|
| Truecolor | 24-bit RGB emitted verbatim (the default) |
| 256 | Nearest xterm-256 palette entry (6×6×6 cube or grayscale ramp) |
| 16 | Nearest of the eight ANSI colours by RGB distance |
| Mono | No colour at all; attributes only |

Depth is detected from the environment: `NO_COLOR` forces mono; `COLORTERM` of
`truecolor`/`24bit` forces truecolor; `TERM` containing `256color` gives 256;
`TERM=dumb` gives mono; otherwise sumi assumes 16. Because quantisation runs at
emission, a component describes colour at full fidelity and sumi narrows it to
fit — named ANSI colours pass through every depth untouched.

## Opacity and alpha

`opacity` has two behaviours. `opacity: dim` sets the terminal's dim (SGR 2)
attribute. A numeric `opacity` below 1 makes the element's RGB foreground and
background translucent so they composite over whatever is behind them:

```css
.hint  { opacity: dim; }   /* SGR dim attribute */
.ghost { opacity: 0.5; }   /* alpha blend, if colours are RGB */
```

Numeric opacity only blends RGB colours. If neither foreground nor background
is RGB — for example, an ANSI palette name that cannot be blended — numeric
opacity falls back to the dim attribute. Compositing blends source over the
backdrop cell (`src·a + dst·(1−a)` per channel); if either side is a non-RGB
colour the source simply paints opaque.

## Borders as box-drawing characters

`border` takes a style keyword; `border-color` is a separate declaration
(`border` is not the web's `width style color` shorthand). The built-in styles
map to Unicode box-drawing characters:

| `border` | Corners and lines |
|---|---|
| `single` | `┌┐└┘ ─ │` |
| `double` | `╔╗╚╝ ═ ║` |
| `rounded` | `╭╮╰╯ ─ │` |
| `heavy` | `┏┓┗┛ ━ ┃` |
| `ascii` | `+ - \|` (portable fallback) |

An unknown style name falls back to `single`; `none` (or an empty value) draws
nothing, and `border-top`/`border-bottom` set individual edges.

```css
.panel { border: single; border-color: cyan; }
```

### Collapsed borders and junctions

`border-collapse: collapse` on a container makes adjacent bordered children
share an edge instead of drawing two parallel lines. Where edges meet, sumi
merges the box-drawing characters into the correct junction — reading the glyph
already on the cell, OR-ing in the new connections, and rewriting it as `┬`,
`├`, `┼`, and so on. This produces tmux-style seamless panel grids.

```css
.layout { display: flex; flex-direction: row; border-collapse: collapse; }
.panel  { border: single; flex-grow: 1; }
```

### Border titles

A bordered box can carry a title drawn into its top edge, set either as a
`border-title` CSS property or, more commonly, as an attribute on the element:

```html
<div class="panel" border-title="Panel 1">…</div>
```

The title renders as `┌─ Panel 1 ───┐`, starting three cells in, and truncates
if it is wider than the box.

## Text attributes map to SGR

Typographic properties become SGR (Select Graphic Rendition) attributes on the
cell:

| CSS | SGR attribute |
|---|---|
| `font-weight: bold` / `bolder` / weight ≥ 700 | bold (1) |
| `opacity: dim` | dim (2) |
| `font-style: italic` / `oblique` | italic (3) |
| `text-decoration: underline` | underline (4) |
| `text-decoration: line-through` | strikethrough (9) |

`text-decoration` recognises only `underline` and `line-through`; other values
are ignored. Font-weight, font-style and the decoration flags inherit
independently, so nested `<em><strong>` composes italic and bold. Sumi also
exposes a non-standard `inverse: true` declaration (SGR 7).

## What has no meaning here

The following drop harmlessly, leaving the surrounding rule valid:

- Pixel-derived length units (`px`, `pt`, `em`, `rem`, `vh`…) resolve to zero.
- Percentages on properties with no percentage meaning are ignored.
- Unknown properties and invalid values (`color: bogus(1)`) are skipped.
- Unresolved `var()` references without a fallback expand to nothing, dropping
  the property.
- Media features a terminal cannot probe are driven by runtime configuration
  rather than hardware: `prefers-color-scheme` follows the resolved scheme
  (dark by default), `display-mode` is always `terminal`, and
  `prefers-reduced-motion` is exposed as a signal (see [motion](motion.md)).

For the escape sequences behind all of this and the per-terminal support
matrix, see [terminal support](terminals.md).
