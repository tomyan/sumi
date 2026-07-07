# Bindings and form controls

Sumi renders the HTML form controls — `<input>`, `<textarea>`,
`<select>`, checkboxes and radios — on the cell grid, with a user-agent
stylesheet and the editing behaviour you'd expect. This chapter covers
how to read and write their values, and the constraints that apply.

## Two-way binding with `bind:`

The primary way to connect a native control to a signal is `bind:` —
`bind:value` on text inputs, textareas, and selects, and `bind:checked`
on checkboxes and radios. It is two-way: the control writes the signal as
the user edits, and an external write to the signal updates the rendered
control.

```sumi
<script>
name := sumi.New("")
</script>
<input type="text" bind:value={name} />
<span>Hello, {name}</span>
```

`bind:value` expects a `*sumi.Signal[string]`; `bind:checked` expects a
`*sumi.Signal[bool]`. A mismatch is a plain Go compile error.

### The event escape hatch

`bind:` is sugar over the control's own DOM events, and you can wire those
directly when you need more than mirroring a value — for example to
validate, transform, or react to the change without holding a signal:

```sumi
<script>
name := sumi.New("")

func nameInput(evt *sumi.DOMEvent) {
	name.Set(evt.Data["value"].(string))
}
</script>
<input type="text" oninput={nameInput} />
```

The control owns its editing state (text buffer, cursor, view offset) in
both cases; `bind:` just writes the mirroring handler and the display sync
for you. The sections below give the event name and `Data` keys each
control fires, which are what `bind:` is built on.

### Text input

`<input>` and `<textarea>` hold an edit buffer. Typing, Backspace/Delete,
Home/End, and the readline motions (Ctrl+A/E/B/F) move and edit it.
Each edit that changes the value fires an `input` event whose
`Data["value"]` is the new string and `Data["cursor"]` the caret index.

- **Initial value**: `bind:value={sig}` seeds the buffer from the signal
  and keeps it in sync thereafter. Without a binding, `value="..."` seeds
  it from a string literal only — a bare expression value (`value={x}`,
  no `bind:`) is treated as empty, because such values are meant to be
  pushed in from your own handler, not read back by the control.
- **Windowing**: when the value is wider than the laid-out control, the
  view slides to keep the caret visible. This happens against the real
  laid-out width, so it is correct even for flex-sized inputs.
- **`<textarea>`**: multi-line. Enter inserts a newline, Up/Down move
  between lines, and the value keeps its line structure (`white-space:
  pre`).

### Constraints

Attributes limit what editing can do:

- **`maxlength="20"`** caps the value at 20 runes. Typing past the cap is
  ignored; a paste is truncated to the room left.
- **`readonly`** blocks edits but still allows caret movement. A key that
  would have edited is swallowed (it doesn't leak to other handlers)
  rather than passed through.
- **`type="password"`** masks the display with bullets; the underlying
  value and the `input` event still carry the real text.
- **`disabled`** removes the control from the focus order entirely, and
  matches `:disabled` (enabled controls match `:enabled`) in CSS.

### Checkboxes and radios

`<input type="checkbox">` and `<input type="radio">` render a glyph and
fire a `change` event (and an `input` event) carrying
`Data["checked"]` (bool) and `Data["value"]` (the `value` attribute).
Bind the checked state with `bind:checked`:

```sumi
<script>
notify := sumi.New(false)
</script>
<input type="checkbox" bind:checked={notify} />
```

The event escape hatch is `onchange` reading `evt.Data["checked"].(bool)`.
Space toggles the focused checkable; Enter and a mouse click do too.
`bind:checked` seeds the initial state from the signal; without a binding,
`checked="true"` sets it. Radios group by `name`:
checking one unchecks the others with the same name across the tree, and
a checked radio does not toggle itself off. `:checked` selects the
checked state in CSS.

### Select

`<select>` shows the current option's label followed by `▾` and fires a
`change` with `Data["value"]` set to the selected option's `value`. Bind
the selected value with `bind:value`:

```sumi
<script>
plan := sumi.New("free")
</script>
<select bind:value={plan}>
	<option value="free">Free</option>
	<option value="pro">Pro</option>
</select>
```

The binding selects the option whose `value` matches the signal, so an
external `plan.Set("pro")` moves the selection. The event escape hatch is
`onchange` reading `evt.Data["value"].(string)`. The arrow keys move the
selection while the select is focused; a click advances it. Without a
binding, `selected="true"` marks the initial option.

## Focus and editing

Editing is always directed at the *focused* control. Only the focused
input receives typed keys, and its caret is shown only while it holds
focus — blur hides the caret. Tab and Shift-Tab move focus between
controls (disabled controls are skipped); a click focuses the control
under the cursor. Because typing into a focused input is a keydown
default action, it runs before the component's raw `handleKey` — a
component's `handleKey` won't see the letters typed into one of its
inputs. See [events](events.md) for the full dispatch order.

## Paste and clipboard

A bracketed paste into a focused input inserts the pasted text at the
caret, honouring `maxlength` (the paste is trimmed to fit). Selection and
copy/cut over the painted frame go through the terminal clipboard (OSC 52
where the terminal supports it); this is terminal-level behaviour rather
than a per-control attribute, covered in [terminals](terminals.md).

## Binding to a child component

The same `bind:` syntax also connects a parent signal to a child
*component* whose prop is a signal:

```sumi
<Field bind:count={clicks} />
```

Here the parent and child share one `*sumi.Signal` directly — a write on
either side is visible to both — rather than the control-value mirroring
that native `bind:` performs. See
[components](components.md#bind-and-cross-component-signal-flow).

## Limitations

A binding inside `{if}` or `{for}` content wires the update half only:
typing still writes the signal, but an external `Set` does not
re-project onto the control (the same limitation as dynamic state
attributes). Bindings on statically-placed controls reflect both
directions. A control cannot declare both a binding and a handler for
the binding's own event (`bind:value` + `oninput`) — that is a
generation error; use one or the other.
