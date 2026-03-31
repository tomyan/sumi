package anim

import (
	"math"

	"github.com/tomyan/sumi/runtime/render"
)

// LerpInt linearly interpolates between two integers.
func LerpInt(from, to int, t float64) int {
	return int(math.Round(float64(from) + t*float64(to-from)))
}

// LerpStyle interpolates between two styles at parameter t in [0, 1].
// Colors are interpolated continuously. Booleans snap at t=0.5.
func LerpStyle(from, to render.Style, t float64) render.Style {
	return render.Style{
		FG:            render.LerpColor(from.FG, to.FG, t),
		BG:            render.LerpColor(from.BG, to.BG, t),
		Bold:          lerpBool(from.Bold, to.Bold, t),
		Dim:           lerpBool(from.Dim, to.Dim, t),
		Italic:        lerpBool(from.Italic, to.Italic, t),
		Underline:     lerpBool(from.Underline, to.Underline, t),
		Strikethrough: lerpBool(from.Strikethrough, to.Strikethrough, t),
		Inverse:       lerpBool(from.Inverse, to.Inverse, t),
	}
}

// lerpBool snaps a boolean value: returns from if t < 0.5, to otherwise.
func lerpBool(from, to bool, t float64) bool {
	if t < 0.5 {
		return from
	}
	return to
}
