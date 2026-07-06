package tui

import (
	"encoding/json"
	"net"
	"os"

	"github.com/tomyan/sumi/runtime/layout"
)

// Inspect protocol: newline-delimited JSON over a Unix socket. sumi dev
// points SUMI_CONTROL_SOCKET at <appdir>/.sumi-dev.sock so
// `sumi inspect` can examine the running app.

// InspectNode is one element in a tree/boxes dump.
type InspectNode struct {
	Tag      string            `json:"tag,omitempty"`
	ID       string            `json:"id,omitempty"`
	Classes  []string          `json:"classes,omitempty"`
	Kind     string            `json:"kind"`
	Content  string            `json:"content,omitempty"`
	Display  string            `json:"display,omitempty"`
	Hidden   bool              `json:"hidden,omitempty"`
	Focused  bool              `json:"focused,omitempty"`
	Style    string            `json:"style,omitempty"`
	Box      *InspectBox       `json:"box,omitempty"`
	Children []*InspectNode    `json:"children,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
}

// InspectBox is laid-out geometry.
type InspectBox struct {
	X, Y, W, H int
	Fragments  []layout.Fragment `json:"fragments,omitempty"`
}

type inspectRequest struct {
	Cmd string `json:"cmd"`
}

type inspectResponse struct {
	Tree  *InspectNode `json:"tree,omitempty"`
	Error string       `json:"error,omitempty"`
}

// ServeInspect listens on socketPath and answers inspect requests from
// the running app. Snapshots are taken on the event-loop goroutine via
// app.Do, so they never race renders.
func ServeInspect(app *App, comp *Component, socketPath string) error {
	os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleInspect(app, comp, conn)
		}
	}()
	return nil
}

func handleInspect(app *App, comp *Component, conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	for {
		var req inspectRequest
		if err := dec.Decode(&req); err != nil {
			return
		}
		result := make(chan *InspectNode, 1)
		app.Do(func() {
			result <- snapshotNode(comp.Tree, boxFor(comp.LayoutResult), req.Cmd == "boxes")
		})
		app.Wake()
		tree := <-result
		if err := enc.Encode(inspectResponse{Tree: tree}); err != nil {
			return
		}
	}
}

func boxFor(b *layout.Box) *layout.Box { return b }

// snapshotNode serializes an Input (+ its Box when withBoxes) into an
// InspectNode, recursively.
func snapshotNode(n *layout.Input, box *layout.Box, withBoxes bool) *InspectNode {
	if n == nil {
		return nil
	}
	kind := "box"
	if n.Kind == layout.KindText {
		kind = "text"
	}
	node := &InspectNode{
		Tag: n.Tag, ID: n.ID, Classes: n.Classes, Kind: kind,
		Content: n.Content, Display: n.Display,
		Hidden: n.Hidden, Focused: n.Focused,
		Style: styleSummary(n), Attrs: n.Attrs,
	}
	if withBoxes && box != nil {
		node.Box = &InspectBox{X: box.X, Y: box.Y, W: box.Width, H: box.Height, Fragments: box.Fragments}
	}
	for i, c := range n.Children {
		var cb *layout.Box
		if box != nil && i < len(box.Children) {
			cb = box.Children[i]
		}
		if child := snapshotNode(c, cb, withBoxes); child != nil {
			node.Children = append(node.Children, child)
		}
	}
	return node
}

// styleSummary renders the resolved style compactly ("" when default).
func styleSummary(n *layout.Input) string {
	s := n.Style
	out := ""
	if s.FG.Name != "" {
		out += "color:" + s.FG.Name + " "
	}
	if s.BG.Name != "" {
		out += "background:" + s.BG.Name + " "
	}
	if s.Bold {
		out += "bold "
	}
	if s.Italic {
		out += "italic "
	}
	if s.Underline {
		out += "underline "
	}
	if len(out) > 0 {
		return out[:len(out)-1]
	}
	return ""
}
