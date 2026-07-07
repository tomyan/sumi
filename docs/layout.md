# Layout

Sumi lays out an element tree onto a terminal cell grid. The layout engine reads
resolved CSS into geometry, so "what layout supports" is "what the resolver reads
and the layout code acts on". This chapter documents exactly that.

## Units

All lengths are in terminal cells. A bare integer is a count of cells; the units
`cell` and `ch` are accepted and mean the same thing (`1ch` = one cell). Other
CSS units (`px`, `em`, `rem`, `vh`, …) parse to `0` — they are silently dropped,
not approximated. Percentages are supported only where noted below (width,
height, flex-basis, grid tracks).

## display

The `display` value chooses the layout algorithm:

- `block` — children stack vertically; the box fills the available width.
  Inline children form an inline formatting context. Vertical margins between
  block siblings collapse.
- `flex` — a flex container: children lay out along a main axis (see
  [Flex](#flex)). HTML containers default to `block` via the UA
  stylesheet, so flex is always an explicit opt-in.
- `grid` — two-dimensional track grid (see [Grid](#grid)).
- `table` — row/column table (see [Tables](#tables)).
- `inline` — the element joins the surrounding inline flow rather than forming
  its own block.
- `inline-block` — the element lays out as a block but is placed as an
  unbreakable atom on an inline line, top-aligned.
- `contents` — the element generates no box; its children are lifted into the
  parent's flow.
- `none` — removed from layout entirely; it takes no space and cannot be
  focused, but keeps its slot so sibling indices stay stable.

Deviation: an unrecognised `display` value falls through to the flex
container path rather than being ignored. Deviation:
`display: table-row` and `display: table-cell` have no effect; table structure
comes from element nesting (`tr`, `td`/`th`), not from `display`.

## Block flow and inline formatting

In a block container, block-level children stack top to bottom. A run of
inline-level content (text, `<span>`, `<strong>`, inline-blocks) between blocks
forms an inline formatting context and wraps as lines.

Text wraps across inline element boundaries: `foo<strong>bar</strong>` is one
unbreakable word. Whitespace is collapsed under the default `white-space: normal`
— any run of spaces, tabs, and newlines becomes a single space, and leading and
trailing whitespace on a line is dropped. Lines soft-break at collapsed spaces; a
word wider than the line hard-breaks at the width. `text-align: center` and
`text-align: right` shift whole lines within the available width.

Margin collapse: adjacent block siblings collapse their touching vertical margins
to the larger of the two. Only positive margins collapse, and only in a `block`
container — the default flex column path sums margins without collapsing. Inline
content between two blocks resets the collapse.

`white-space: pre` preserves spaces and newlines exactly (used by `<pre>` and by
`<textarea>`).

## Flex

Flex is the default layout. The main axis is set by `flex-direction`:

- `flex-direction` — `column` (default), `row`, `column-reverse`, `row-reverse`.
- `gap` — cells between children. There is no `row-gap`/`column-gap`.
- `flex-grow` — integer factor; free space is distributed in proportion, with any
  rounding remainder going to the first growing child.
- `flex-shrink` — children shrink proportionally (weighted by shrink × size) when
  they overflow. `flex-shrink: 0` opts out.
- `flex-basis` — cells or a percentage of the main-axis size. `auto` means unset.
- `flex` shorthand — `flex: none`, `flex: <grow>`, or `flex: <grow> <shrink>
  <basis>`.
- `flex-wrap: wrap` — wraps a **row** container onto multiple lines.
- `justify-content` — `start` (default), `end`, `center`, `space-between`,
  `space-around`, `space-evenly` (`flex-start`/`flex-end` are accepted aliases).
- `align-items` / `align-self` — `stretch` (default), `start`, `end`, `center`.
- `order` — integer; lower orders lay out first, ties keep source order.

```sumi
<style>
.toolbar {
	display: flex;
	flex-direction: row;
	gap: 1;
	justify-content: space-between;
	align-items: center;
}
</style>
```

Cross-axis `auto` margins centre a flex child. Deviation: `flex-wrap` only wraps
row containers (no column wrap), and grow, `justify-content`, and `align-items`
are not applied to wrapped rows. Deviation: main-axis `auto` margins are not
implemented. On a column container, `justify-content` applies only when the
container has a fixed height.

## Grid

Set `display: grid` and define tracks:

- `grid-template-columns` / `grid-template-rows` — a list of track sizes.
- Track units: `fr` (fraction of free space), `%` (of the axis), cells, and
  `auto` (treated as `1fr`).
- `repeat(n, …)` expands inline; `minmax(min, max)` is accepted.
- `grid-template-areas` — quoted rows of area names; `.` is an empty cell.
  Reference an area with `grid-area`.
- `grid-column` / `grid-row` — `"2"` (a single 1-based line), `"1 / 3"` (start /
  end, end exclusive), or `"span N"`.
- `gap` — cells between tracks.

```sumi
<style>
.grid {
	display: grid;
	grid-template-columns: repeat(3, 1fr);
	gap: 1;
}
</style>
```

Items that are not explicitly sized stretch to fill their grid area. Unplaced
items auto-flow row by row into free cells. Deviation: `grid-auto-flow: column` is
not implemented (auto-flow is always row-major). Deviation: `minmax(min, max)`
sizes the track at `max` and enforces `min` as a floor without redistributing
space.

## Tables

`<table>` lays out rows and columns. `<thead>`, `<tbody>`, and `<tfoot>` are
flattened into their rows. A `<caption>` child renders centred above the table.

- `colspan` / `rowspan` — span a cell across columns/rows.
- `border-spacing` — `"h v"` or a single value; cells between table cells (UA
  default `2 0`).
- `border-collapse: collapse` — overlaps adjacent cell borders by one cell and
  ignores `border-spacing`, drawing junction characters where borders meet.
- `table-layout: fixed` — column widths come from the first row only; later rows
  never widen a column. The default (auto) sizes each column to its widest cell.
- `<colgroup>`/`<col>` — a `col` with a fixed width overrides that column's width.
- `empty-cells: hide` — clears the borders of cells with no content.

Deviation: spanning cells do not contribute to automatic column-width sizing.

## Box model

- `width` / `height` — cells, a percentage of the containing block, or a
  `calc()` expression. An unset size means auto (content- or flex-sized).
- `min-width` / `max-width` / `min-height` / `max-height` — cells (percentages on
  these are dropped). On a scroll container, `min-width` sets the minimum
  **content** width that drives horizontal scrolling.
- `padding` — the `padding` shorthand with 1, 2, or 4 values (cells).
- `margin` — the `margin` shorthand (1/2/4 values) plus the per-side longhands
  `margin-top`/`right`/`bottom`/`left`. `auto` on both cross-axis sides centres
  the box; on a block child, horizontal `auto` margins centre it.
- `box-sizing` — `border-box` (default) or `content-box`.

```sumi
<style>
.card {
	width: 40;
	max-width: 100%;
	padding: 1 2;
	margin: 0 auto;
	box-sizing: border-box;
}
</style>
```

### Overflow and scrolling

`overflow` takes one value:

- unset — visible; content is not clipped.
- `hidden` — clips to the box; no scrollbar, no scrolling.
- `scroll` — clips and always shows a scrollbar.
- `auto` — clips and shows a scrollbar only when content overflows.

Scroll containers track a scroll offset, clamp it to the content, and can follow
new content to the bottom. Both vertical and horizontal scrollbars are drawn, and
the clip narrows to make room for them.

## Position

`position` values: static (unset), `relative`, `absolute`, `fixed`, `sticky`.
Offsets `top`/`right`/`bottom`/`left` and `z-index` are integer cells.

- `relative` — stays in flow (does not move siblings) and is shifted visually by
  its offsets. `top` wins over `bottom`, `left` over `right`.
- `absolute` — removed from flow and positioned within the parent's content area.
  With both opposing offsets set and no fixed size, the element stretches to fill
  the gap.
- `fixed` — removed from flow and positioned relative to the viewport; it escapes
  ancestor scroll offsets and clipping.
- `sticky` — stays in flow and contributes to size, then clamps at the top of its
  scroll container while scrolled past. Deviation: sticky clamps vertically only,
  and only inside a scroll container.

`z-index` sets paint order: children are painted in ascending `z-index`, ties
keeping document order.

```sumi
<style>
dialog {
	position: absolute;
	top: 2;
	left: 4;
	z-index: 10;
}
</style>
```
