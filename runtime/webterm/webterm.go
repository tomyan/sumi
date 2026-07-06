//go:build js

// Package webterm runs a sumi component in the browser against an
// xterm.js host. The page provides three globals before starting the
// wasm module:
//
//	sumiWrite(str)      — receives ANSI output (feed to term.write)
//	sumiCols / sumiRows — current terminal dimensions (numbers)
//
// and drives input by calling the global this package registers:
//
//	sumiInput(str)      — raw key bytes (wire to term.onData)
//	sumiResize(c, r)    — viewport changes (wire to term.onResize)
package webterm

import (
	"io"
	"syscall/js"

	"github.com/tomyan/sumi/runtime/tui"
)

// Run runs the component against the page's terminal until it quits.
func Run(comp *tui.Component) {
	in := newJSInput()
	var app *tui.App
	js.Global().Set("sumiResize", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 2 {
			js.Global().Set("sumiCols", args[0])
			js.Global().Set("sumiRows", args[1])
		}
		if app != nil {
			app.NeedsFullRedraw = true
			app.Wake()
		}
		return nil
	}))
	tui.RunWithOptions(comp, tui.RunOptions{
		In:  in,
		Out: jsWriter{},
		Size: func() (int, int) {
			return jsInt("sumiCols", 80), jsInt("sumiRows", 24)
		},
		SetApp: func(a *tui.App) { app = a },
	})
}

func jsInt(name string, fallback int) int {
	v := js.Global().Get(name)
	if v.Type() != js.TypeNumber {
		return fallback
	}
	return v.Int()
}

// jsWriter forwards frames to the page.
type jsWriter struct{}

func (jsWriter) Write(p []byte) (int, error) {
	js.Global().Call("sumiWrite", string(p))
	return len(p), nil
}

// jsInput is a Reader fed by the page's sumiInput calls.
type jsInput struct {
	ch  chan byte
	buf []byte
}

func newJSInput() *jsInput {
	in := &jsInput{ch: make(chan byte, 4096)}
	js.Global().Set("sumiInput", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 1 {
			for _, b := range []byte(args[0].String()) {
				select {
				case in.ch <- b:
				default:
				}
			}
		}
		return nil
	}))
	return in
}

// Read blocks until at least one byte is available (the event reader
// goroutine parses escape sequences byte by byte).
func (in *jsInput) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	b, ok := <-in.ch
	if !ok {
		return 0, io.EOF
	}
	p[0] = b
	n := 1
	for n < len(p) {
		select {
		case b, ok := <-in.ch:
			if !ok {
				return n, nil
			}
			p[n] = b
			n++
		default:
			return n, nil
		}
	}
	return n, nil
}
