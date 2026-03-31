package render

import "math"

// NamedColorRGB maps ANSI color names to standard RGB values.
var NamedColorRGB = map[string]Color{
	"black":   {IsRGB: true, R: 0, G: 0, B: 0},
	"red":     {IsRGB: true, R: 255, G: 0, B: 0},
	"green":   {IsRGB: true, R: 0, G: 255, B: 0},
	"yellow":  {IsRGB: true, R: 255, G: 255, B: 0},
	"blue":    {IsRGB: true, R: 0, G: 0, B: 255},
	"magenta": {IsRGB: true, R: 255, G: 0, B: 255},
	"cyan":    {IsRGB: true, R: 0, G: 255, B: 255},
	"white":   {IsRGB: true, R: 255, G: 255, B: 255},
}

// ToRGB converts a named color to its RGB equivalent. RGB colors pass through.
// Unknown names or zero colors return unchanged.
func (c Color) ToRGB() Color {
	if c.IsRGB {
		return c
	}
	if rgb, ok := NamedColorRGB[c.Name]; ok {
		return rgb
	}
	return c
}

// IsZeroColor returns true if the color is unset (no name and not RGB).
func (c Color) IsZeroColor() bool {
	return c.Name == "" && !c.IsRGB
}

// LerpColor linearly interpolates between two colors at parameter t in [0, 1].
// Both colors are resolved to RGB first. If either color is zero (unset),
// the other is returned directly.
func LerpColor(from, to Color, t float64) Color {
	if from.IsZeroColor() {
		return to
	}
	if to.IsZeroColor() {
		return from
	}
	a := from.ToRGB()
	b := to.ToRGB()
	// If ToRGB failed (unknown named color), fall back.
	if !a.IsRGB || !b.IsRGB {
		if t < 0.5 {
			return from
		}
		return to
	}
	return Color{
		IsRGB: true,
		R:     lerpByte(a.R, b.R, t),
		G:     lerpByte(a.G, b.G, t),
		B:     lerpByte(a.B, b.B, t),
	}
}

func lerpByte(a, b uint8, t float64) uint8 {
	v := float64(a) + t*(float64(b)-float64(a))
	return uint8(math.Round(v))
}
