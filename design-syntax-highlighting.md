# Design: syntax highlighting, editor support, LSP, skill

Decided with Tom 2026-07-07: targets are (1) the site via shiki (same
lib as svelterm-site — already in use, `.sumi` currently aliased to
html), (2) VS Code with marketplace publish (Tom runs the publish step
with his publisher account), (3) `sumi lsp` FULL v1 — diagnostics,
completion, hover, document symbols, go-to-definition, (4) classic vim
syntax file (vim + nvim, no treesitter for now), (5) a Claude skill,
(6) an "Editor support" docs chapter.

## One grammar, many consumers

`editors/grammar/sumi.tmLanguage.json` — TextMate grammar, scope
`source.sumi`:

- `<script>` region → embedded `source.go`
- `<style>` region → embedded `source.css`
- template: tags (known HTML + Component tags), attributes
  (`bind:value`, `on*={...}`, `class`, expression attrs `{expr}` →
  embedded `source.go`), text interpolation `{expr}` → `source.go`
- control tags `{if expr}`, `{else}`, `{else if expr}`, `{for x in y
  key=expr}`, `{/if}`, `{/for}` — keyword scopes, expr → `source.go`

Consumers: shiki on the site (custom lang registration, drop the
`sumi→html` alias; sync-docs.mjs copies the grammar into
`src/lib/`), and the VS Code extension (same file by path).
Validation: node smoke script in sumi-site runs shiki `codeToTokens`
over doc samples and asserts scopes for control tags, embedded Go,
embedded CSS.

## VS Code — `editors/vscode`

Language contribution (`.sumi` ext), the shared grammar,
`language-configuration.json` (comment `<!-- -->` in template — NOTE
sumi has no comment syntax of its own; brackets, autoclose pairs).
`vsce package` → `.vsix` committed workflow, marketplace publish
prepped (`publisher: "tomyan"` placeholder — confirm against the
account Tom creates). LSP client wiring in the extension comes later
(v1 ships grammar-only; the LSP is usable from nvim first).

## `sumi lsp` — full v1

Subcommand of the existing binary (ships with brew, no version skew).
Hand-rolled stdio JSON-RPC — Content-Length framing + typed structs
for exactly the methods used; no new module deps (matches repo ethos).

- **Diagnostics**: didOpen/didChange → parse + the generate-path
  validation. Parser gets a typed `ParseError{Offset int}` (today
  errors are strings with byte offsets); offset → UTF-16 line/col.
  Validation/codegen errors without positions surface as whole-file
  diagnostics (line 0) in v1.
- **Completion**: element tags (UA set + project components), CSS
  properties/values inside `<style>` (from `runtime/css`
  `supportedProperties` + value tables), `on*`/`bind:` attrs,
  component prop names (two-pass registry over the file's dir).
- **Hover**: CSS property → support note (generated from the same
  tables; reference.md stays the human matrix), element → UA note.
- **documentSymbol**: script decls (state/derived/funcs via the
  existing script parser) + template outline (components/ids).
- **definition**: component tag → its `.sumi` file; `on*={handler}`
  → the func decl in script.

Testing: table-driven over fixture `.sumi` files; JSON-RPC codec
round-trip tests; each feature = its own slice (red/green/commit).

## Vim — `editors/vim`

`ftdetect/sumi.vim` (`au BufRead,BufNewFile *.sumi setf sumi`) +
`syntax/sumi.vim`: `syn include` of `@GO` (syntax/go.vim), `@CSS`,
html-ish tag/attr matches, sumi control-tag keywords, `{...}` Go
regions. Docs get an nvim-lspconfig snippet pointing at `sumi lsp`.

## Skill + docs

`skills/sumi/SKILL.md` in-repo (orientation for writing .sumi:
element/CSS surface, gotchas from MEMORY patterns, pointer to
docs/reference.md; modelled on the svelterm skill) + installed copy at
`~/.claude/skills/sumi`. New `docs/editors.md` chapter → site.

## Slices (in order)

1. Grammar + shiki smoke test → site wiring → deploy (visible win).
2. VS Code extension package + local vsix; publish steps handed off.
3. Vim syntax + ftdetect.
4. LSP: core+diagnostics → completion+hover → symbols+definition.
5. Skill; docs chapter; site deploy; plan/memory updates.
