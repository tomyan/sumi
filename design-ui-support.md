# Multi-Target UI Support

Sumi's component model — single-file components with script/style/template sections, Solid-style signals, and CSS — is not inherently terminal-specific. This document explores how Sumi can target non-terminal surfaces, particularly mobile, and the architectural decisions behind the chosen approach.

## Design Principle: LLM-Native Authoring

Sumi's CSS is real CSS. The template syntax mirrors Svelte/Vue SFCs. The script block is valid Go. These are not novel DSLs — they are syntaxes that frontier LLMs have been trained on billions of times. An AI agent generating Sumi components is working with patterns it already knows deeply.

This matters because the future of UI development is increasingly agentic. A framework's LLM-friendliness — how reliably an AI can generate correct components — is at least as important as its ergonomics for human developers. Sumi's use of standard CSS and familiar SFC structure makes it a natural target for AI-generated UI code.

## Target Surfaces

```
.sumi components
    ├── terminal (today) — Go renders to cell buffer
    ├── mobile (planned) — Go sends DOM patches to a WebView
    ├── desktop (future) — same WebView approach
    └── web (future) — same architecture, no native shell needed
```

## Options Considered

### Option 1: WebView (Web Technologies in a Native Shell)

The UI renders as HTML/CSS in a platform WebView (WKWebView on iOS, WebView on Android). Go runs natively and controls the DOM remotely via a message-passing bridge.

**Advantages:**
- CSS is CSS — no translation, no custom renderer. The browser engine handles layout, text rendering, scrolling, selection, accessibility.
- Text selection, input, scroll physics, accessibility — all free from the browser.
- Smallest implementation scope by a wide margin.
- CSS developer tools (Safari Web Inspector, Chrome DevTools) work for debugging.
- Sumi's CSS media queries work natively — terminal-specific rules ignored, screen rules applied.

**Disadvantages:**
- WebView apps can feel like WebView apps — a perceptible quality gap vs truly native UI.
- Dependent on platform WebView versions and rendering quirks.
- Apple has historically scrutinised WebView-only apps, though Tauri-style apps generally pass review.

### Option 2: Custom Canvas (Flutter Model)

A GPU rendering surface (Metal on iOS, Vulkan/OpenGL on Android) draws everything — text, borders, backgrounds, scrolling. Full control over every pixel.

**Advantages:**
- Pixel-perfect cross-platform rendering. Identical output everywhere.
- One rendering tree — no sync problems between native views and a custom layout engine.
- Full control over visual effects, animations, custom drawing.
- CSS maps directly to rendering commands without platform-specific interpretation.

**Disadvantages:**
- Text rendering is genuinely hard: font loading, glyph shaping (HarfBuzz), text layout, line breaking, bidirectional text, emoji, variable fonts. This is years of work to get right.
- Text selection requires custom implementation: hit-testing, selection handles, long-press gesture detection, OS copy/share menu integration. Each platform has specific expected behaviours.
- Accessibility must be built explicitly — a semantic tree exposed to VoiceOver/TalkBack via platform APIs. No native views for the screen reader to discover.
- Platform conventions (scroll bounce on iOS, edge glow on Android, haptic feedback) must be reimplemented.
- Text input and IME (input method editors for CJK languages) require a hidden native text field as a keyboard bridge.
- App size increases by ~15MB+ for the rendering engine.
- Massive scope — this is building a browser rendering engine.

### Option 3: Native Views (React Native Model)

The layout engine computes positions and sizes. Each `<box>` becomes a UIView (iOS) / ViewGroup (Android). Each `<text>` becomes a UILabel / TextView. The platform renders the actual pixels.

**Advantages:**
- Native look and feel. Buttons look like platform buttons. Accessibility works automatically.
- Text rendering, selection, input, IME — all handled by the platform.
- Smaller app size — no rendering engine shipped.
- Platform features (Dark Mode, Dynamic Type, safe area insets) work automatically.

**Disadvantages:**
- Two view trees to keep in sync (Sumi component tree and platform view tree). This is React Native's biggest source of bugs.
- Platform inconsistency. UIView and Android View don't behave identically — platform-specific fixups accumulate.
- CSS subset doesn't map cleanly to native view properties. Each CSS property needs platform-specific interpretation, and edge cases multiply.
- Performance ceiling from bridge overhead. Every component crosses the Go→native boundary. Scrolling large lists requires virtualisation.
- Limited visual control — custom animations, gradients, or effects require fighting the native view system.

### Option 4: Hybrid (Custom Canvas + Native Text Views)

Custom canvas draws layout containers (backgrounds, borders, padding). Native text views handle all text rendering and input. The layout engine positions a mix of canvas-drawn elements and native views.

