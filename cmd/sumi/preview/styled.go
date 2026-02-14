package preview

import (
	"io"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// styledSegment is a run of text with a single style.
type styledSegment struct {
	text  string
	style render.Style
}

// parseStyledLine parses a line of styled markup (e.g. "hello <<dim>>world<</>>")
// into segments with resolved styles.
func parseStyledLine(line string) []styledSegment {
	var segments []styledSegment
	var current render.Style
	rest := line

	for len(rest) > 0 {
		tagStart := strings.Index(rest, "<<")
		if tagStart < 0 {
			segments = append(segments, styledSegment{text: rest, style: current})
			break
		}
		if tagStart > 0 {
			segments = append(segments, styledSegment{text: rest[:tagStart], style: current})
		}
		tagEnd := strings.Index(rest[tagStart:], ">>")
		if tagEnd < 0 {
			segments = append(segments, styledSegment{text: rest[tagStart:], style: current})
			break
		}
		tagContent := rest[tagStart+2 : tagStart+tagEnd]
		rest = rest[tagStart+tagEnd+2:]

		if tagContent == "/" {
			current = render.Style{}
		} else {
			current = parseStyleAttrs(tagContent)
		}
	}
	return segments
}

// parseStyleAttrs converts comma-separated style attrs to a render.Style.
func parseStyleAttrs(attrs string) render.Style {
	var s render.Style
	for _, attr := range strings.Split(attrs, ",") {
		attr = strings.TrimSpace(attr)
		switch {
		case attr == "bold":
			s.Bold = true
		case attr == "dim":
			s.Dim = true
		case attr == "italic":
			s.Italic = true
		case attr == "underline":
			s.Underline = true
		case attr == "inverse":
			s.Inverse = true
		case attr == "strikethrough":
			s.Strikethrough = true
		case strings.HasPrefix(attr, "bg:"):
			s.BG = render.Color{Name: strings.TrimPrefix(attr, "bg:")}
		default:
			s.FG = render.Color{Name: attr}
		}
	}
	return s
}

// writeStyledLine renders parsed styled segments at the given position.
func writeStyledLine(w io.Writer, row, col int, segments []styledSegment) {
	offset := col
	for _, seg := range segments {
		if seg.text == "" {
			continue
		}
		render.WriteAt(w, row, offset, seg.text, seg.style)
		offset += len([]rune(seg.text))
	}
}
