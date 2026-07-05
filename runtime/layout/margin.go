package layout

import "strings"

// Margin holds per-side margins plus CSS `auto` flags. Auto margins on the
// cross axis centre the box (flexbox auto-margin behaviour); main-axis auto
// margins are not yet implemented.
type Margin struct {
	Top, Right, Bottom, Left                 int
	AutoTop, AutoRight, AutoBottom, AutoLeft bool
}

// ParseMargin parses a CSS margin shorthand: 1, 2, or 4 space-separated
// values, each a cell length or `auto`.
func ParseMargin(s string) Margin {
	parts := strings.Fields(strings.TrimSpace(s))
	vals := make([]int, len(parts))
	autos := make([]bool, len(parts))
	for i, p := range parts {
		if p == "auto" {
			autos[i] = true
			continue
		}
		vals[i] = ParseCellLength(p)
	}
	switch len(parts) {
	case 1:
		return Margin{
			Top: vals[0], Right: vals[0], Bottom: vals[0], Left: vals[0],
			AutoTop: autos[0], AutoRight: autos[0], AutoBottom: autos[0], AutoLeft: autos[0],
		}
	case 2:
		return Margin{
			Top: vals[0], Bottom: vals[0], Right: vals[1], Left: vals[1],
			AutoTop: autos[0], AutoBottom: autos[0], AutoRight: autos[1], AutoLeft: autos[1],
		}
	case 4:
		return Margin{
			Top: vals[0], Right: vals[1], Bottom: vals[2], Left: vals[3],
			AutoTop: autos[0], AutoRight: autos[1], AutoBottom: autos[2], AutoLeft: autos[3],
		}
	}
	return Margin{}
}

// horizontal returns the total horizontal margin (auto counts as zero).
func (m Margin) horizontal() int { return m.Left + m.Right }

// vertical returns the total vertical margin (auto counts as zero).
func (m Margin) vertical() int { return m.Top + m.Bottom }

// autoCentreX reports whether both horizontal margins are auto.
func (m Margin) autoCentreX() bool { return m.AutoLeft && m.AutoRight }

// autoCentreY reports whether both vertical margins are auto.
func (m Margin) autoCentreY() bool { return m.AutoTop && m.AutoBottom }
