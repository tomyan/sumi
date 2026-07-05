package render

import (
	"fmt"
	"strings"
)

// buildSGR returns the ANSI SGR escape sequence for a style.
// Returns an empty string if the style has no attributes set.
// Used by the direct-write path (ApplyChanges).
func buildSGR(s Style) string {
	if s.IsZero() {
		return ""
	}
	var b strings.Builder
	b.WriteString("\x1b[0m")
	appendAttrCodes(&b, s)
	appendColorCodes(&b, s)
	return b.String()
}

// appendAttrCodes appends SGR codes for text attributes (bold, dim, etc.).
func appendAttrCodes(b *strings.Builder, s Style) {
	attrs := []struct {
		set  bool
		code int
	}{
		{s.Bold, 1},
		{s.Dim, 2},
		{s.Italic, 3},
		{s.Underline, 4},
		{s.Inverse, 7},
		{s.Strikethrough, 9},
	}
	for _, a := range attrs {
		if a.set {
			fmt.Fprintf(b, "\x1b[%dm", a.code)
		}
	}
}

// appendColorCodes appends SGR codes for foreground and background colors,
// degraded to the active colour depth.
func appendColorCodes(b *strings.Builder, s Style) {
	fg, bg := quantize(s.FG), quantize(s.BG)
	switch {
	case fg.IsRGB:
		fmt.Fprintf(b, "\x1b[38;2;%d;%d;%dm", fg.R, fg.G, fg.B)
	case fg.Is256:
		fmt.Fprintf(b, "\x1b[38;5;%dm", fg.Index256)
	default:
		if code, ok := colorToFGCode(fg.Name); ok {
			fmt.Fprintf(b, "\x1b[%dm", code)
		}
	}
	switch {
	case bg.IsRGB:
		fmt.Fprintf(b, "\x1b[48;2;%d;%d;%dm", bg.R, bg.G, bg.B)
	case bg.Is256:
		fmt.Fprintf(b, "\x1b[48;5;%dm", bg.Index256)
	default:
		if code, ok := colorToBGCode(bg.Name); ok {
			fmt.Fprintf(b, "\x1b[%dm", code)
		}
	}
}
