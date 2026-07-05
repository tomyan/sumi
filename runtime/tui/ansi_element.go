package tui

import (
	"strings"

	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/vt100"
)

// syncAnsiElement parses the element's raw text body — which may contain
// SGR escape sequences — into per-cell styled content. The source child
// is hidden; the parsed cells render instead. Re-parsed each projection
// pass so signal-driven content stays fresh.
func syncAnsiElement(n *layout.Input) {
	source := ansiSourceChild(n)
	if source == nil {
		return
	}
	source.Hidden = true
	width, height := ansiSize(source.Content)
	if width == 0 || height == 0 {
		n.Cells = nil
		return
	}
	screen := vt100.NewScreen(width, height)
	// The parser treats \n as a bare line feed; give it carriage returns.
	screen.Write([]byte(strings.ReplaceAll(source.Content, "\n", "\r\n")))
	n.Cells = screen.Buffer()
	if _, ok := n.Attrs["width"]; !ok {
		n.FixedWidth = width
	}
	if _, ok := n.Attrs["height"]; !ok {
		n.FixedHeight = height
	}
}

// ansiSourceChild finds the text body holding the raw sequence source.
func ansiSourceChild(n *layout.Input) *layout.Input {
	for _, c := range n.Children {
		if c != nil && c.Kind == layout.KindText {
			return c
		}
	}
	return nil
}

// ansiSize measures the visible extent of text containing escape
// sequences: the widest line in runes, and the line count.
func ansiSize(s string) (width, height int) {
	col := 0
	inEscape := false
	csi := false
	for _, r := range s {
		switch {
		case inEscape:
			if csi {
				if r >= 0x40 && r <= 0x7e {
					inEscape = false
				}
			} else if r == '[' {
				csi = true
			} else {
				inEscape = false
			}
		case r == 0x1b:
			inEscape = true
			csi = false
		case r == '\n':
			height++
			col = 0
		default:
			col++
			if col > width {
				width = col
			}
		}
	}
	if col > 0 || height > 0 {
		height++
	}
	return width, height
}
