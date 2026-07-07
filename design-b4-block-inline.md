# B4 design — block/inline flow + vertical margin collapse

Status: IMPLEMENTED 2026-07-05 (commits 6bcf73d..edfb18e; all slices
shipped; text-align inside an IFC deferred — see plan follow-up).
Reference findings from svelterm
(`~/projects/svelterm/src/layout/engine.ts`, `layout/text.ts`,
`render/paint.ts`, `render/paint-text.ts`).

## Key finding (changes the plan's framing)

The parity plan says B4 "needs a text-run model; inline runs through
wrapped text are the hard part". **Svelterm does not have one.** Its flow
model is deliberately simplified:

- No line boxes, no inline formatting context, no anonymous block boxes.
- "Inline" only means: advance a horizontal cursor instead of breaking to
  a new line. Each inline child (text node, `inline`/`inline-block`
  element) is measured **independently** and placed at `(cursorX, cursorY)`
  with the remaining width (`availW - cursorX`); its own text wraps
  internally against that remaining width, but wrapped continuation lines
  start at `cursorX`, and the next sibling starts after the child's widest
  line. **Text never flows from one inline element into the next across a
  wrap.**
- Inline items are top-aligned on the line (`lineHeight = max(heights)`);
  no baseline/vertical-align outside table cells.
- Text-level styling (`strong` bold, span colour, underline) is not a
  layout concern at all — svelterm re-derives it at paint time by folding
  visual attributes down the ancestor chain to each text node. Sumi
  already achieves the same via the ResolveStyles cascade stamping each
  node, so **no per-fragment styled-segment machinery is required** if we
  match svelterm.

So "parity" is much cheaper than the plan assumed. The real question is
whether we want parity or better-than-svelterm (browser-like) inline flow.

## Decision 1 (RESOLVED with Tom, 2026-07-05): real inline runs

Tom chose **Option B — a real inline formatting context** (text runs +
line boxes), exceeding svelterm's simplified cursor model. Browser-correct
wrapping: runs gathered across sibling/nested inline elements, lines
broken across element boundaries, per-line fragments emitted.

(Rejected alternative, for the record: svelterm-parity cursor model —
per-text-node wrapping, no fragments; cheaper but continuation lines
indent at cursorX and siblings sit after the widest wrapped line.)

### Inline formatting context (IFC) design

When a block container's in-flow children include inline-level content,
consecutive inline-level children form an IFC laid out as a unit:

1. **Run gathering.** Walk the inline subtree depth-first (text nodes,
   `display:inline` elements — recursing, `display:contents` — flattened,
   `inline-block` — atomic item). Produce a flat sequence of items:
   - `textRun{input *Input, text string}` — one per text node, style
     already stamped on the Input by the cascade;
   - `atom{input *Input}` — inline-block; measured via layoutNode
     (shrink-to-fit), placed as one unbreakable unit.
2. **Line breaking.** Break the concatenated run sequence against the
   available width: soft-wrap at spaces (breaking space consumed),
   hard-break overlong words, `white-space: nowrap|pre` respected,
   run boundaries are NOT break opportunities by themselves (browser
   behaviour: `a<strong>b</strong>` is one word). Rune-based widths for
   now (matches the rest of the engine; cell-width/grapheme awareness is
   a separate pre-existing gap).
3. **Fragment emission.** Each text run yields ≥1 fragments, each a
   rectangle on one line: `Fragment{X, Y int; Text string}` (parent-
   relative, like all pre-absolutePositions coords). Fragments are stored
   on the **text node's own Box** (`Box.Fragments`); the Box's X/Y/W/H
   become the bounding rect. An inline element's Box gets the union rect
   of its descendants' fragments. Atoms keep their ordinary Box.
4. **Line boxes.** Line height = max item height on the line (top-aligned,
   no baseline — svelterm also has none; terminal cells make baseline
   moot). `text-align` shifts whole lines within the container width —
   applied per line box in the IFC (subsumes today's per-text-node align
   when inside an IFC).
