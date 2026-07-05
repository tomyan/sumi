package tui

import (
	"os/exec"
	"runtime"
)

// OpenURL opens a URL or file with the platform opener. The a element's
// activation default action calls it; tests and embedders may replace it.
var OpenURL = func(href string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	return exec.Command(opener, href).Start()
}
