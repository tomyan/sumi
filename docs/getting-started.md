# Getting started

Sumi is a declarative TTY framework for Go. You write `.sumi` single-file
components — a `<script>` of Go, a `<style>` block of CSS, and an HTML
template — and the `sumi` tool compiles them to plain Go source that renders to
a terminal cell grid.

This chapter takes you from an empty directory to a running, hot-reloading app.

## Installing the CLI

Sumi is not yet published as a Go module, so install the `sumi` binary from a
checkout of the repository:

```
git clone https://github.com/tomyan/sumi
cd sumi
go install ./cmd/sumi
```

That puts `sumi` on your `PATH` (in `$(go env GOPATH)/bin`). Every scaffolded
app depends on the framework through a `replace` directive pointing back at this
checkout, so keep it around — see [Scaffolding an app](#scaffolding-an-app).

## Scaffolding an app

```
sumi init myapp
```

`sumi init` refuses a non-empty directory. It writes three files, runs code
generation, and runs `go mod tidy`:

- `app.sumi` — your component (walked through below).
- `main.go` — the entry point.
- `go.mod` — the module file.

The module path defaults to `example.com/<dir>`; pass `--module you/app` to set
your own. The generated `go.mod` requires the framework and redirects it to your
local checkout:

```
module example.com/myapp

go 1.25

require github.com/tomyan/sumi v0.0.0

replace github.com/tomyan/sumi => /path/to/your/sumi/checkout
```

The checkout path is located automatically: `sumi init` walks up from the
current directory looking for sumi's own `go.mod`, or reads the `SUMI_PATH`
environment variable, or takes `--sumi-path`. If none resolve, it stops and
tells you to set one.

`main.go` is small and stable — you rarely touch it:

```go
package main

import "github.com/tomyan/sumi/runtime/tui"

//go:generate sumi generate .

func main() {
	tui.Run(NewApp(AppProps{}))
}
```

The `//go:generate` line means `go generate` regenerates the compiled Go from
your `.sumi` files; `sumi generate .` does the same directly.

## The scaffold component

`app.sumi` has three sections. Here is the whole file, then a walk through each
part.

```
<script>
count := sumi.New(0)

func increment(evt *sumi.DOMEvent) {
	count.Update(func(n int) int { return n + 1 })
}

func handleKey(evt sumi.Event) {
	if evt.Kind == sumi.EventSignal { sumi.Quit(); return }
	if evt.Rune == 'q' || (evt.Ctrl && evt.Rune == 'c') { sumi.Quit(); return }
}
</script>

<style>
h1 {
	color: cyan;
}
button:focus {
	color: yellow;
}
.hint {
	opacity: dim;
}
</style>

<div onkey="handleKey">
	<h1>Hello, sumi</h1>
	<p>You have pressed the button <strong>{count}</strong> times.</p>
	<button onclick={increment}>Press me</button>
	<div class="hint">Tab to focus, Enter to press; q quits</div>
</div>
```

### The script

The `<script>` block is Go. `count := sumi.New(0)` creates a *signal* — a
reactive container for a value. Reading it in the template (`{count}`) subscribes
that part of the UI to it; calling `count.Set(...)` or `count.Update(...)` marks
the UI dirty and schedules a re-render.

Functions in the script are event handlers. `increment` takes a `*sumi.DOMEvent`
because it is wired to `onclick`; `handleKey` takes a `sumi.Event` (a raw input
event) because it is wired to the `onkey` attribute. `sumi.Quit()` exits the app;
`sumi.EventSignal` is the terminal telling the app to stop (Ctrl+C at the OS
level, window close).

### The style

The `<style>` block is scoped CSS with a user-agent stylesheet underneath, so
HTML elements already have sensible terminal defaults. Selectors, the cascade,
and the supported property set are covered in [Selectors](selectors.md); the
`:focus` pseudo-class here styles whichever control currently holds focus, and
`opacity: dim` maps to the terminal dim attribute.

### The template

The template is HTML. `{count}` interpolates a signal's value into text.
`onclick={increment}` binds the click handler (curly braces pass the function
itself); `onkey="handleKey"` binds a raw key handler. `<button>` is focusable and
activates on Enter; Tab cycles focus between controls. The element vocabulary is
documented in [Elements](elements.md), and how boxes are sized and placed in
[Layout](layout.md).

Deviation: the primitive `<box>` and `<text>` tags were removed — use `<div>`
(or any container element) and `<span>` (or any text-level element). Self-closing
tags are written `<input />`, but a container you want to keep as a box must use
an explicit close tag: `<div></div>`, not `<div />` for content.

## Running

```
cd myapp
go run .
```

`go run .` builds and runs the generated Go. Press `q` or Ctrl+C to quit.

## The dev loop

`sumi dev` runs a supervisor that rebuilds and relaunches your app whenever a
`.sumi` or `.go` file under the directory changes:

```
sumi dev
```

It shows a status bar and mirrors your app inside it. The watcher polls the tree
every 300ms. On save it regenerates, rebuilds, and swaps in the new binary. If a
build fails, the bar shows the first line of the compiler error and the message
"(last good build still running)" — the previous working version keeps running,
so a syntax error never drops you out of the app. When your app exits, `sumi dev`
exits with it.

`sumi dev` also opens an inspect socket at `<appdir>/.sumi-dev.sock`, which the
next command uses.

## Inspecting the tree

While `sumi dev` is running, from another shell in the same directory:

```
sumi inspect tree
```

This connects to the dev socket and prints the live element tree — tag, id,
classes, text content, resolved style summary, and focus/hidden flags:

```
div
  h1 "Hello, sumi" {color:cyan}
  p
    #text "You have pressed the button "
    strong "0" {bold}
    #text " times."
  button :focus
    #text "Press me"
  div.hint
    #text "Tab to focus, Enter to press; q quits"
```

Container elements (`div`, `button`, and the other box-form tags) hold their text
in an implicit untagged `#text` child, which is why "Press me" appears under
`button` rather than beside it. The style summary reports colour, bold, italic,
and underline; other attributes such as `dim` are applied but not summarised.

`sumi inspect boxes` adds the laid-out geometry to each node (`@x,y WxH`), which
is useful when a box is not where you expect. Pass `--json` for the raw tree, or
`--dir <path>` / `--socket <path>` to target a specific dev session.