5. **Painting.** `renderContent` paints `Box.Fragments` when present
   (each fragment with the box's style), else falls back to
   Lines/Content as today. Single-text-node paragraphs can keep the
   existing Lines path (IFC with one run degenerates to it) — snapshot
   churn limited to genuinely mixed content.
6. **Hit-testing / cursor / selection.** Hit-test checks fragments when
   present (point-in-any-fragment instead of point-in-rect). D5 selection
   later maps naturally onto fragments.

Tree-shape constraint honoured: every Input still yields exactly one Box
(`input.Children[i] ↔ box.Children[i]` mapping intact for stampSizes /
self-pointers / diff); fragments are data on the Box, not extra boxes.

Inline element boxes (v1 limits): style only — no border, padding, or
horizontal margin on `display:inline` elements (defer; svelterm doesn't
support them either). `inline-block` supports the full box model.

## Scope common to either option

### 1. Parser: mixed content (prerequisite)

Today `parseNextChild` errors on loose text inside a container.
`<p>hello <strong>bold</strong> tail</p>` cannot be written.

- Loose text runs (with `{expr}` parts) inside a container body become
  tagless `TextElement`s interleaved with element children.
- `htmlBodyIsText` fast path stays for all-text bodies.
- Whitespace-only gaps between sibling elements (RESOLVED with Tom,
  2026-07-05 — JSX newline rule): a gap containing a newline is dropped
  at parse time (source formatting); a single-line whitespace gap
  becomes one space text node; runs with any non-whitespace are always
  kept verbatim (collapsed later at layout per Decision 3). Existing
  multi-line .sumi files parse identically — zero churn.
  (Rejected: always-emit + layout-time drop — literal CSS pipeline but
  adds whitespace children to every existing container.)
- Codegen: nothing new — loose text nodes are ordinary KindText inputs
  (extraction/sync already handles expression text nodes).

### 2. Display values

`Input.Display` gains `block | inline | inline-block | contents`
(today: "", none, grid, table + direction row/column for flex).

- UA stylesheet stamps defaults: `div,p,h1..h6,ul,ol,li,blockquote,pre,hr`
  → block; `span,strong,b,em,i,u,s,del,mark,kbd,abbr,samp,var,a,code`
  → inline. Text nodes are inline by nature.
- `display: flex` becomes an explicit value (current direction-based flex
  paths keep working; `flex` + `flex-direction` map onto them).
- Outer sizing: block fills available width (minus margins);
  inline/inline-block shrink-to-fit; height stays content-based.
- Open sub-question (Decision 2 below): what is the default display for
  a plain `div`/container — today's flex-column behaviour or block flow?

### 3. Block flow layout (`layoutBlockFlow`)

Port of svelterm's single-pass cursor algorithm into
`runtime/layout` (new file `flow.go`):

```
cursorX, cursorY, lineHeight, prevBlockMarginBottom
for each in-flow child (contents flattened):
  inline child  → layoutNode at (cursorX,cursorY) with remaining width;
                  cursorX += w; lineHeight = max; reset prevBlockMarginBottom
  block child   → close open line; collapse margins (§4);
                  layoutNode at full width; cursorY += h; record marginBottom
close trailing line
```

Positions parent-relative as elsewhere; absolute/fixed children keep the
existing partitionPositioned path.

### 4. Vertical margin collapse

Svelterm implements adjacent-block-sibling collapse only, positive
margins only: gap = `max(prevBottom, nextTop)`. Not implemented there:
parent/first-last-child collapse-through, empty-block collapse, negative
margins. Proposal: match exactly, in block flow only (flex containers do
not collapse margins, per CSS). Implemented in `layoutBlockFlow`, not in
`layoutColumn` (flex).

### 5. display: contents

Flatten recursively into the parent's flow list before the loop.
Tree-shape constraint: contents node still yields a Box placeholder
(zero-size at parent origin) with its children's boxes under it carrying
their flow positions — keeps pairwise Input↔Box mapping intact (same
trick as display:none nil placeholders, but children remain live).

### 6. text-align / baseline

Keep sumi's existing approach (resolver-inherited text-align applied at
layout). No baseline/vertical-align in flow (svelterm parity); table
cells already have their own handling.

## Decision 2 (RESOLVED with Tom, 2026-07-05): UA block default

HTML container tags get `display: block` from the UA stylesheet; block
flow is their normal path; flex requires explicit `display: flex`.
Existing examples/components/tests need a migration sweep — its own
slice (B4-migrate below).

Open rider: whether to ship a compat shim (presence of
direction/gap/justify attrs implies flex) or migrate all call sites to
explicit `display: flex` — see Decision 4.

(Rejected: opt-in block flow keyed on inline-level children — zero
migration but diverges from the browser mental model.)

## Decision 3 (RESOLVED with Tom, 2026-07-05): CSS-style collapse

Full `white-space: normal` semantics inside an IFC: whitespace runs
collapse to one space, leading/trailing whitespace per line stripped,
breaking space consumed at wrap, whitespace-only text nodes adjacent to
blocks (or in flex/grid/table parents) dropped. `white-space: pre`
preserves exactly; `nowrap` collapses but never wraps. Snapshot churn
from disappearing trailing spaces is accepted. Collapsing happens at
layout time (display is only known after the runtime cascade).

(Rejected: svelterm's subset — drop whitespace-only nodes only, keep
preserve-exact wrapping.)

## Decision 4 (RESOLVED with Tom, 2026-07-05): clean migration, no shim

No flex compat shim. direction/gap/justify attributes on a block
container are ignored (like a browser). The migration slice sweeps all
in-repo .sumi files and tests to explicit `display: flex` where flex
behaviour is relied on. Pre-1.0, no external users.

## Proposed slices (elephant carpaccio, updated for Decision 1 = IFC)

1. **B4a parser mixed content** — loose text in container bodies parses;
   codegen emits interleaved KindText children. RED: parser tests for
   `<p>a <strong>b</strong> c</p>`, expr parts, whitespace nodes.
   Value: mixed markup stops being a parse error (renders stacked until
   B4c, but writable).
2. **B4b line breaker + fragments (single block)** — run gathering over
   text runs only (no atoms yet), cross-run line breaking with CSS
   whitespace collapse (Decision 3), Fragment emission, fragment
   painting. RED: `<p>a <strong>b</strong> c</p>` wraps across the
   strong boundary browser-style; styles per fragment; whitespace
   collapses.
3. **B4-migrate** — sweep in-repo .sumi files/tests to explicit
   `display: flex` wherever flex behaviour is relied on (Decision 4).
   Must land immediately before B4c (same session) so the default flip
   doesn't break the tree.
4. **B4c block flow container** — UA display defaults (Decision 2),
   `layoutBlockFlow`: consecutive inline children form IFCs, block
   children break lines; block fills width, inline shrink-wraps.
5. **B4d inline-block atoms** — shrink-to-fit measurement, atomic
   placement on lines, top alignment.
6. **B4e margin collapse** — adjacent-block-sibling max() in block flow.
7. **B4f display:contents** — flatten into flow + placeholder Box.
8. **B4g fragment hit-testing** + C2 fidelity integration tests
   (`<p>` with strong/em/mark wrapping, click on a wrapped span).

Each slice: red → green → refactor → commit.
