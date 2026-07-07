# Events

Sumi has two event layers. Element events (`onclick`, `oninput`, …) use a
DOM-style model with capture targets and bubbling. Raw input events (every
keypress, resize, signal, and frame) reach a single per-component handler.
Element events are dispatched first; whatever they don't consume falls
through to the raw handler.

## Element events

Write `on<type>={handler}` on an element. The value must be the
`{expr}` form — a bare Go expression, usually a function name. A string
value (`onclick="foo"`) is not wired.

```sumi
<script>
func save(evt *sumi.DOMEvent) {
	evt.PreventDefault()
}
</script>
<button onclick={save}>Save</button>
```

A handler declared with a parameter receives the `*sumi.DOMEvent`
directly. A zero-argument expression is called through a nil check, so
`onclick={maybeNil}` is safe when the expression can be nil.

### The DOMEvent

```go
type DOMEvent struct {
	Type   string         // "click", "keydown", "focus", "blur",
	                       // "input", "change", "paste", "toggle"
	Key    sumi.Event     // the underlying terminal event, when there is one
	Data   map[string]any // payload; keys depend on Type (below)
	Target *sumi.Input    // the deepest element on the dispatch path
}
```

Methods: `StopPropagation()` stops the event reaching handlers higher up
the path; `PreventDefault()` suppresses the built-in action that would
otherwise follow (see below). `Stopped()` and `DefaultPrevented()` report
each.

`Data` carries the payload for value-bearing events:

| Type              | `Data` keys                          |
|-------------------|--------------------------------------|
| `input`           | `value` (string), `cursor` (int)     |
| `change`, `input` on a checkbox/radio | `checked` (bool), `value` (string) |
| `change` on a select | `value` (string)                  |

Read them with a type assertion: `evt.Data["value"].(string)`.

### Dispatch and bubbling

An event is dispatched along the path from the tree root down to a
target element, then handlers run from the **deepest element upward**
(bubbling). `StopPropagation` on an inner handler keeps outer handlers
from seeing it:

```sumi
<div onclick={outer}>
	<button onclick={inner}>bump</button>
</div>
```

Clicking the button runs `inner`, then `outer` — unless `inner` calls
`evt.StopPropagation()`.

What produces each event:

- **click** — a left mouse press hit-tests to the deepest element under
  the cursor. Enter (and Space, on checkables) also synthesize a click on
  the focused element.
- **keydown** / **paste** — dispatched along the path to the *focused*
  element only.
- **focus** / **blur** — delivered to the element directly; these do not
  bubble.
- **input** — fires after a focused text input's value actually changes.
- **change** — fires when a checkbox/radio toggles or a select moves.
- **toggle** — fires when a `<details>` opens or closes.

A click outside an open `<dialog>` is dropped: an open dialog traps
focus and captures clicks.

### Default actions

Built-in behaviour runs *after* element dispatch, and only if no handler
called `PreventDefault`. In order, the runtime tries to:

1. close the open dialog on Escape,
2. move focus on Tab / Shift-Tab,
3. type the key into the focused `<input>` / `<textarea>`,
4. move the selection on a focused `<select>` with the arrows,
5. activate the focused element: Enter fires a click on anything with a
   click handler, an `<a href>`, a checkable, or a `<summary>`; Space
   toggles checkables only (it types into text inputs and is otherwise
   ignored).

Calling `evt.PreventDefault()` in a `keydown` or `click` handler cancels
whichever of these would have applied.

## Raw events

Beyond element events, a component can handle the raw input stream:
keys, resize-driven frames, OS signals, animation frames, paste, focus
and blur. Declare a function named `handleKey` taking a `sumi.Event`; it
becomes the component's `OnEvent`. The wiring is by name — only
`handleKey` is picked up. The conventional `onkey="handleKey"` attribute
on the root documents the intent but is not what does the wiring.

```sumi
<script>
func handleKey(evt sumi.Event) {
	if evt.Kind == sumi.EventSignal { sumi.Quit(); return }
	if evt.Ctrl && evt.Rune == 'c' { sumi.Quit(); return }
	if evt.Rune == 'j' { /* ... */ }
}
</script>
```

`handleKey` sees an event only after element dispatch and default
actions. An event a default action consumed (a Tab, a keystroke typed
into a focused input, an Enter that activated a button) does not reach
it — so a component with focusable children receives only the keys they
leave unhandled, while a component with none (like a plain counter) sees
everything.

### The Event

```go
type Event struct {
	Kind      sumi.EventKind
	Rune      rune              // set for EventKey
	Ctrl, Shift, Alt bool       // modifiers
	Special   sumi.SpecialKey   // set for EventSpecial
	PasteText string            // set for EventPaste
	Signal    syscall.Signal    // set for EventSignal
	// Mouse, Scheme, CursorRow/Col for the remaining kinds
}
```

`Kind` is one of `EventKey`, `EventSpecial`, `EventMouse`,
`EventSignal`, `EventPaste`, `EventFocus`, `EventBlur`, `EventFrame`.
A printable key is `EventKey` with `Rune` set; Ctrl+letter is `EventKey`
with `Ctrl` true and `Rune` the letter.

`SpecialKey` is a string. The prelude re-exports the common ones —
`sumi.KeyUp`, `KeyDown`, `KeyLeft`, `KeyRight`, `KeyHome`, `KeyEnd`,
`KeyPgUp`, `KeyPgDn`, `KeyTab`, `KeyShiftTab`, `KeyEnter`, `KeyEscape`,
`KeyBackspace`, `KeyDelete`. Function keys are not aliased; match them by
value: `evt.Special == "f1"`. Kitty-protocol key disambiguation (which
lets a terminal report modifiers the legacy encoding can't) is covered in
[terminals](terminals.md).

## Quitting

`sumi.Quit()` ends the app from any handler. Two things quit by default
without your code:

- **Ctrl+C** — the runtime quits after `handleKey` runs, via the app's
  exit chords (default `["ctrl+c"]`). Override with `RunOptions.ExitOn`
  (see [lifecycle](lifecycle.md)); an explicit list replaces the default.
- Nothing else. An OS signal (`SIGINT`/`SIGTERM`) arrives as an
  `EventSignal` at `handleKey` and is ignored unless you handle it — the
  convention in every example is
  `if evt.Kind == sumi.EventSignal { sumi.Quit() }`.
