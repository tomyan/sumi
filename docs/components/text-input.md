# sumi:TextInput

A styled single-line text input composed from `<textedit>` and `<scrollbar>`. Adds visual chrome (brackets, scroll indicators).

**Tier**: standard library (`sumi:` prefix, always available)

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | `string` | `""` | The text content. Use `bind:value` for two-way binding. |
| `placeholder` | `string` | `""` | Dimmed text shown when value is empty. |
| `inputType` | `string` | `""` | Set to `"password"` to mask characters with `*`. |
| `maxlength` | `int` | `0` | Maximum character count. 0 means unlimited. |
| `readonly` | `bool` | `false` | Blocks all editing. Cursor movement, selection, and copy still work. |
| `strip` | `bool` | `false` | Passed through to `<textedit>`. Trims whitespace on blur. |

## Appearance

```
[< hello world          >]
  --##------------------
```

- `[` `]` — brackets framing the input
- `<` `>` — scroll indicators (dim), shown when content overflows in that direction, space when not
- Scrollbar row underneath, visible only when focused and content overflows

## Composition

```html
<box>
    <box direction="row">
        <text>[</text>
        <text class="indicator">{leftIndicator()}</text>
        <textedit
            bind:value={value}
            bind:viewOffset={viewOffset}
            bind:focused={focused}
            placeholder={placeholder}
            inputType={inputType}
            maxlength={maxlength}
            readonly={readonly}
            strip={strip}
        />
        <text class="indicator">{rightIndicator()}</text>
        <text>]</text>
    </box>
    <scrollbar
        content-size={len(value)}
        view-size={contentW}
        bind:offset={viewOffset}
        visible={focused && len(value) > contentW}
    />
</box>
```

## Behavior

All editing, selection, cursor, clipboard, and mouse behaviour is provided by `<textedit>`. TextInput adds:

### Scroll Indicators

- Left indicator: `<` when `viewOffset > 0`, otherwise space
- Right indicator: `>` when content extends beyond the visible area, otherwise space
- Both rendered with `dim` style

## Usage

```html
<sumi:TextInput bind:value={username} placeholder="Username" />
```

```html
<sumi:TextInput bind:value={password} inputType="password" maxlength={64} />
```

```html
<sumi:TextInput bind:value={search} placeholder="Search..." strip={true} />
```
