package dev

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/tui"
)

// Bridge state between the dev.sumi supervisor UI and the loop.
var (
	devStatus   *signal.Signal[string]
	devBarClass *signal.Signal[string]
	devApp      *tui.App
	devChild    *Child
	devRegion   *layout.Input
	devRows     = 23
	devCols     = 80
)

// devBind receives the supervisor UI's signals (called from dev.sumi).
func devBind(status, barClass *signal.Signal[string]) {
	devStatus = status
	devBarClass = barClass
}

// devForward relays every input event to the app under development.
func devForward(evt input.Event) {
	if devChild != nil {
		devChild.SendEvent(evt)
	}
}

// devRegionResize records the child viewport size and re-feeds the
// mirror screen into the region (called from dev.sumi on resize).
func devRegionResize(evt *layout.DOMEvent) {
	devRegion = evt.Target
	if w, ok := evt.Data["width"].(int); ok && w > 0 {
		devCols = w
	}
	if h, ok := evt.Data["height"].(int); ok && h > 0 {
		devRows = h
	}
	if devChild != nil {
		devChild.Resize(devRows, devCols)
		devRegion.Cells = devChild.Screen().Buffer()
	}
}

// setStatus updates the bar (must run on the event-loop goroutine).
func setStatus(text, class string) {
	devStatus.Set(text)
	devBarClass.Set("bar " + class)
}

// attachChild points the region at a freshly started child.
func attachChild(c *Child) {
	devChild = c
	if devRegion != nil {
		devRegion.Cells = c.Screen().Buffer()
	}
}

// RunDev runs the sumi dev supervisor for the app in dir: initial
// build+launch, watch → rebuild → swap keeping the last good child on
// errors, exit when the child exits.
func RunDev(dir string, generate func(string) error) error {
	binary := filepath.Join(dir, ".sumi-dev-bin")
	socket := DevSocketPath(dir)
	comp := NewDev(DevProps{})

	spawn := func() error {
		var c *Child
		child, err := StartChild(binary, socket, devRows, devCols,
			func() { devApp.Wake() },
			func(code int) {
				// Only the CURRENT child's exit ends the session —
				// children killed during a swap must not quit dev.
				devApp.Do(func() {
					if devChild == c {
						devApp.Quit()
					}
				})
			})
		if err != nil {
			return err
		}
		c = child
		attachChild(c)
		return nil
	}

	rebuild := func() {
		setStatus("rebuilding…", "ok")
		go func() {
			res := Build(dir, binary, generate)
			devApp.Do(func() {
				if res.Err != "" {
					msg := strings.ReplaceAll(firstLine(res.Err), dir+string(filepath.Separator), "")
					setStatus(msg+"  (last good build still running)", "err")
					return
				}
				old := devChild
				devChild = nil // silence the old child's exit callback target
				if old != nil {
					old.Stop()
				}
				if err := spawn(); err != nil {
					setStatus("launch failed: "+err.Error(), "err")
					return
				}
				setStatus(fmt.Sprintf("rebuilt in %s — %s", res.Duration.Round(time.Millisecond), dir), "ok")
			})
		}()
	}

	res := Build(dir, binary, generate)
	if res.Err != "" {
		return fmt.Errorf("initial build failed:\n%s", res.Err)
	}

	watcher := WatchTree([]string{dir}, []string{".sumi", ".go"}, 300*time.Millisecond, func() {
		devApp.Do(rebuild)
	})
	defer watcher.Stop()

	tui.RunWithOptions(comp, tui.RunOptions{
		SetApp: func(a *tui.App) {
			devApp = a
			a.Selection = nil // the child owns mouse gestures
		},
		OnPostRender: func() {
			if devChild == nil {
				if err := spawn(); err != nil {
					setStatus("launch failed: "+err.Error(), "err")
				} else {
					setStatus("watching "+dir+" — edit and save to reload", "ok")
				}
			}
		},
	})

	if devChild != nil {
		devChild.Stop()
	}
	return nil
}

// firstLine trims a multi-line tool error to the bar's single line.
func firstLine(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	return s
}
