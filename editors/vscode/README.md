# sumi for VS Code

Syntax highlighting for [sumi](https://gosumi.dev) single-file
components (`.sumi`): Go in `<script>`, CSS in `<style>`, and the
HTML template with `{expr}` interpolation and `{if}`/`{for}` control
tags.

Install sumi itself with `brew install tomyan/tap/sumi` — see the
[getting started guide](https://gosumi.dev/docs/getting-started).

## Development

The grammar's source of truth is `editors/grammar/sumi.tmLanguage.json`
in the [sumi repo](https://github.com/tomyan/sumi); `npm run package`
copies it in and builds the `.vsix`.
