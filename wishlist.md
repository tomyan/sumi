# Sumi Wishlist

Features needed to support building full-featured conversational terminal applications.

Tags indicate which real-world applications require each feature:
- `[opencode]` — Go/Bubble Tea coding agent
- `[gemini]` — Google's Gemini CLI (TypeScript/Ink)

## Critical

### Color Support `[opencode]` `[gemini]`
- 16-color (bright variants: `bright-red`, `bright-cyan`, etc.)
- 256-color palette (`color-196` syntax)
- Truecolor 24-bit (`#ff0088` hex values)
- Graceful degradation based on terminal capability detection
- Perceptual color quantization for truecolor→256→16 fallback
- Adaptive colors: specify different values for light/dark terminal backgrounds `[opencode]` `[gemini]`
- Color blending / alpha compositing

### Layout `[opencode]` `[gemini]`
- Margin (all sides, with shorthand) `[opencode]` `[gemini]`
- min-height, max-width, max-height constraints `[opencode]` `[gemini]`
- flex-shrink, flex-basis `[gemini]`

### Text `[opencode]` `[gemini]`
- Truncation modes: end-truncate, middle-truncate, truncate with ellipsis `[opencode]` `[gemini]`
- text-align: left, center, right `[opencode]` `[gemini]`
- Raw pre-styled ANSI string rendering (passthrough element for externally-styled content) `[gemini]`

### Rendering
- Synchronized output (DEC 2026 BSU/ESU) to prevent flicker during redraws
- Virtual scrolling with viewport culling (only render/layout visible children in scroll containers) `[opencode]` `[gemini]`

### Markdown Rendering `[opencode]` `[gemini]`
- Component or library for rendering styled markdown in terminal: headings, bold/italic, lists, tables, links, code blocks with syntax highlighting
- Syntax highlighting via tree-sitter (already proven in hubcap project)

### Hyperlinks `[gemini]`
- OSC 8 clickable hyperlinks in text content

## Important

### Terminal Capabilities `[gemini]`
- Probe terminal at startup: DA1/DA2 device attributes, XTVERSION, DECRQM mode queries
- Detect color depth, keyboard protocol support, mouse mode support, synchronized output support
- Terminal identification (iTerm2, Ghostty, Kitty, VS Code, Alacritty, etc.) `[gemini]`

### Keyboard `[gemini]`
- Kitty keyboard protocol (CSI u) for unambiguous modifier detection `[gemini]`
- xterm modifyOtherKeys fallback
- Bracketed paste detection
- Priority-based key event routing (multiple handlers with priority levels) `[gemini]`

### Text Selection `[gemini]`
- Full-screen drag-to-select with mouse
- Word selection (double-click) and line selection (triple-click)
- Copy to clipboard on selection complete

### Clipboard
- General-purpose clipboard write (OSC 52) with platform fallbacks (pbcopy, wl-copy, tmux load-buffer)
- Clipboard read where supported

### Scrolling
- DECSTBM hardware scroll regions for efficient partial-screen scrolling
- Scroll-drain smoothing (cap scroll rate for animation feel)
- Nested / multi-area scroll coordination (innermost scrollable region scrolls first) `[gemini]`

### Animation `[gemini]`
- Shared animation clock (synchronized tick across all animated components) `[gemini]`
- Configurable frame interval with pause-on-blur `[gemini]`
- Frame-based sprite/ASCII animation (pre-defined frame sequences)
- Shimmer / sweep effects (time-based highlight across text)

### Cursor
- Cursor shape control: block, underline, bar (DECSCUSR)
- Blinking vs static variants

### Components `[opencode]` `[gemini]`
- Dialog/modal overlay (z-layered, dismissible, shadow backdrop) `[opencode]` `[gemini]`
- Selectable list (keyboard-navigable, used in dialogs and pickers) `[opencode]` `[gemini]`
- Spinner / loading indicator `[opencode]` `[gemini]`
- Tabs (switchable tab bar with content panels)
- ProgressBar (determinate and indeterminate)
- FuzzyPicker (filterable searchable list) `[gemini]`
- Divider (horizontal/vertical rule)
- Spacer (flexible space-filler, though achievable with flex-grow)
- Table (styled rows/columns with selection) `[opencode]`
- File browser (directory navigation with selection) `[opencode]` `[gemini]`
- Diff renderer (unified diff with line numbers, tree-sitter highlighting, add/remove coloring) `[opencode]` `[gemini]`
- Toast / transient notification `[gemini]`
- Gradient text (multi-color text spans) `[gemini]`

### Theming `[opencode]` `[gemini]`
- Runtime theme switching (swap color palette without restart) `[opencode]` `[gemini]`
- Theme object with named color tokens (primary, secondary, accent, error, text, background, border, etc.) `[opencode]` `[gemini]`
- Runtime theme detection (light/dark) via OSC 11 background color query `[opencode]` `[gemini]`
- Syntax highlighting theme integration (map tree-sitter scopes to theme colors) `[opencode]` `[gemini]`

### Search
- Search highlighting across scrollable content (current match + all matches, visually distinct)

### Concurrency `[opencode]` `[gemini]`
- Goroutine-safe signal updates (Set/Update from background goroutines triggering UI re-render)

### Image Rendering `[opencode]`
- Half-block (`▀`) image rendering with truecolor fg/bg (depends on truecolor support)

### Input Modes `[gemini]`
- Vim mode support (normal/insert mode with separate key handling)
- External editor launch ($EDITOR integration) `[opencode]`

## Nice to Have

### Text
- BiDi / RTL text reordering
- text-overflow: ellipsis as CSS property
- word-break control
- line-height (extra rows between lines)

### Rendering
- Style transition caching / interning for zero-allocation steady-state rendering
- Screen buffer pooling and reuse across frames

### Layout
- align-self per-child override
- flex-wrap
- order (visual reorder without changing DOM order)
- display: contents (unwrap container)
- visibility: hidden (takes space but renders blank)

### Interaction
- :hover pseudo-class (mouse motion tracking, DEC 1003)
- pointer-events control

### Modes
- Inline rendering mode (non-alt-screen, within terminal scrollback)

### Accessibility `[gemini]`
- Screen reader mode (alternative layout/output for assistive technology)
