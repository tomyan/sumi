package tui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// systemClipboard writes text to the clipboard two ways, best-effort:
// OSC 52 in-band (survives ssh and OSC-forwarding multiplexers) plus
// the platform clipboard tool when one exists. Failures are silent.
func systemClipboard(text string) {
	render.CopyToClipboard(os.Stdout, text)
	copyViaPlatformTool(text)
}

// copyViaPlatformTool pipes text to the OS clipboard utility.
func copyViaPlatformTool(text string) {
	cmd := platformClipboardCmd()
	if cmd == nil {
		return
	}
	cmd.Stdin = strings.NewReader(text)
	go func() { _ = cmd.Run() }()
}

func platformClipboardCmd() *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("pbcopy")
	case "linux":
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			return exec.Command("wl-copy")
		}
		return exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		return exec.Command("clip")
	}
	return nil
}
