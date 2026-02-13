package sumitest

import (
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// Parse converts styled markup text back into cell rows.
// <<attrs>> opens a style context, <</>> closes it. Each line becomes
// a row of cells. This is the inverse of Buffer.ToStyledText().
func Parse(markup string) [][]render.Cell {
	if markup == "" {
		return nil
	}
	lines := strings.Split(markup, "\n")
	rows := make([][]render.Cell, len(lines))
	for i, line := range lines {
		rows[i] = parseLine(line)
	}
	return rows
}

func parseLine(line string) []render.Cell {
	var cells []render.Cell
	var current render.Style
	runes := []rune(line)
	i := 0

	for i < len(runes) {
		if i+1 < len(runes) && runes[i] == '<' && runes[i+1] == '<' {
			end := findClosingBrackets(runes, i+2)
			if end >= 0 {
				inner := string(runes[i+2 : end])
				if inner == "/" {
					current = render.Style{}
				} else {
					current = parseAttrs(inner)
				}
				i = end + 2 // skip past >>
				continue
			}
		}
		cells = append(cells, render.Cell{Ch: runes[i], Style: current})
		i++
	}
	return cells
}

func findClosingBrackets(runes []rune, start int) int {
	for i := start; i+1 < len(runes); i++ {
		if runes[i] == '>' && runes[i+1] == '>' {
			return i
		}
	}
	return -1
}

var colorNames = map[string]bool{
	"red": true, "green": true, "cyan": true, "yellow": true,
	"blue": true, "magenta": true, "white": true, "black": true,
}

func parseAttrs(s string) render.Style {
	var style render.Style
	for _, attr := range strings.Split(s, ",") {
		attr = strings.TrimSpace(attr)
		if strings.HasPrefix(attr, "bg:") {
			style.BG = render.Color{Name: strings.TrimPrefix(attr, "bg:")}
		} else if colorNames[attr] {
			style.FG = render.Color{Name: attr}
		} else {
			switch attr {
			case "bold":
				style.Bold = true
			case "dim":
				style.Dim = true
			case "italic":
				style.Italic = true
			case "underline":
				style.Underline = true
			case "strikethrough":
				style.Strikethrough = true
			case "inverse":
				style.Inverse = true
			}
		}
	}
	return style
}
