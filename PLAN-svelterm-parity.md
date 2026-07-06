# Plan: Sumi parity with Svelte + svelterm

Companion to `COMPARISON-svelterm.md` (read it first — it is the gap analysis this
plan closes). Written 2026-07-05 to be executed across fresh sessions.

## Goal

A `.sumi` author gets what a svelterm author gets: standard HTML elements with UA
semantics, standard CSS (the cell-grid-sensible subset defined by svelterm's
`~/projects/svelterm/docs/reference.md`), W3C-style events, form controls, focus,
animation, and terminal capability handling — while keeping sumi's Go, compile-time,
single-static-binary model. Plus: a comprehensive documentation site and a published
announcement post on tomyandell.dev.

Parity target = svelterm's reference matrix, with `elements.md` as newer truth where
they disagree (images ARE supported there). Sumi keeps its extras (border-title,
box border-collapse, sticky, textedit depth).

## Method

Per-slice TDD (red → green → refactor → commit, no push). Each slice is a thin,
demonstrable unit of author-visible value with its own tests. Validate after each
slice that the next slice is still the most valuable. Keep files ~200 lines; split
test files to mirror source splits. Use scenario tests (`runtime/sumitest`) +
vt100 assertions for end-to-end slices; unit tests for parser/css/layout.

## Decisions (resolved with Tom, 2026-07-05)

1. **Element vocabulary: HTML tags only — `box`/`text` are DROPPED, not aliased.**
   The standard HTML tag set (`div`, `span`, `h1`–`h6`, `p`, `ul/ol/li`, `button`,
   `input`, …) becomes the sole vocabulary. Every existing `.sumi` file, component,
   example, and test migrates off `box`/`text` (see C1).
2. **CSS rename: clean break in one slice (A1).** No aliases, no deprecation
   period. Behavioural attrs (`focusable`, `onkey`, `onclick`, `contenteditable`,
   `cursor-*`, `scroll`) stay template attributes per the CSS purity rule (though
   C3/D1 supersede `onkey`/`focusable` for standard controls).
3. **Site: new `sumi-site` repo; launch with precompiled Go→WASM demos** in
   xterm.js (pure static hosting). Live-edit playground is a **fast-follow after
   launch** using **fully client-side compilation** (Go toolchain in wasm,
   lazy-downloaded on first edit with progress bar + cancel — see I6). No server
   compile service. Domain: **gosumi.dev**, REGISTERED via Route53
   2026-07-06 (sumi.dev squatted; sumi-rs.dev free for a future Rust
   port). Untrusted playground origin (I6): a SUBDOMAIN of
   svelterm-untrusted.net (decided with Tom) — no new registration.
   Deploy runbook: ~/projects/sumi-site/PLAN-deploy.md — read it first
   when resuming site/deploy work.
4. **Release gate for the blog post: full parity complete** — J3 publishes only
   after phases A–H are done. At publish time the repo must also be public with a
   version tag and the site live (site convention: never link unreleased projects).
5. **Scope: grid, tables, and images all stay in.** No trims — and since the
   announcement now gates on full parity, these sit on the critical path.

---

## Progress

