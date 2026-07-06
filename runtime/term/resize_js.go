//go:build js

package term

// WatchResize on js/wasm: the browser host drives sizing through the
// injected streams; there is no SIGWINCH. The channel never fires.
func WatchResize() (resize <-chan struct{}, stop func()) {
	ch := make(chan struct{})
	return ch, func() {}
}
