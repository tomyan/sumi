# Elements

Sumi templates use HTML element names. Each element has a default style from the
user-agent stylesheet and, for controls, built-in keyboard and mouse behaviour.
This chapter lists the vocabulary, what each element renders on the cell grid, and
the events and state pseudo-classes it participates in.

The recognised tags are fixed — anything outside this list is treated as a
component reference, not an element. The legacy primitives `<box>` and `<text>`
were removed; use `<div>` and `<span>`.

## Structural and block elements

`div`, `p`, `h1`–`h6`, `ul`, `ol`, `li`, `blockquote`, `pre`, and `hr` are
block-level (they carry `display: block` from the user-agent stylesheet). The
generic containers `section`, `main`, `header`, `footer`, `nav`, `article`, and
`aside` are also recognised, but have no UA styling — they lay out as default
containers (a vertical column), which you can override with `display`.

- `h1`–`h6` — bold, with `margin: 1 0` (one blank row above and below). All six
  levels render identically apart from author styling.
- `p` — a paragraph with `margin: 1 0`.
- `ul`, `ol` — `margin: 1 0`, `padding: 0 0 0 2`. Each `li` is prefixed with a
  `• ` bullet via `li::before`. Deviation: `<ol>` is not numbered — ordered lists
  render the same bullet as `<ul>`.
- `blockquote` — indented (`margin: 1 2`, left padding) and dimmed.
- `pre` — `white-space: pre`, so spaces and newlines are preserved verbatim.
- `code` — inline; no special styling by default (pair with a class for colour).
- `hr` — a one-row horizontal rule drawn with a top border across the full width.

```sumi
<div>
	<h1>Report</h1>
	<p>Totals for the current period.</p>
	<ul>
		<li>First item</li>
		<li>Second item</li>
	</ul>
	<hr />
	<pre>  indented   spacing   kept</pre>
</div>
```

## Text-level elements

These are inline and can be mixed inside a line of text; runs wrap across them.

| Element | Rendering |
| --- | --- |
| `span` | no default style |
| `strong`, `b` | bold |
| `em`, `i`, `var` | italic |
| `u`, `a`, `abbr` | underline |
| `s`, `del` | strike-through |
| `mark` | black text on a yellow background |
| `kbd` | inverse video |
| `samp` | cyan text |

```sumi
<p>Press <kbd>Ctrl</kbd>+<kbd>C</kbd> to <strong>quit</strong>.</p>
```

## Links

`<a href="…">` is focusable when it has an `href`. Enter (while focused) or a
mouse click dispatches a `click` event and then opens the URL through the
platform opener. Underlined by default.

```sumi
<a href="https://example.com/docs" onclick={docsClicked}>Documentation</a>
```

## Form controls

### button

Focusable and centred (`text-align: center`). Enter or a mouse click dispatches a
`click` event. Deviation: Space does not activate a button (it activates only
checkboxes and radios). Add `disabled` to remove it from focus and match
`:disabled`.

```sumi
<button onclick={increment}>Press me</button>
```

### input (text and password)

Default width 20 cells. Editable with readline-style keys: arrows, Home/End,
Backspace, Delete, and Ctrl+A/E/B/F/H/D/K/U/W/Y/T. An `input` event fires with
`{value, cursor}` whenever the value changes. The initial value comes from the
`value` attribute (literal strings only). `type="password"` masks the value with
`•`. `maxlength` and `readonly` constrain editing. The value scrolls horizontally
to keep the cursor visible when it exceeds the width.

```sumi
<input type="text" value="hello" maxlength="40" oninput={nameChanged} />
```

### input (checkbox and radio)

Rendered as glyphs, 3 cells wide: checkbox `[ ]` / `[x]`, radio `( )` / `(•)`.
Space, Enter, a click, or a click on an associated `<label>` toggles the control
and dispatches `change` and `input` events carrying `{checked, value}`. A radio
checks itself, unchecks every other radio sharing its `name`, and never toggles
itself off. The `checked` attribute drives the `:checked` pseudo-class.