**Advantages:**
- Native text quality for display text. Native input handling with IME, autocorrect, accessibility.
- Full visual control for containers and layout.
- Avoids the hardest rendering problems (text shaping, input handling).

**Disadvantages:**
- Compositing two rendering systems. Canvas draws behind native views — can't draw over text.
- Native text views have limited inline styling. Bordered inline code spans, rounded-corner code blocks, and custom block elements within text don't map to attributed string APIs.
- Z-ordering constraints between canvas and native layers.
- Still need platform-specific text view configuration.

The hybrid model was considered with a further refinement: WYSIWYG editing is not required. Editable text is monospace with syntax highlighting (a native monospace text view with attributed strings from tree-sitter tokenisation). Display text is styled but read-only.

However, even for read-only display, native text views can't express all the block-level structure needed (bordered code blocks, tables, diff views embedded in flowing content). This pushes back toward either full custom canvas (with its text rendering complexity) or the WebView approach.

## Chosen Approach: WebView with Native Go Bridge

The WebView approach (Option 1) provides the best trade-off between capability and implementation scope. Sumi's CSS is already real CSS — running it in an actual browser engine means zero translation, zero custom rendering, and access to the full web platform for free.

### Architecture

```
Go (native, compiled to static C library via cgo)
├── signals runtime
├── component lifecycle
├── event handling
├── app logic
├── computes DOM patches (insert/update/remove nodes)
│
├── native message API ──► WebView
│                           ├── thin JS bridge (~200 lines)
│                           ├── applies DOM patches to real DOM
│                           ├── real HTML elements
│                           ├── real CSS (Sumi stylesheets loaded directly)
│                           └── sends events back via message API
│
◄── native message API ◄── JS: postMessage(event data)
```

### Native Shell

```
Go code → cgo → static C library (.a)
                    ↓
        ┌───────────┴───────────┐
        │                       │
   Swift shell (iOS)      Kotlin shell (Android)
   - links Go static lib  - links Go static lib via JNI
   - creates WKWebView    - creates WebView
   - bridges Go ↔ JS      - bridges Go ↔ JS
```

The Swift and Kotlin shells are thin and generic — they create a WebView, set up the message bridge, and forward events. This is a library that any Sumi app links against, not per-app generated code. All app-specific logic is in Go.

### Bridge Communication

The bridge uses **platform-native message-passing APIs**, not JavaScript eval:

**iOS:**
- Go → WebView: `callAsyncJavaScript` with typed arguments, or `WKUserScript` injection
- WebView → Go: `WKScriptMessageHandler` via `window.webkit.messageHandlers.sumi.postMessage(data)`

**Android:**
- Go → WebView: `WebMessagePort` / `postWebMessage` — bidirectional message channel
- WebView → Go: `@JavascriptInterface` exposing native methods, or `WebMessagePort`

No `evaluateJavaScript` string construction. Data passed directly through typed message channels.

### Latency

The Go process and WebView are in the same process on the same device. Bridge latency is **~0.1-0.5ms** per message. Batching multiple DOM patches into one message keeps total round-trip well under **1ms**. A 60fps frame budget is 16.6ms — there is ample headroom for hundreds of DOM patches per frame.

### What Go Sends to the WebView

- Insert node (type, id, parent ID, attributes, CSS classes)
- Update text content (node ID, text)
- Update attribute (node ID, attribute name, value)
- Remove node (node ID)
- Update CSS classes (node ID, class list)

### What the WebView Sends to Go

- Click / touch (target node ID, coordinates)
- Key event (key, modifiers)
- Input value changed (node ID, value)
- Scroll position changed (node ID, offset)
- Resize (width, height)
- Focus / blur (node ID)

### CSS Handling

Sumi stylesheets are loaded by the WebView as regular CSS files. No translation or interpretation at runtime. The same stylesheet works on terminal and web — media queries select target-appropriate rules:

```css
.container {
    display: flex;
    padding: 1rem;
    gap: 0.5rem;
}

@media (display-mode: terminal) {
    .container {
        border: single;
        border-color: cyan;
    }
}

@media (display-mode: screen) {
    .container {
        border: 1px solid cyan;
        border-radius: 4px;
    }
}
```

The terminal runtime evaluates its media queries and interprets `border: single` as box-drawing characters. The browser evaluates its media queries and applies standard CSS borders. Same stylesheet, different rendering target.

### Template Mapping

| Sumi element | Terminal | WebView |
|---|---|---|
| `<box>` | Layout node in cell grid | `<div>` |
| `<text>` | Styled character spans | `<span>` |
| `{if}` / `{for}` | Conditional/repeated nodes in tree | DOM insert/remove via bridge |
| `<slot:name />` | Slot placeholder in component tree | DOM slot insertion point |

