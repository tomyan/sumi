package tui_test

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

// G3a: the inspect socket dumps the live tree and geometry.

func TestInspectTreeAndBoxes(t *testing.T) {
	// Given: a running test app serving inspect.
	comp := &tui.Component{Tree: &layout.Input{
		Kind: layout.KindBox, CursorCol: -1, CursorRow: -1,
		Children: []*layout.Input{
			{Tag: "h1", Kind: layout.KindText, Content: "Title", CursorCol: -1, CursorRow: -1,
				Style: render.Style{Bold: true, FG: render.Color{Name: "cyan"}}},
			{Tag: "div", Kind: layout.KindBox, Classes: []string{"panel"}, CursorCol: -1, CursorRow: -1},
		},
	}}
	app := tui.TestApp(comp, 30, 5)
	sock := filepath.Join(t.TempDir(), "inspect.sock")
	if err := tui.ServeInspect(app, comp, sock); err != nil {
		t.Fatal(err)
	}
	// The test app has no event loop; service Do callbacks manually.
	go func() {
		for {
			app.Step(input.Event{Kind: input.EventFrame})
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// When
	conn, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	enc, dec := json.NewEncoder(conn), json.NewDecoder(conn)
	if err := enc.Encode(map[string]string{"cmd": "boxes"}); err != nil {
		t.Fatal(err)
	}
	var resp struct {
		Tree *tui.InspectNode `json:"tree"`
	}
	if err := dec.Decode(&resp); err != nil {
		t.Fatal(err)
	}

	// Then
	if resp.Tree == nil || len(resp.Tree.Children) != 2 {
		t.Fatalf("tree = %+v", resp.Tree)
	}
	h1 := resp.Tree.Children[0]
	if h1.Tag != "h1" || h1.Content != "Title" || h1.Style == "" {
		t.Errorf("h1 = %+v, want tag/content/style", h1)
	}
	if h1.Box == nil || h1.Box.W == 0 {
		t.Errorf("h1 box = %+v, want geometry", h1.Box)
	}
	div := resp.Tree.Children[1]
	if len(div.Classes) != 1 || div.Classes[0] != "panel" {
		t.Errorf("div classes = %v", div.Classes)
	}
}
