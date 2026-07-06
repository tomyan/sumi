//go:build js

package tui

// stopSelf is a no-op on js/wasm — there is no shell to suspend to.
func stopSelf() {}
