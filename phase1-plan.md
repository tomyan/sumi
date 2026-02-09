# Phase 1: Static Text Rendering — Elephant Carpaccio Breakdown

Goal: `.sumi` file with `<text>` → compiler → `.go` file → binary that renders text on alternate screen.

## Slice 1: .sumi file section splitter
- Parser that splits a `.sumi` file into optional sections: `<script>`, `<style>`, and template (everything else)
- When no script/style blocks exist, entire content is template
- TDD: test with template-only file, file with all sections, empty file
- **Value**: we can read .sumi files and know what's in each section

## Slice 2: Template parser — `<text>` elements
- Parse `<text>Hello</text>` into an AST (TextElement with string content)
- Handle self-closing, whitespace, multiple sibling text elements
- TDD: single text, multiple texts, nested content (error for now), whitespace handling
- **Value**: we understand the structure of a template

## Slice 3: Runtime — alternate screen + text rendering
- `runtime/render` package with alternate screen enter/exit
- Screen buffer (2D grid of cells)
- Render plain text string at a position in the buffer
- Flush buffer to terminal (write changed cells)
- TDD: buffer creation, writing text to buffer, diff detection
- **Value**: we can put text on screen and clean up properly

## Slice 4: Code generator — Go source from template AST
- Takes template AST, produces Go source code string
- Generated code imports runtime, enters alternate screen, renders text, waits for input, exits
- TDD: generate from single TextElement, verify output compiles (go/parser check)
- **Value**: we can turn parsed templates into runnable Go

## Slice 5: CLI `sumi generate` — end to end
- `cmd/sumi/main.go` with `generate` subcommand
- Reads `.sumi` file(s), runs parser + codegen, writes `_sumi.go` file
- Integration test: `.sumi` file → `sumi generate` → `.go` file exists and compiles
- **Value**: the full pipeline works end-to-end

## Dependencies
```
Slice 1 (section splitter) ──→ Slice 2 (template parser) ──→ Slice 4 (codegen) ──→ Slice 5 (CLI)
                                                               ↑
Slice 3 (runtime) ─────────────────────────────────────────────┘
```

Slices 1+2 and Slice 3 can be built in parallel.
Slice 4 depends on AST types from Slice 2 and runtime API from Slice 3.
Slice 5 ties everything together.
