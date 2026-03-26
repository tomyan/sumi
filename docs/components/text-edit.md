# textedit

Fundamental editable text primitive. Handles editing, cursor, selection, undo, clipboard, and mouse interaction. Renders visible text with selection highlighting. No chrome, no scrollbar, no indicators.

This is the raw editing surface — the equivalent of `contenteditable` in web terms.

**Tier**: fundamental (lowercase, always available)

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | `string` | `""` | The text content. Use `bind:value` for two-way binding. |
| `placeholder` | `string` | `""` | Dimmed text shown when value is empty. |
| `inputType` | `string` | `""` | Set to `"password"` to mask characters with `*`. |
| `maxlength` | `int` | `0` | Maximum character count. 0 means unlimited. |
| `readonly` | `bool` | `false` | Blocks all editing. Cursor movement, selection, and copy still work. |
| `strip` | `bool` | `false` | Trim leading and trailing whitespace when the input loses focus. |

## Bound State

These can be bound by the parent with `bind:` to read or coordinate with the internal state.

| Name | Type | Description |
|------|------|-------------|
| `viewOffset` | `int` | Horizontal scroll position (first visible character index). |
| `focused` | `bool` | Whether the component currently has focus. |

## Behavior

### Cursor

Blinking hardware cursor positioned within the visible text. Hidden when a selection is active.

### Editing

- Character insert at cursor position
- Backspace / Delete (single character and word-level with Ctrl)
- Readline keybindings: Ctrl+A (start), Ctrl+E (end), Ctrl+F/B (forward/back), Ctrl+D (delete), Ctrl+K (kill to end), Ctrl+U (kill to start), Ctrl+W (kill word back), Ctrl+T (transpose), Ctrl+Y (yank)
- Alt+F/B (word forward/back), Alt+D (kill word forward)
- Undo via Ctrl+/ (100-entry stack)

### Selection

- Shift+Left/Right: character-level selection
- Shift+Ctrl+Left/Right: word-level selection
- Shift+Home/End: select to start/end
- Ctrl+A: select all
- Any non-shift movement clears selection
- Typing with active selection replaces it
- Backspace/Delete with selection removes selected range

### Clipboard

- Ctrl+C: copy selection to kill buffer + OSC 52 system clipboard (does not stop propagation when no selection, allowing Ctrl+C to bubble for quit)
- Ctrl+X: cut (copy + delete selection, respects readonly)
- Ctrl+Y: yank from kill buffer (respects maxlength)

### Paste

- Handles `EventPaste` (bracketed paste) — inserts text at cursor, replaces selection if active, respects maxlength and readonly

### Mouse

- Click: position cursor
- Click + drag: character-level selection
- Double-click: select word (stops at word boundary, not trailing space)
- Double-click + drag: word-level selection
- Shift+click: extend selection

### Focus

- Receives `EventFocus` / `EventBlur` from the focus system
- Tracks focused state internally
- Click outside (unhandled mouse press) defocuses via framework

### Strip on Blur

When `strip` is true and the input loses focus, leading and trailing whitespace is removed from the value. The cursor is clamped if it would be past the end.

### View Offset

Automatically scrolls to keep the cursor visible within the component's width (determined via `$self(width)`).

## Rendering

Renders a single row of text split into four segments:

```
[pre-selection][selection (inverse)][post-selection][placeholder (dim)]
```

When no selection is active, all text appears in the pre-selection segment. The component pads with spaces to fill its width.

## Usage

```html
<textedit bind:value={name} placeholder="Enter name" />
```

```html
<textedit bind:value={password} inputType="password" maxlength={32} />
```

```html
<textedit bind:value={query} bind:viewOffset={offset} bind:focused={hasFocus} />
```
