package tui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// systemClipboard writes text to the clipboard two ways, best-effort:
// OSC 52 in-band (survives ssh and OSC-forwarding multiplexers) plus
// the platform clipboard tool when one exists. Failures are silent.
func (a *App) systemClipboard(text string) {
	a.writeOSC52(text)
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
