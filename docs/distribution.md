# Distribution

A sumi app ships as a single static binary. There is no runtime to install
alongside it, no `node_modules`, no interpreter — you hand someone one file and
it runs. This is the concrete payoff of building a TUI framework in Go rather
than on a scripting runtime, and it shapes how sumi apps are built, cross-
compiled, and delivered.

## One binary

`.sumi` components compile to ordinary Go source, and Go source compiles to a
single self-contained executable:

```sh
sumi generate .   # .sumi -> _sumi.go (committed alongside the source)
go build .        # _sumi.go + your Go -> one static binary
```

`sumi generate` is a source-generation step, not a hidden build stage: it emits
a sibling `foo_sumi.go` next to each `foo.sumi`, and those generated files are
checked into version control like any other source. `sumi init myapp` scaffolds
a new app (component, `main.go`, `go.mod`) and wires a `//go:generate sumi
generate .` directive, so `go generate ./...` regenerates before a build. After
generation the app is plain Go — `go build` is all that stands between source
and a runnable binary.

## Near-zero dependencies

Sumi depends on two modules outside the standard library: `golang.org/x/sys`
and `golang.org/x/term`. Both are pure Go and part of Go's own extended
standard library. There is no cgo anywhere in the tree, and no third-party
dependency beyond those two Google-maintained modules. That is the whole
dependency surface an app inherits from the framework.

Because there is no cgo, the binary is genuinely static: it has no shared-library
dependencies to satisfy on the target machine.

## Cross-compilation

A pure-Go, cgo-free codebase cross-compiles with the stock toolchain by setting
`GOOS` and `GOARCH` — no C cross-toolchain, no per-platform build environment:

```sh
GOOS=linux   GOARCH=amd64 go build -o myapp-linux    .
GOOS=darwin  GOARCH=arm64 go build -o myapp-macos-arm .
GOOS=windows GOARCH=amd64 go build -o myapp.exe       .
```

The only platform-specific code in sumi is guarded by Go build tags and uses
`syscall` and `golang.org/x/sys` rather than C, so each target selects the right
implementation at compile time. Sumi also builds for `GOOS=js GOARCH=wasm`,
running the same component in the browser against an xterm.js host — the
declarative component is written once and targets either a native TTY or the
web.

## Binary size and startup

A compiled sumi app is a handful of megabytes — a small example lands in the
low single-digit MB range — reflecting the Go runtime and the sumi packages
statically linked in. There is no interpreter warm-up or module resolution at
launch: the process starts and draws its first frame immediately.

## Degrading across terminals

One binary runs across a wide range of terminals because sumi degrades
gracefully rather than requiring specific capabilities. Colour is quantised to
the terminal's detected depth, unsupported escape sequences are either ignored
harmlessly or backed by a legacy fallback, and CSS that has no cell-grid meaning
is dropped. You do not build separate binaries for capable and limited
terminals — the same executable adapts at runtime. See
[terminal support](terminals.md) for the full capability matrix and
[CSS on a cell grid](terminal-css.md) for how styling degrades.

## Current caveat

Sumi is not yet published as a module. Until it is, `sumi init` inserts a local
`replace` directive pointing `go.mod` at your checkout, so the plain-`go build`
story currently assumes that checkout is present. Once the module is published
the `replace` directive goes away and `go build` resolves sumi like any other
dependency.
