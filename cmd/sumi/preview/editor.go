package preview

import (
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/pty"
	"github.com/tomyan/sumi/runtime/vt100"
)

// Editor wraps an nvim process on a PTY with a VT100 screen.
type Editor struct {
	master *os.File
	screen *vt100.Screen
	cmd    *exec.Cmd
	mu     sync.Mutex
	done   chan struct{}
}

// NewEditor starts nvim editing filePath on a PTY of the given size.
// The optional wake callback is called after each PTY read to trigger re-render.
func NewEditor(filePath string, rows, cols int, wake func()) (*Editor, error) {
	cmd := exec.Command("nvim",
		"--noplugin", "-u", "NONE",
		"+set updatetime=300", "+set autowriteall",
		filePath,
	)

	master, err := pty.Start(cmd, rows, cols)
	if err != nil {
		return nil, err
	}

	ed := &Editor{
		master: master,
		screen: vt100.NewScreen(cols, rows),
		cmd:    cmd,
		done:   make(chan struct{}),
	}

	go ed.readLoop(wake)
	return ed, nil
}

// Screen returns the VT100 screen. Callers must hold ed.mu for concurrent access.
func (ed *Editor) Screen() *vt100.Screen {
	return ed.screen
}

// SendEvent encodes an input event and writes it to the nvim PTY.
func (ed *Editor) SendEvent(evt input.Event) {
	ed.SendBytes(input.EncodeEvent(evt))
}

// SendBytes writes raw bytes to the nvim PTY.
func (ed *Editor) SendBytes(b []byte) {
	if len(b) > 0 {
		ed.master.Write(b)
	}
}

// Resize updates the PTY size and recreates the screen buffer.
func (ed *Editor) Resize(rows, cols int) {
	pty.SetSize(ed.master, rows, cols)

	ed.mu.Lock()
	ed.screen.Resize(cols, rows)
	ed.mu.Unlock()

	// Send SIGWINCH to let nvim know about the resize.
	if ed.cmd.Process != nil {
		ed.cmd.Process.Signal(syscall.SIGWINCH)
	}
}

// Close terminates the nvim process and cleans up.
func (ed *Editor) Close() {
	if ed.cmd.Process != nil {
		ed.cmd.Process.Signal(syscall.SIGTERM)
	}
	ed.master.Close()
	<-ed.done
}

// readLoop reads PTY output and feeds it to the VT100 screen.
func (ed *Editor) readLoop(wake func()) {
	defer close(ed.done)

	buf := make([]byte, 4096)
	for {
		n, err := ed.master.Read(buf)
		if n > 0 {
			ed.mu.Lock()
			ed.screen.Write(buf[:n])
			ed.mu.Unlock()
			if wake != nil {
				wake()
			}
		}
		if err != nil {
			return
		}
	}
}
