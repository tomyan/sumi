package tui

import (
	"strings"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
)

const (
	closedMarker = "▶ "
	openMarker   = "▼ "
)

// boolAttr reads an HTML boolean attribute (present and not "false").
func boolAttr(n *layout.Input, name string) bool {
	v, ok := n.Attrs[name]
	return ok && v != "false"
}

// syncDetailsElement marks the summary with a disclosure triangle and
// hides non-summary children while closed. Idempotent: the marker is
// stripped before being re-applied, so re-projection and content sync
// from signals both stay correct.
func syncDetailsElement(n *layout.Input) {
	open := boolAttr(n, "open")
	marker := closedMarker
	if open {
		marker = openMarker
	}
	for _, c := range n.Children {
		if c == nil {
			continue
		}
		if c.Tag == "summary" {
			label := strings.TrimPrefix(strings.TrimPrefix(c.Content, closedMarker), openMarker)
			c.Content = marker + label
			continue
		}
		c.Hidden = !open
	}
}

// toggleDetails flips the open attribute, re-projects, and dispatches a
// toggle event carrying {open}.
func toggleDetails(comp *Component, path []*layout.Input, details *layout.Input, evt input.Event) {
	if details.Attrs == nil {
		details.Attrs = map[string]string{}
	}
	open := !boolAttr(details, "open")
	if open {
		details.Attrs["open"] = "true"
	} else {
		delete(details.Attrs, "open")
	}
	syncDetailsElement(details)
	layout.DispatchDOM(path, &layout.DOMEvent{Type: "toggle", Key: evt,
		Data: map[string]any{"open": open}})
}
