# Editor support

Syntax highlighting, a language server, and file-type plumbing for
`.sumi` files. Everything below ships in the sumi repo under
`editors/`.

## VS Code

Install the **sumi** extension from the marketplace, or build it from
source:

```sh
cd editors/vscode
npm run package          # produces sumi-lang-<version>.vsix
code --install-extension sumi-lang-*.vsix
```

You get full highlighting: Go inside `<script>` and `{expressions}`,
CSS inside `<style>`, and the template's tags, `on*`/`bind:`
attributes, and `{if}`/`{for}` control tags.

## Vim / Neovim

The classic syntax plugin works in both editors with no build step.
Add `editors/vim` to your runtimepath — with lazy.nvim:

```lua
{ dir = "/path/to/sumi/editors/vim" }
```

or copy `ftdetect/` and `syntax/` into `~/.vim` (or
`~/.config/nvim`). Go and CSS regions use your existing Go/CSS
highlighting.

## Language server

`sumi lsp` speaks LSP over stdio and ships with the CLI (`brew
install tomyan/tap/sumi`). It reports parse and validation errors as
you type, completes element tags, CSS properties, and component
props, shows hover notes for CSS properties, provides document
symbols, and jumps to component definitions.

Neovim (with nvim-lspconfig):

```lua
vim.api.nvim_create_autocmd("FileType", {
	pattern = "sumi",
	callback = function()
		vim.lsp.start({ name = "sumi", cmd = { "sumi", "lsp" } })
	end,
})
```

VS Code LSP client wiring is planned for a future extension release;
the extension currently ships highlighting only.

## The site's highlighting

gosumi.dev highlights `.sumi` blocks with the same TextMate grammar
via shiki — `editors/grammar/sumi.tmLanguage.json` is the single
source of truth for every consumer.
