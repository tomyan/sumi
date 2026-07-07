# Editor support

- `grammar/sumi.tmLanguage.json` — the TextMate grammar, source of
  truth for every consumer (the site's shiki highlighting, VS Code).
- `vscode/` — VS Code extension. `npm run package` copies the grammar
  in and builds the `.vsix`.
- `vim/` — classic vim syntax (vim + neovim). Install by adding this
  directory to your runtimepath, e.g. with lazy.nvim:
  `{ dir = "~/path/to/sumi/editors/vim" }`, or copy `ftdetect/` and
  `syntax/` into `~/.vim` / `~/.config/nvim`.

See the "Editor support" chapter on https://gosumi.dev for setup
including the language server.