```sumi
<input type="checkbox" onchange={notifyChanged} />
<input type="radio" name="size" value="small" checked="true" onchange={sizeChanged} />
```

### select, option, optgroup

`<select>` is focusable and renders the selected option's label followed by ` ▾`,
sized to the longest option. Up/Down arrows move the selection with wraparound;
Space, Enter, or a click advance to the next option. Each change dispatches a
`change` event with `{value}` (the option's `value` attribute, falling back to
its text). `<option selected="true">` sets the initial choice; `<optgroup>`
groups options (its own label is not rendered).

```sumi
<select onchange={themeChanged}>
	<option value="dark" selected="true">Dark</option>
	<option value="light">Light</option>
</select>
```

### textarea

A multi-line text input. Content keeps its line structure (`white-space: pre`)
and the cursor is a (row, column). Editing keys match `input`, plus Enter inserts
a newline.

### progress, meter

Not focusable; 20 cells wide by default. Renders a bar of full blocks (`█`) with
an eighth-block partial, then track (`░`), from `value` against `min` (default 0)
and `max` (default 1). With no numeric `value`, the bar renders as an
indeterminate all-track strip.

```sumi
<progress value="0.4" />
<meter value="72" min="0" max="100" />
```

### details, summary

`<details>` shows its `<summary>` with a disclosure marker (`▶ ` closed, `▼ `
open) and hides its other children while closed. The summary is focusable; Enter
or a click toggles the open state and dispatches a `toggle` event with `{open}`.

```sumi
<details>
	<summary>Advanced options</summary>
	<p>Hidden until expanded.</p>
</details>
```

### dialog

Hidden unless it has the `open` attribute. An open dialog is modal: it traps
focus inside its subtree and captures mouse clicks outside it. Escape closes the
dialog and dispatches a `close` event, returning focus to the page.

```sumi
<dialog open="true" onclose={dialogClosed}>
	<p>Delete this item?</p>
	<button onclick={confirmYes}>Yes</button>
	<button onclick={confirmNo}>No</button>
</dialog>
```

### label

Associates a caption with a control, either through `for="id"` or by wrapping the
control. Clicking or activating the label focuses its control and synthesises a
click on it, so a label toggles the checkbox or radio it names.

```sumi
<label for="notify">Notifications</label>
<input id="notify" type="checkbox" />
```

## Tables

`<table>` with `<tr>` rows and `<td>`/`<th>` cells. `<thead>`, `<tbody>`, and
`<tfoot>` group rows; `<caption>` renders centred above the table; `<th>` is bold.
`<colgroup>`/`<col>` supply per-column width hints. Cells accept `colspan` and
`rowspan`. See [Layout](layout.md#tables) for sizing, `border-spacing`, and
`border-collapse`.

## Embeds

### region

A container that hands its content area to your code as a raw cell grid. When its
laid-out content size changes, it dispatches a `resize` event carrying
`{width, height}`; the handler fills `evt.Target.Cells` with a buffer of that
size, which is painted in place. Used to embed arbitrary rendered output (a
terminal mirror, a canvas) inside a sumi layout.

### ansi

Renders its text body — which may contain SGR escape sequences — as per-cell
styled content, sizing itself to the widest line and the line count. Useful for
dropping pre-coloured terminal output into the tree. Content is re-parsed each
render, so signal-driven bodies stay live.

### img

Renders the image at `src` (PNG, JPEG, or GIF) as half-block glyphs (`▀`), packing
two vertical pixels per cell. The decode is cached against `src`.

```sumi
<img src="logo.png" width="20" height="10" />
```

## State pseudo-classes

Controls expose state to CSS: `:focus` (the focused control), `:hover` (under the
mouse), `:checked` (a checked checkbox or radio), and `:disabled`/`:enabled`
(form controls with or without the `disabled` attribute). See
[Selectors](selectors.md) for how these resolve.
