# Phase 7: Responsive Design — Elephant Carpaccio Breakdown

Goal: Use actual terminal dimensions, re-render on resize, and provide `$env(width)` / `$env(height)` reactive variables for templates.

Target demo:
```html
<script>
width := $env(width)
height := $env(height)
</script>

<style>
.header { border: single; justify: center; }
.dims { color: yellow; bold: true; }
</style>

<box class="header">
    <text>Terminal: {width}x{height}</text>
    <text class="dims">Resize to see this update!</text>
</box>
```

## Current State

- Layout + buffer hardcoded to 80×24
- No terminal size detection at runtime
- No SIGWINCH signal handling
- No `$env` rune in the script parser
- Event loop only reads keypresses (blocking `ReadKey`)

## Slice 7.1: Runtime terminal size + dynamic layout

Replace hardcoded 80×24 with actual terminal dimensions.

- Add `runtime/term/term.go` with `GetSize() (width, height int)`
  - Uses `golang.org/x/term.GetSize()` or TIOCGWINSZ ioctl
  - Falls back to 80×24 if detection fails
- Update codegen: generated `doRender` calls `term.GetSize()` for dimensions
- Generated code: `w, h := term.GetSize(fd)` then `layout.Layout(root, w, h)` and `render.NewBuffer(w, h)`
- TDD: unit test for GetSize (mock/fallback), codegen test for term.GetSize in output

## Slice 7.2: SIGWINCH handling + re-render on resize

Terminal resize triggers re-render without waiting for keypress.

- Add `runtime/term/resize.go` with `WatchResize(fd int) <-chan struct{}`
  - Listens for SIGWINCH, sends on channel
- Update codegen event loop: select on both key input and resize channel
  - Key input needs to become non-blocking or run in a goroutine
  - Use goroutine for ReadKey, select on keyChan and resizeChan
- On resize: set dirty=true, re-render
- TDD: codegen test for select-based event loop, resize channel usage

## Slice 7.3: $env parsing in script block

Parse `width := $env(width)` — similar to $state/$prop.

- Add `EnvDecl` type: `{Name string, Key string}`
- Add `EnvDecls []EnvDecl` to `Script`
- Parse `name := $env(key)` — key is an unquoted identifier (width, height)
- $env variables are NOT reactive for assignments (read-only)
- TDD: parse single env, multiple envs, mixed with state/prop

## Slice 7.4: $env codegen

Generated code initializes $env variables from terminal state and updates on resize.

- `width := $env(width)` → generates `width, height := term.GetSize(fd)` (width/height come as pair)
- On resize: update width/height variables, set dirty
- Template expressions `{width}` and `{height}` work like state variables
- TDD: codegen produces term.GetSize, env vars used in content expressions

## Slice 7.5: E2E responsive demo

- Create `examples/responsive/` with a terminal-size-aware layout
- Shows current dimensions, adapts layout on resize
- Verify: compiles, runs, responds to terminal resize

## Dependencies
```
Slice 7.1 (term size) → Slice 7.2 (SIGWINCH) → Slice 7.4 ($env codegen) → Slice 7.5 (E2E)
                                                 Slice 7.3 ($env parse) ↗
```