### Text Input

Editable text uses a native monospace text view rendered by the browser — a `<textarea>` or `contenteditable` element with a monospace font. Syntax highlighting is applied via tree-sitter tokenisation producing styled `<span>` elements or CSS classes. No WYSIWYG — input is code-style, display is fully styled.

### Device API Plugins

Native device features (camera, GPS, haptics, notifications) are implemented as Go packages that call through to platform APIs via cgo. They are not part of the WebView — plugins communicate with the Go runtime directly, and results are reflected in the UI via signal updates that generate DOM patches.

```go
// Plugin interface
type CameraPlugin interface {
    TakePhoto(opts PhotoOptions) (*Photo, error)
}

// Usage in a .sumi component's script block
camera := plugin.Get[CameraPlugin]()
photo := signal.New[*Photo](nil)

func takePhoto() {
    p, _ := camera.TakePhoto(PhotoOptions{})
    photo.Set(p)
}
```

### Comparison with Existing Approaches

This architecture is similar to:

- **Tauri v2** — Rust backend, WebView frontend, native message bridge. Same model, Go instead of Rust.
- **Phoenix LiveView / Hotwire** — server controls the DOM, browser is a thin rendering layer. Except the "server" is a Go process on the same device, so latency is microseconds, not milliseconds.

### What This Gives Up

- WebView apps don't feel 100% native. There is a perceptible (if small) quality gap.
- Dependent on platform WebView versions and rendering behaviour.
- Apple has historically scrutinised WebView-only apps, though Tauri-style apps with native backend logic generally pass review.
- Cannot use platform-native UI controls (UIKit/SwiftUI/Material). The UI is web-rendered.

### What This Gets

- Zero custom rendering code. The browser is the renderer.
- Full CSS support. Sumi's CSS works unchanged.
- Text rendering, selection, accessibility, scroll physics, input, IME — all free.
- Debugging via web inspector.
- A single Go codebase produces terminal apps, mobile apps, desktop apps, and web apps.
- The component model that LLMs can generate most reliably — standard CSS, HTML-like templates, familiar SFC structure.

## Developer Experience: `sumi dev`

A Vite-style development server that provides hot module replacement across both web and terminal targets simultaneously.

### Usage

```
sumi dev                  # starts web preview + terminal preview
sumi dev --web            # web preview only
sumi dev --terminal       # terminal preview only
```

### Architecture

```
sumi dev
├── file watcher (fsnotify on .sumi files)
├── incremental compiler (only recompiles changed files)
│
├── web target
│   ├── HTTP server (serves HTML + CSS + bridge JS)
│   ├── WebSocket (pushes updates on recompile)
│   └── CSS hot-swap (replace stylesheet, no remount)
│
└── terminal target
    ├── terminal preview (existing tool)
    ├── hot-reload on recompile
    └── CSS changes trigger repaint only
```

### Update Granularity

The `.sumi` compiler parses script, style, and template as separate sections. It diffs at the section level on each file change and emits the minimum update:

| What changed | Update type | Signal state | Speed |
|---|---|---|---|
| Style block only | Hot-swap stylesheet | Preserved | Instant |
| Template only | Re-mount render function | Preserved | Fast |
| Script block | Full component re-mount | Resets | Fast (single component) |

Only the changed component is affected — the rest of the app continues running. This matches how Vite handles Svelte/Vue SFCs.

### Web Preview

The web preview server is also the development environment for mobile UI work. You develop in a browser with real CSS rendering, browser DevTools for inspection, and instant feedback. The same components deploy to mobile WebView unchanged.

- Edit `.sumi` file → file watcher detects change → compiler runs → update pushed via WebSocket
- CSS changes: replace `<link>` tag, no DOM changes, no component remount
- Template changes: DOM patches sent via WebSocket, component re-mounts with preserved signal state
- Script changes: component fully re-mounts, signals reinitialise

### Terminal Preview

The existing terminal preview tool (with embedded nvim editors and VT100 parser) receives the same incremental updates:

- CSS changes: re-evaluate styles, repaint affected cells
- Template changes: rebuild affected subtree, re-render
- Script changes: re-mount component

### Dual Preview

Running both targets simultaneously from a single `sumi dev` process allows side-by-side development — edit once, see the result in both terminal and browser. The file watcher and compiler are shared; only the output targets differ.

This is particularly useful for building components that need to work across targets, verifying that media queries select the right rules for each surface, and ensuring layout works at both terminal cell granularity and browser pixel granularity.