- 2026-07-05: A1+A2 (692dbde), A3 (4091197), A4 (cabdff1), A5a/b/c
  (ccf436a, 9fdcc2e, 2b680af), A6 (5237859) complete. :hover was already
  fully wired; :focus now wired (FocusStyle/Focused + sync patch);
  :checked/:disabled/:enabled wait for C-phase controls. A7 done
  (amended commit: ParseColorValue in runtime/css/color.go + color_lab.go,
  named table generated from svelterm-ui data, alpha parsed-but-dropped
  until F2). A8 done (95929e0: render.ColorDepth + quantize at SGR
  emission, term.DetectColorDepth, RunOptions.ColorDepth). A9 done
  (83809a0 light-dark()→render.ColorPair resolved at emission;
  f526ef7 OSC 11 → EventScheme consumed by App loop). A10a done
  (b9def7d: @media parse + compile-time display-mode evaluation; rules
  carry Media query in the flat list preserving cascade order).
  RUNTIME STYLES MIGRATION DONE (Tom decided runtime over compile-time;
  RS1 b98ad91+4e17f0a, RS2 9c5eefe, RS3 2614a70, RS4 d5ac957). CSS is
  now fully runtime: Input carries Tag/ID/Classes/Attrs identity,
  components embed their stylesheet (style.Serialize →
  MustParseStylesheet), layout.ResolveStyles(root, ss, w, h) runs each
  converge with true path/sibling context (structural pseudos work in
  runtime-built {for} lists), inline attrs beat CSS, @media evaluates
  live viewport + scheme (A10 COMPLETE), transitions/animations resolve
  at runtime, annotation pre-pass deleted. Resolution is currently
  uncached (every converge) — cache/invalidation is a later perf slice
  (the "resolved ahead where possible" story). PHASE A COMPLETE:
  A11 var() w/ inheritance+fallbacks (622e770), A12 calc/min/max/clamp
  incl. %-deferred-to-layout via WidthCalc/HeightCalc (2119d40), A13
  @container (Layout stamps LastW/LastH; second resolve+layout pass in
  component render paths) + @supports name-checks (deef9c5), A14
  ::before/::after with content strings/attr()/concat synthesized
  idempotently by the resolver, invisible to sibling matching (4795ece).
  PHASE B PROGRESS: B1 margins+auto centring (aeb1116), B2 min/max/
  box-sizing w/ scroll-container MinWidth exception (b304f30), B3a
  space-around/evenly+align-self+order+reverse (ba7bb9e), B3b shrink/
  basis/flex shorthand (68b7d3d), B3c row flex-wrap (2c16ebe), B5 text
  props (7d8ca14), B6 GRID: tracks fr/%/repeat/minmax, areas, span,
  auto-flow rows, implicit auto rows (655e80f), B7a core tables:
  display table/table-row/table-cell, auto columns, proportional
  shrink, row-stretched cells (80338f6). REMAINING IN B: B7b tables
  (colspan/rowspan/caption/colgroup/collapse/table-layout:fixed), B4
  block/inline flow + margin collapse (flex column already approximates
  block; inline runs are the hard part). PHASE C STARTED: C1 COMPLETE —
  C1a HTML tags parse additively, text-vs-container by body content,
  whitelist so legacy lowercase components (<textedit>) survive
  (43dfa89); C1b UA stylesheet layered under author rules: bold
  headings, margins, ul/ol bullets via li::before (li always container,
  text bodies wrap in implicit "text"-tagged child), blockquote, hr
  rule, text-level defaults; nil-stylesheet components get UA styling
  (2f0e493); C1c box/text REJECTED with pointer errors, all .sumi +
  test fixtures migrated to div/span, dead parse paths removed
  (60da012). B7b done: colspan/rowspan via occupancy,
  caption above table, thead/tbody/tfoot flattening (5ab8a9e); UA
  table/tr display defaults make real <table><tr><td> markup work with
  th bold. Also landed: D2 F1-F12 keys
  (9477224), E1 steps() easing (f433369), D6 DEC 2026 synchronized
  output on screen frames (0ac7c42).

  C3-pre DONE: runtime-owned focus. tui.Component.FocusIndex +
  layout.CollectFocusables(tree); Tab/Shift-Tab cycling, Focused
  stamping (syncFocus before every render — survives dynamic
  rebuilds), and EventFocus/EventBlur dispatch all live in
  runtime/tui/focus.go; initial focus goes to the first focusable
  with an EventFocus. Deviations from the sketch (implementation
  choices, flag if wrong): (1) generated sync does NOT read
  comp.FocusIndex — Effects only re-run on signal writes, so the
  runtime stamps Input.Focused directly and codegen's focusIndex
  sync/conditional-cursor emission was deleted; (2) Input gained
  OnKey func(input.Event) (layout imports input, no cycle) emitted
  for focusable+onkey boxes — the blur/focus dispatch target; C3a
  merges it into the On map; (3) cursor-hide-when-unfocused moved
  into textedit's cursorX() (guards on its focused signal);
  (4) unused Input.FocusIndex field removed; triplicated OnEvent
  closures in component.go deduped into componentEventHandler.
  Tab is consumed only when the tree has focusables. New
  examples/focus scenario (snapshot shows :focus border + signal
  text flipping on Tab/Shift-Tab). Typing still does NOT reach
  focused elements — that is C3b by design.

  C3a DONE: layout.DOMEvent {Type, Key, Data, Target, StopPropagation(),
  PreventDefault(), DefaultPrevented()} in runtime/layout/domevent.go;
  Input.On map[string]func(*DOMEvent); HitTestPath returns root→deepest
  Input ancestry (ancestors stay on path for fixed-position children);
  DispatchDOM bubbles deepest→root honouring stopPropagation. tui click
  dispatch = HitTestPath + DispatchDOM{Type:"click"}. OnClick field +
  FindClickHandler deleted; HasClickHandlers checks On["click"].
  Codegen: any on<type>={expr} attr → On map entry; declared funcs WITH
  params emit direct refs (signature convention func(evt *sumi.DOMEvent));
  zero-arg exprs get nil-checked wrappers. Legacy onkey="name" string
  attrs stay on the OnKey path until C3b. prelude exports DOMEvent.
  New examples/clicker scenario (click inside button increments, miss
  does not). Button component (components/sumi/button.sumi) NOT wired —
  stdlib component consumption from app .sumi files is currently absent
  from the generate CLI (no cross-dir component registry); that revives
  with C4/H-phase work.

  C3b DONE: key (EventKey/EventSpecial) and paste events dispatch as
  "keydown"/"paste" DOMEvents bubbling along layout.FocusablePath(root,
  FocusIndex); a stopped event never reaches the root component handler
  (Component.OnEvent), unconsumed events still do. Focus/blur are now
  DOM events ("focus"/"blur") dispatched to the target only (no
  bubble, like the DOM). Input.OnKey deleted (field + codegen
  emission); DOMEvent gained Stopped(). examples/focus upgraded:
  onfocus/onblur/onkeydown={fn} with func(evt *sumi.DOMEvent), typing
  lands in the focused field, StopPropagation keeps typed keys from
  the root counter. NOT done in C3b (deliberate): textedit.sumi
  migration to DOM handlers — textedit is embedded source that nothing
  currently compiles (no cross-dir component consumption in the
  generate CLI), so it migrates when C6 wraps it and makes it testable.
  Mouse motion/release DOM types also deferred until a consumer exists
  (textedit selection at C6).

  C3c DONE: Tab/Shift-Tab dispatch as keydown to the focused path
  FIRST; cycling runs as the default action unless preventDefault
  (focus trap works). Enter on a focused element that has On["click"]
  synthesizes a bubbling click and is consumed; Enter on a plain
  focusable passes through to the root handler.

  C4 DONE (button element): parser makes button a containerTag
  (implicit untagged label child → borders/padding work); UA sheet
  gains button{text-align:center}; TEXT-ALIGN NOW INHERITS in the
  resolver (child stamped from parent before own declarations — own
  rules override; note resolver props are set-when-declared/sticky
  across passes, pre-existing behavior). layout.IsFocusable: control
  tags (button, more later) focus without focusable attr; disabled
  attr (value != "false") skips traversal — dynamic disabled={expr}
  needs C13. Click-to-focus: clicking a path containing a focusable
  focuses the deepest one (blur/focus dispatched). Codegen: On
  handlers now also emitted for TEXT-form elements (writeTextInput/
  writeExtractedTextNode). examples/buttons scenario: Enter presses
  Save, Tab+Enter presses Cancel, click presses+refocuses Save;
  snapshot shows bordered centred labels + cyan :focus ring.
  DEVIATION from svelterm noted: svelterm starts with NOTHING focused
  (focusIndex -1) until Tab/click; sumi focuses the first focusable
  at startup (kept from old sumi semantics — revisit with Tom).

  C5 DONE: a[href] focusable (no href → not); click/Enter activation
  runs the open default (nearest ancestor a[href] → tui.OpenURL, a
  replaceable hook defaulting to open/xdg-open); preventDefault
  suppresses. examples/links scenario.

  C6a DONE (input element, v1). ARCHITECTURE DECISION (flag to Tom):
  <input> is a UA-IMPLEMENTED ELEMENT like svelterm, NOT a wrapper
  over the textedit.sumi component — the orphaned runtime/edit
  package (full readline engine: kill ring, undo, history — built in
  the contenteditable phase, zero consumers) got a HandleKey keymap
  and now powers it. textedit.sumi stays as the H-phase component-
  library artifact. Mechanics: Input gains Edit *edit.State;
  controlTags += input; UA input{width:20}; runtime lazily attaches
  an implicit value child + edit state (value attr literal = initial
  value), syncs value/cursor per render (cursor only while focused);
  keydown default action edits + dispatches "input" DOMEvent
  Data{value,cursor}; preventDefault blocks editing. Readline map:
  arrows/home/end/backspace/delete, Ctrl+A/E/B/F/H/D/K/U/W/Y/T,
  paste. examples/textinput rewritten around the element
  (oninput={fn} → evt.Data["value"] → signal → greeting).
  NOT YET (C6b+): maxlength/readonly/password masking, view
  scrolling when value exceeds width, selection/mouse, word ops
  (alt arrows), bind:value sugar for elements, controlled
  value={expr} sync INTO the element.

  C6b DONE: edit.HandleKeyWith constraints (readonly = nav-only,
  consumed edits; maxlength caps typing + truncates paste);
  type=password bullet masking (real value in events); view
  windowing vs laid-out width (post-layout resyncInputElements +
  converge pass, since LastW is stamped by Layout); "input" events
  now fire only on value CHANGE (DOM semantics — cursor moves
  don't).

  C7a DONE (checkbox + radio): glyphs [x]/[ ] and (•)/( ) in the
  value child, UA width 3 via attr selectors, no caret; toggle is
  the CLICK DEFAULT ACTION (click / Enter / Space all synthesize or
  are clicks; Space only on checkables — it types into text
  inputs); radio checks self, unchecks same-name radios tree-wide,
  never self-untoggles; change+input DOMEvents carry
  {checked, value}. :checked/:disabled/:enabled resolve
  attribute-backed via css.ResolveWithStates — base+state rules
  cascade TOGETHER in spec+source order (not per-pseudo merges).
  examples/form scenario. KNOWN GAP: FocusStyle attributes other
  than colours (e.g. inverse) don't propagate from a box to its
  glyph child — pre-existing render inheritance behavior; use
  colour-based :focus styles on checkables for now.

  C7b DONE: label association (for=id or first wrapped control from
  input/button/textarea/select); synthesized click on the control
  runs its defaults with label-following off (no recursion);
  layout.PathTo added.
  C8 DONE: select projects "label ▾" sized to longest option
  (options Display:none; optgroup flattened); arrows wrap,
  Space/Enter/click advance; change {value}. examples/form has a
  theme select.
  C10 DONE: progress/meter eighth-block bars (█ + ▏▎▍▌▋▊▉ + ░),
  value/max/min, indeterminate track; UA width 20; UA projections
  unified into syncProjections walk (pre-layout) + resync
  post-layout (LastW-dependent).
  C11 DONE: details/summary — ▶/▼ marker (idempotent prepend),
  closed hides non-summary children; Enter/click on summary toggles
  + "toggle" {open}; summary in controlTags; display:none subtrees
  excluded from focus walks.
  C12 DONE: dialog — closed = display:none; open dialog is the FOCUS
  SCOPE (Component.lastFocusScope private field): Tab traps inside,
  focus pulled in on open (no blur/focus events on scope change —
  v1 gap), clicks outside captured (no dispatch), Escape removes
  open + "close" event, focus returns to page index 0 (DOM would
  restore previous — v1 gap). focusedPath crosses the scope for
  bubbling.

  C13 DONE: dynamic attr sync. Expression-valued STATE attrs (class,
  disabled, checked, open, selected, value, href, maxlength,
  readonly — allowlist dynamicSyncAttrs in codegen_layout.go) force
  extraction and emit sync-Effect patches: class → Classes +
  Attrs["class"] via sumi.SplitClasses; rest → Attrs[k] =
  sumi.AttrString(expr) (fmt.Sprint; bools become "true"/"false"
  which boolAttr/IsFocusable treat correctly). Attr exprs are RAW
  (no signal auto-unwrap — write .Get()). examples/dialog:
  open={confirming.Get()} modal, button opens, No/Escape close (close
  event resets signal). NOTE: value={expr} patches the attr but Edit
  state still initializes once — controlled-input semantics and
  bind:value sugar remain open.

  C9 DONE: textarea — container tag, focusable, edits through the
  same input default action with Constraints.Multiline (Enter
  inserts newline honoring readonly/maxlength; Up/Down move lines
  via edit.LineCol/CursorUp/CursorDown, column-preserving with
  clamp); projection keeps line structure via white-space:pre on
  the value child and a (row,col) cursor. v1 gaps: no horizontal
  windowing, no vertical scrolling, initial value via value attr
  only (body-as-initial-value TODO). ALSO: select/details/dialog
  added to parser containerTags (borders/padding on all controls).

  C2 DONE: UA gains abbr underline, samp cyan, kbd inverse
  (DEVIATION: svelterm borders kbd — sumi text nodes can't take
  borders), mark black-on-yellow (svelterm parity; dark-scheme
  variant TODO with UA light-dark pass), caption centred.

  C14+C15 DONE (Tom approved option (a)): Input/Box.Cells
  *render.Buffer — per-cell styled content blitted at the content
  origin (border+padding aware; Box now carries Padding), clipped
  to the content area (runtime/layout/rendertree.go renderCells).
  <ansi>: text body with raw SGR parsed via runtime/vt100 into
  Cells each projection pass (\n fed as \r\n), sized to visible
  content unless width/height attrs; source child hidden.
  <region>: consumer-fed — post-layout resync dispatches "resize"
  DOMEvent {width,height} (content area) when the laid-out size
  changes (tracked via internal Attrs["sumi:region-size"] marker);
  the handler feeds evt.Target.Cells. Preview's OnPostRender
  injection can migrate onto <region> later (G/preview cleanup).

  C16 DONE — PHASE C COMPLETE. img renders src (png/jpeg/gif via
  stdlib decode, file paths only) as half-blocks into Cells: '▀'
  fg=top/bg=bottom, '▄' when only the bottom half is opaque, blank
  when both transparent (alpha < 50%); nearest-neighbour sampling;
  intrinsic size = pxW × ceil(pxH/2) cells, width/height attrs
  scale; decode cached via Attrs["sumi:img-src"]. Deferred: data:
  URIs, async load/reflow, kitty graphics protocol.

  F1 DONE (line styles): DrawStyledBorder glyph families single/
  double/rounded/heavy/ascii, unknown → single fallback; collapse
  junction merging stays single-only (sumi extra). F1b remains:
  eighth/half/full-cell block edges + border-corner + per-side
  left/right toggles (svelterm border.ts BLOCK_EDGES is the
  reference; inner/outer corner semantics differ per style).
  D4a DONE: App.ExitOn / RunOptions.ExitOn quit chords ("ctrl+x",
  single char, special-key name), default ctrl+c; fires only when
  nothing consumed the event. D4b remains: Ctrl+Z suspend/restore
  (restore termios+altscreen+modes, SIGTSTP self, SIGCONT re-enter
  + repaint; needs a subprocess/PTY test — runtime/pty exists).
  E3a DONE: animation-play-state paused freezes elapsed (pause point
  recorded at the render that sees paused; resume shifts animStart).
  E4 DONE: prefers-reduced-motion media feature; RunOptions.
  ReducedMotion + SUMI_REDUCED_MOTION env; authors gate animation
  rules in @media. E3b remains: var()/light-dark() resolution at
  animation start — keyframe stops are baked to render.Style at
  CODEGEN (writeKeyframeRegistration), so vars need stops to carry
  raw props resolved at start; NOTE light-dark() may already work
  via the ColorPair emission mechanism — verify before building.

  E2 DONE: Engine.StepLength — width/height transitions step whole
  cells; state keyed per node (anim.LengthState on Input, no render
  IDs); mid-flight retargeting restarts from the displayed value;
  Run paths step between resolve and layout (TestApp deliberately
  doesn't animate — deterministic snapshots). CSS-driven lengths
  only (resolver restamps targets; inline width attrs don't).
  F2a DONE: Color.A (0=opaque); alpha survives parsing (#rrggbbaa,
  #rgba, slash + legacy 4-arg fn syntax, clamped [1,254]); Buffer
  cell writes composite BG over backdrop BG and FG over effective
  backdrop; non-RGB backdrops paint opaque; stored cells always
  opaque so diff/SGR/depth never see alpha.
  F2b DONE: numeric opacity <1 = alpha on the element's RGB colours;
  `dim` keyword and non-blendable colours keep the Dim attribute.
  F4a DONE: RunOptions.ColorScheme (forces + locks against OSC 11
  via App.SchemeLocked) and RunOptions.Mouse *bool override.
  B7c-1 DONE: border-spacing (h v; UA table default 2 0 = svelterm
  parity) threaded through column offsets/row widths/colspans;
  table-layout: fixed sizes from first row (explicit widths hold,
  remainder splits evenly).

  B7c COMPLETE: B7c-2 collapsed cell borders (BorderCollapse table →
  cells overlap by 1 via spacing -1 through the shared offset math;
  Collapsed edges marked → junctions ┬├┼ emerge from the existing
  merge machinery; rowspans shorten by overlapped lines). ALSO FIXED
  a latent B7a bug: cells were placed table-relative inside row
  boxes and double-shifted by the final absolutePositions pass
  (second-row cells at 2× Y; padded tables double-shifted columns) —
  cells are now row-relative. KEY INSIGHT for future layout work:
  Box children coordinates are PARENT-RELATIVE until the
  absolutePositions pass at the end of Layout. B7c-3: colgroup/col
  (+optgroup) parse; UA hides colgroup; col width hints override
  computed columns; empty-cells: hide clears borders on contentless
  cells. F1b: block-edge borders (eighth/half/full-cell; inner =
  blank corners, outer/full extend horizontal through corners;
  svelterm's quadrant half-cell corners + border-corner attr still
  open).

  NEXT (bigger slices, in rough value order — each wants a fresh
  session):
  - B4 block/inline flow + margin collapse (LAST remaining B item;
    gates C1/C2 full fidelity). DESIGNED 2026-07-05 — see
    design-b4-block-inline.md (decisions resolved with Tom: real
    inline-run IFC with Box.Fragments — note svelterm itself has NO
    text-run model, we exceed parity here; UA block default with clean
    migration to explicit display:flex, no compat shim; CSS whitespace
    collapse at layout time; parser whitespace gaps use the JSX newline
    rule; margin collapse = adjacent block siblings, positive-only,
    block flow only). Slices B4a..B4g in the design doc.
    COMPLETE 2026-07-05 (Phase B now fully COMPLETE): B4a
    mixed-content parser (6bcf73d), B4b IFC line breaker +
    Box.Fragments (9e02e01), B4c-1 block flow + fill width +
    flex-attr gating (d007b1a), B4c-2 nested inline elements +
    per-property Style.Inherit (ce0f724), B4c-3 UA display flip +
    Hidden runtime flag + display:flex migration + golden churn
    (ec3d8bd), B4e margin collapse (7b3781d), B4d inline-block atoms
    with per-line heights (71adbc4), B4f display:contents flatten +
    union placeholder (2ff23e1), B4g fragment-aware hit testing
    (edfb18e). Display semantics: "" = legacy flex-column (raw Input
    trees unaffected), "block" = block flow, "flex" = explicit flex;
    UA flip only affects tag-resolved trees.
    NOTE: projections must NEVER write CSS-owned Input fields (the
    cascade re-stamps them) — use runtime flags like Hidden instead.
    (text-align per-line inside an IFC shipped as a follow-up.)
  - D5 global selection + clipboard — DONE 2026-07-06 (eb9d0db).
    Screen-space SelectionController in runtime/tui/selection.go
    (svelterm model: cell coords, drag/double-word/triple-line, click
    clears); inverse-toggle overlay applied to frameBuf after tree
    render (Ch untouched → extraction exact); copy on left release via
    OSC 52 + pbcopy/wl-copy/xclip (App.Clipboard injectable). Presses
    on editable controls skip global selection (they own their drag).
    Mouse mode now DEFAULTS ON in Run/RunWithOptions (RunOptions.Mouse
    overrides). Gaps kept at svelterm parity (no Escape-clear, no
    Ctrl+C copy, no wide-glyph cells, no drag auto-scroll) —
    SelectionController.Clear() exists for a future Escape binding.
  - D4b Ctrl+Z suspend — DONE 2026-07-06 (e76c77e): default action
    unless preventDefault; enterTerminal/exitTerminal factored from
    Run; SIGTSTP + NeedsFullRedraw on resume; PTY subprocess test
    (SIGSTOP under env var — PTY session leader's orphaned pgrp
    discards default SIGTSTP; raw 0x7f wait check — Go darwin
    Stopped() misreads SIGSTOP stops as continued).
  - D3 kitty keyboard — DONE 2026-07-06 (2205c74): flag-1
    disambiguate-only per svelterm (optimistic push \x1b[>1u after
    altscreen in enterTerminal, pop in exitTerminal — suspend/panic
    covered by the shared lifecycle); CSI-u decode with shared
    modifier mask; PUA functional map; Shift+Tab convention kept.
    Also DONE: F4b io injection + onLog + Run/RunWithOptions dedupe
    (bc34d60).
  - D7 terminal matrix CI — DONE 2026-07-06 (ff1a56a): ANSI
    round-trip tests (RunWithOptions with injected streams → replay
    through runtime/vt100, assert content/styles/diffs; caught a
    vt100 '<'-marker gap on day one); docs/terminals.md support
    matrix; GH Actions ubuntu+macos build/vet/test/gofmt;
    runtime/pty split darwin/linux (TIOCSPTLCK+TIOCGPTN).
    PHASE D COMPLETE. Also: keyframe registry duplication removed
    (7294e37).
  - E3b keyframe var()/light-dark() — DONE 2026-07-06 (94792d5):
    resolver stamps AnimationSpec.Stops per node (var() from node
    scope, ColorPair kept); engine prefers spec stops over registry;
    LerpColor collapses pairs to the active scheme before lerping
    (fixes transitions too); IFC boxes propagate anim fields.
    Simplification opportunity: codegen writeKeyframeRegistration +
    Component.Keyframes registry is now redundant (stylesheet already
    serializes @keyframes; resolver stamping always wins) — remove in
    a cleanup slice.
  - F3 inline mode + FrameLog — DONE 2026-07-06 (2ef01b5 InlineScreen
    driver, 211a1db RunOptions.Inline, 5d6abfe FrameLog, f7b6bf2 CPR
    origin + inline mouse). Design in design-f3-inline-mode.md
    (untracked). PHASES E AND F COMPLETE.
  - G tooling — DONE 2026-07-06 (PHASE G COMPLETE): G1 init scaffold
    w/ replace-to-local-checkout (cc15d0b), G2a rebuild pipeline +
    tree watcher (54990a6), G2b PTY-supervisor sumi dev with vt100
    mirror, keep-last-good error bar, hot swap (0a8b91b; design
    decision: PTY supervisor over bare exec loop, resolved with Tom;
    NOTE component OnEvent wiring is name-based on handleKey), G3
    inspect over SUMI_CONTROL_SOCKET w/ short hashed socket path —
    sun_path 104-byte cap (dae0235). Design: design-g-tooling.md
    (untracked). H PARKED per Tom 2026-07-06 (leave the UI package).
    I3 WASM SPIKE DONE 2026-07-06 (5fa8341): whole runtime + example
    apps build GOOS=js GOARCH=wasm (pty tagged unix, term resize js
    stub, suspend stopSelf seam); F4b RunOptions.In/Out is the
    xterm.js IO bridge; CI guards the wasm target. Live demos are GO
    for the site. I1+I2+I3 CORE DONE 2026-07-06: 11 docs chapters in
    docs/ (ea9e8be; reference.md = support matrix); sumi-site repo at
    ~/projects/sumi-site (SvelteKit + adapter-static + mdsvex/shiki,
    sync-docs from sumi/docs, sumi-e identity per Tom: washi/ink/
    vermillion, brushstroke, 墨 wordmark); live wasm demos via
    runtime/webterm + xterm.js — verified interactive in Chrome
    against the true static build (NOTE vite preview serves
    .svelte-kit/output, NOT build/ — test with a plain static
    server). I4 DEPLOYED 2026-07-06: https://gosumi.dev LIVE (zone
    adopted via import block, S3+CloudFront+ACM via OpenTofu, profile
    tyanroot; module fix vs svelterm's copy: s3 objects use
    source_hash not etag — multipart etag on >5MB wasm never equals
    filemd5, perpetual drift; static/404.html added, was missing).
    I5 launch checks PASS (wasm content-type + edge compression,
    live demo keypress-verified via CDP, docs deep-links + rewritten
    internal links, light/dark, mobile, 404). Remaining in I: more
    demos, docs review pass (docs pages ship no <title>), sitemap
    line in robots.txt. NOTE:
    J blog gate says phases A–H — needs revisiting with Tom since H
    is parked.
  - H component library (needs cross-dir component consumption in
    the generate CLI).
  - I docs/site, J blog (gated on A–H complete).
  (original C3a sketch follows for reference)
  - C3a-sketch: layout.DOMEvent {Type, Key input.Event, Data map, Target
    *Input, StopPropagation, PreventDefault} (layout may import input —
    no cycle). Input gains On map[string]func(*DOMEvent), emitted by
    codegen for on<type>={expr} attrs (onclick migrates INTO the map;
    legacy OnClick field retired same commit to avoid dual paths;
    FindClickHandler replaced by hit-test returning the Input PATH to
    the deepest node).
  - C3b: dispatch in tui: mouse events target the hit-tested path;
    key/paste events target the focused input's path; bubble
    target→root calling On[type]; capture phase via "capture" wrapper
    later if needed (svelterm has it; v1 bubble+stopPropagation).
  - C3c: default actions after dispatch unless defaultPrevented
    (Tab cycling, Enter activation) — foundation for controls C4+.
  Also open: B7c table borders (collapse/spacing/table-layout:fixed/
  colgroup), B4 block/inline + margin collapse, dynamic classes
  (class={expr} → extraction sync of Input.Classes), ol counters,
  B3 wrapped-row grow/align, kitty keyboard (D3), selection (D5),
  suspend/exitOn (D4). Notes: codegen resolves styles via annotateStyles pre-pass
  onto template nodes; sibling context is statically-known siblings only
  ({if}/{for} bodies are their own sibling scope); :not/:is args are
  compound-only; state pseudos are subject-only. Memory correction:
  component INLINING was removed — components are runtime NewFoo
  constructors (memory updated).

## Phase A — Standard CSS core (foundation; everything else builds on it)

- **A1. Standard property names.** Rename to `justify-content`, `align-items`,
  `font-weight` (≥700=bold), `font-style: italic`, `text-decoration: underline |
  line-through`. Keep `dim`/`inverse` as documented terminal extensions
  (svelterm uses `opacity`≈dim — adopt `opacity` too). Migrate all examples,
  components, tests. Value: sumi CSS is real CSS from here on.
- **A2. Graceful-drop policy.** Unknown properties and pixel-derived units parse and
  drop silently (svelterm's rule), never error. Tested for px/em/rem/font-size etc.
- **A3. Units.** `cell` unit + `ch` alias; unitless 0; ints without units remain
  valid (sumi extension). `%` for width/height/basis against containing block.
- **A4. Selector engine v1.** `#id`, `*`, selector lists, descendant and `>` child
  combinators; specificity + source-order cascade; inline attrs keep highest
  precedence. (Requires resolving styles per-node against an element tree — the
  structural prerequisite for everything below.)
- **A5. Selector engine v2.** `+`, `~`, attribute selectors (all 7 operators),
  structural pseudo-classes incl. full An+B `:nth-*`, `:not()`, `:is()`, `:where()`.
- **A6. State pseudo-classes wired.** `:focus`, `:hover` (finish DESIGN-hover.md),
  `:checked`, `:disabled`, `:enabled` re-resolve on runtime state change.
- **A7. Colour values.** `#rgb`/`#rrggbbaa`, `rgb()`/`hsl()`/`hwb()`/`lab()`/`lch()`/
  `oklab()`/`oklch()` (legacy + modern syntax), 148 named colours, `transparent`,
  `currentColor`. (Port logic from svelterm-ui's `src/color.ts` semantics; write in Go.)
- **A8. Colour depth detection + degradation.** Probe `COLORTERM`/`TERM`; degrade
  truecolor→256→16→mono; honour `NO_COLOR`; `colorDepth` override in RunOptions.
- **A9. Scheme detection.** OSC 11 query → dark/light; `light-dark()` function;
  `@media (prefers-color-scheme)` once A10 lands; `colorScheme` override.
- **A10. `@media`.** `display-mode: terminal|browser`, `prefers-color-scheme`,
  `min/max-width/height` in cells; re-evaluate on resize; nested rules in
  declarations.
- **A11. `var()` custom properties** with inheritance + fallbacks; `inherit`/
  `initial`/`unset` keywords.
- **A12. `calc()`, `min()`, `max()`, `clamp()`** over cells and `%`.
- **A13. `@container` (size queries) and `@supports`.**
- **A14. `::before`/`::after`** with `content:` strings, `attr()`, concatenation.

## Phase B — Layout parity

- **B1. Margin** (4-side, shorthand incl. fixed 3-value parsing bug, `auto`
  centring, vertical margin collapse).
- **B2. min/max-width/height** + percentage sizing + `box-sizing`.
- **B3. Flex completion:** `flex-shrink`, `flex-basis`, `flex` shorthand,
  `flex-wrap`, `row-reverse`/`column-reverse`, `align-self`, `order`,
  `space-around`/`space-evenly`.
- **B4. Block/inline flow.** `display: block | inline | inline-block | contents`;
  inline runs with text-level styling flowing through wrapped text.
- **B5. Text properties:** `text-align`, `text-overflow: ellipsis` (+
  `ellipsis-middle`), `white-space: normal|nowrap|pre`, `text-transform`,
  `word-break`, `visibility`.
- **B6. Grid:** `grid-template-columns/rows` (`cell`/`ch`/`%`/`fr`, `repeat()`,
  `minmax()`), `gap`, numeric/named placement + `span`, `grid-template-areas`.
  Row-based auto-flow only (match svelterm's deviation).
- **B7. Tables:** table display types, colspan/rowspan, header/footer groups,
  caption, colgroup width hints, `border-collapse`/`border-spacing`,
  `table-layout`, `empty-cells`. (Reuse sumi's junction-drawing.)

## Phase C — Elements and form controls

(Each control slice includes: rendering, focusability, keyboard/mouse defaults,
events, state pseudo-class, UA style, docs entry, example.)

- **C1. Element tree + UA stylesheet v1:** `div`, `span`, `h1`–`h6`, `p`, `ul/ol/li`
  (markers), `blockquote`, `pre`, `code`, `hr`. **`box`/`text` are removed** (per
  decision 1): parser rejects them with a helpful error, and all in-repo `.sumi`
  files (components/, examples/, cmd/sumi/preview/, testdata) migrate to
  `div`/`span` in the same slice.
- **C2. Text-level elements:** `strong/b`, `em/i`, `u`, `s/del`, `mark`, `kbd`,
  `abbr`, `samp`, `var` → attribute styling.
- **C3. Event model.** Capture/bubble dispatch on the element tree;
  `stopPropagation`/`preventDefault`; standard types `click`, `keydown`, `input`,
  `change`, `paste`, `toggle`; payload struct (Go's `event.Data`). Template syntax
  `onclick={handler}` etc. This replaces/deprecates bespoke `onkey` dispatch.
- **C4. `button`** (centred label, Enter/click → click event) — port Button component
  semantics into the element.
- **C5. `a`** (underlined, focusable, Enter/click opens href via `open`/`xdg-open`).
- **C6. `input` (text)** — wrap existing textedit engine; `value`, cursor,
  `input` events with `{value, cursor}`.
- **C7. `checkbox` + `radio`** (`[x]`/`(•)` glyphs, Space toggles, radio `name`
  groups, `:checked`, `change`/`input` events) + **`label`** association
  (wrapping and `for=`).
- **C8. `select`/`option`/`optgroup`** — popup-less cycling control per svelterm.
- **C9. `textarea`** — multi-line textedit.
- **C10. `progress` + `meter`** — block-glyph bars with eighth partials.
- **C11. `details`/`summary`** — ▶/▼ disclosure, `open`, `toggle` event.
- **C12. `dialog`** — modal, Tab focus trap, Escape closes, `close` event.
- **C13. `disabled` attribute** across controls; focus skips; `:disabled`.
- **C14. `<region>`** (svt-region equivalent) — consumer-fed cell area + `resize`
  event; wire to `runtime/vt100` for embedded terminals.
- **C15. `<ansi>`** (svt-ansi equivalent) — raw SGR passthrough, treated as `pre`.
- **C16. `img`** — half-block rendering (2 pixels/cell) from file paths; then kitty
  graphics protocol on capable terminals (Ghostty/kitty/WezTerm).

## Phase D — Input & terminal capability

- **D1. Focus polish:** focus follows control types automatically (no `focusable`
  attr needed on standard controls), click-to-focus, disabled skipped.
- **D2. F1–F12 + full modifier decoding.**
- **D3. Kitty keyboard protocol** (detect, enable, decode; legacy fallback).
- **D4. Ctrl+Z suspend/restore** (restore + repaint on fg) and `exitOn`
  configuration (default ctrl+c).
- **D5. Global text selection + clipboard:** drag/double(word)/triple(line)
  selection painted inverse; copy via OSC 52 + pbcopy/wl-copy/xclip fallback.
- **D6. DEC 2026 synchronized output** (detect + wrap frames).
- **D7. Terminal matrix CI:** ANSI round-trip tests against a terminal model;
  document a support matrix like svelterm's `terminals.md`.

## Phase E — Animation completion

- **E1.** `steps()`/`step-start`/`step-end` easing; per-svelterm discrete-at-midpoint
  semantics for non-interpolable properties.
- **E2.** Transition property lists beyond color/background (lengths step whole
  cells); `transition-delay`.
- **E3.** Keyframes: `var()`/`light-dark()` resolution at start; `animation-delay`;
  wire `play-state`.
- **E4.** `prefers-reduced-motion` media feature.

## Phase F — Runtime API & rendering parity

- **F1. Border styles rendered:** double, rounded, heavy, ascii, eighth-cell-inner/
  outer, half-cell-inner/outer, full-cell; `border-corner: h|v`; per-side
  left/right toggles (top/bottom exist).
- **F2. Alpha compositing** at paint time; numeric `opacity` as blend factor / dim.
- **F3. Inline mode** (render at shell cursor, no altscreen) + **FrameLog**
  equivalent (append/update/archive frames into scrollback).
- **F4. RunOptions parity:** `io` injection (custom reader/writer), `onLog` capture
  (stdlib `log`/fmt guidance), `mouse` toggle, `fullscreen` toggle, `colorDepth`,
  `colorScheme`, `exitOn`.

## Phase G — Tooling

- **G1. `sumi init`** — scaffold a runnable app.
- **G2. `sumi dev`** — watch `.sumi`, regenerate, rebuild, relaunch (simple loop
  first; the design-ui-support.md HMR/section-diffing comes later, out of parity
  scope).
- **G3. `sumi inspect`** — tree/computed-style/box dump over the existing sumitest
  control socket (svt-like).

## Phase H — Component library parity (`components/sumi/`)

Leverage new elements where possible; components remain for composition value.

- **H1. Dialog** (on `dialog` element) — title, width, onclose.
- **H2. List** — keyboard-navigable selection, wraparound, onselect, scroll.
- **H3. Tabs** — tab bar + active panel (add keyboard nav — improve on svelterm-ui).
- **H4. Toaster + `toast()`** — timed queue, info/success/error, top-right overlay.
- **H5. FuzzyPicker** — fuzzy matcher (Go port of subsequence scorer) + filtered list.
- **H6. Colour suite (optional, decision 5):** swatch/palette/slider/panel/picker;
  needs A7's colour engine.

## Phase I — Documentation + site

Docs authored in `sumi/docs/*.md` (mirroring svelterm: site imports them).

- **I1. Docs skeleton:** `getting-started`, `terminal-css`, `layout`, `selectors`,
  `elements`, `motion`, `reference` (the authoritative support matrix — start it in
  Phase A and update per slice, svelterm-style: MDN links, deviations stated
  plainly, three-bucket model).
- **I2. Site repo `sumi-site`:** SvelteKit + adapter-static + mdsvex docs pipeline +
  shiki dual-theme, cloned structurally from svelterm-site; S3+CloudFront via
  OpenTofu (`tyanroot` profile). Landing = hero + live demo + "Why" / "Status"
  prose. Distinct visual identity (do NOT reuse svelterm's terminal-hardware
  theme; sumi is ink-brush/Japanese-minimal territory — design at build time).
- **I3. Examples on site:** precompiled Go→WASM demos rendered via xterm.js
  (counter, todo, panels, flexbox dashboard, textinput, animation). Requires a
  WASM IO shim for `runtime/term` + `runtime/input` (no PTY in browser) — spike
  first; fall back to asciinema-style recorded casts if WASM is blocked.
- **I4. Docs chapters filled** as phases complete: `inline-mode`, `terminals`,
  `compatibility`, `distribution` (single-binary story is sumi's headline),
  `testing` (scenario/sumitest — sumi's unique chapter), `tooling`.
- **I5. Launch checks:** 404s, robots, RSS n/a, deep links, mobile.
- **I6. Live-edit playground (fast-follow, post-launch): fully client-side
  compilation.** No server, no infra — the whole pipeline runs in the browser:
  - sumi codegen compiled to `GOOS=js GOARCH=wasm`, run in a web worker
    (.sumi → Go, instant).
  - Go toolchain (`cmd/compile` + `cmd/link`) compiled to wasm, run in a worker
    against a virtual in-memory FS; hand-orchestrate compile → link with an
    importcfg (no exec in browsers). Prior art: ccbrown/wasm-go-playground,
    GopherJS playground.
  - Stdlib + sumi runtime precompiled to js/wasm `.a` archives at site-build
    time and shipped as fetchable assets — a structurally closed world (the
    only importable packages are the ones we ship).
  - Linked output runs in the untrusted-origin iframe with xterm.js, same as
    precompiled demos.
  - **UX (agreed):** demos are precompiled and instant; clicking "edit" starts
    the toolchain + archive download in parallel with a progress bar and a
    cancel button. Cache via service worker with immutable URLs so the
    download happens once per toolchain version. Mobile can stay demo-only.
  - **Spike first:** measure toolchain wasm + archive download size (expect
    tens of MB compressed), compile+link latency for playground-sized apps
    (expect seconds; `GOMAXPROCS=1` under wasm), and peak memory (expect a few
    hundred MB). Toolchain artifacts regenerate on every Go version bump —
    script it in the site build.
  - Rejected alternatives: server compile service (infra + abuse surface for
    little gain once this works); yaegi-in-wasm interpreter (generics coverage
    insufficient for `signal.New[T]`-style generated code).

## Phase J — Blog announcement (tomyandell-site)

- **J1. Rewrite `src/routes/blog/introducing-sumi/+page.svelte`** — the existing
  draft predates the signals migration (shows `$state` runes; wrong API). Rewrite
  against current reality. Follow `tomyandell-site/CLAUDE.md` voice guide strictly
  (British English, no em dashes, no LLM-isms, lead with the argument). Structure
  per `introducing-svelterm`: hook with runnable snippet → disclosure → Why → What
  it does → relationship to svelterm (they share a thesis; different trade-offs:
  Go/static binary/compile-time vs Svelte/Node) → How it was built → Status.
  ~1,000–1,450 words. Escape `{`/`}` as `{'{'}`. Update `date`, keep `draft: true`.
- **J2. Update the sumi project page** (`src/routes/projects/sumi/+page.svelte`)
  and the projects index card.
- **J3. Publish (only after phases A–H are complete, repo public + tagged, site
  live):** flip `draft: false`, consider a
  follow-up line/link from the svelterm post's "Go detour" section, `npm run build`,
  `tofu apply` in `tofu/`.

## Sequencing and dependencies

- A is strictly first (A1–A6 unlock everything; A4 element-tree resolution is the
  keystone). C3 (event model) gates C4–C13. B4 (inline flow) gates C1/C2 full
  fidelity but C can start with block-level elements. F1 (borders) is independent —
  good early win. H needs C3+C12. I1/reference.md grows continuously from Phase A.
  J1 can be drafted any time; J3 waits for the release gate.
- Suggested first five working sessions: A1+A2 → A3+A4 → A5+A6 → A7+A8 → F1.

## Verification

- Every slice: unit tests (parser/css/layout/render) + at least one scenario test
  through `runtime/sumitest` asserting final cells via vt100.
- Port svelterm's docs examples into sumi scenario tests as acceptance criteria
  where semantics should match.
- Keep `docs/reference.md` in lockstep with code (the svelterm repo's own
  reference-vs-elements drift is the cautionary tale).
