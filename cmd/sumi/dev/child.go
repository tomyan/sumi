package dev

import (
	"os"
	"os/exec"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/pty"
	"github.com/tomyan/sumi/runtime/vt100"
)

// Child runs the app under development on a PTY, mirroring its output
// into a vt100 screen the supervisor displays.
type Child struct {
	cmd    *exec.Cmd
	master *os.File
	screen *vt100.Screen
}

// StartChild launches binary on a rows×cols PTY. wake is called after
// each output chunk lands in the screen; onExit fires once when the
// process ends (with its exit code). socket, when non-empty, points the
// child's inspect listener there (sumi inspect attaches to it).
func StartChild(binary, socket string, rows, cols int, wake func(), onExit func(code int)) (*Child, error) {
	cmd := exec.Command(binary)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	if socket != "" {
		cmd.Env = append(cmd.Env, "SUMI_CONTROL_SOCKET="+socket)
	}
	master, err := pty.Start(cmd, rows, cols)
	if err != nil {
		return nil, err
	}
	c := &Child{cmd: cmd, master: master, screen: vt100.NewScreen(cols, rows)}
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				c.screen.Write(buf[:n])
				if wake != nil {
					wake()
				}
			}
			if err != nil {
				break
			}
		}
		err := cmd.Wait()
		if onExit != nil {
			code := 0
			if exit, ok := err.(*exec.ExitError); ok {
				code = exit.ExitCode()
			}
			onExit(code)
		}
	}()
	return c, nil
}

// Screen returns the live terminal model (mutated by the reader
// goroutine; renders read the current frame).
func (c *Child) Screen() *vt100.Screen { return c.screen }

// SendEvent forwards an input event to the child as terminal bytes.
func (c *Child) SendEvent(evt input.Event) {
	if b := input.EncodeEvent(evt); len(b) > 0 {
		c.master.Write(b)
	}
}

// Resize adjusts the PTY and the mirror screen.
func (c *Child) Resize(rows, cols int) {
	pty.SetSize(c.master, rows, cols)
	c.screen.Resize(cols, rows)
}

// Stop terminates the child and releases the PTY.
func (c *Child) Stop() {
	c.cmd.Process.Kill()
	c.master.Close()
}
