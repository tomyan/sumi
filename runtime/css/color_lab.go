package css

import (
	"math"

	"github.com/tomyan/sumi/runtime/render"
)

// chToAB converts chroma + hue (degrees) to the a, b rectangular components.
func chToAB(c, h float64) (float64, float64) {
	rad := h * math.Pi / 180
	return c * math.Cos(rad), c * math.Sin(rad)
}

// labColor converts CIELAB (D50, as CSS lab() specifies) to sRGB.
func labColor(l, a, b float64) render.Color {
	// Lab → XYZ (D50 reference white)
	fy := (l + 16) / 116
	fx := fy + a/500
	fz := fy - b/200
	xyz := func(t float64) float64 {
		if t3 := t * t * t; t3 > 0.008856 {
			return t3
		}
		return (t - 16.0/116) / 7.787
	}
	x := xyz(fx) * 0.9642
	y := xyz(fy) * 1.0
	z := xyz(fz) * 0.8249

	// XYZ (D50) → linear sRGB (Bradford-adapted matrix)
	r := 3.1338561*x - 1.6168667*y - 0.4906146*z
	g := -0.9787684*x + 1.9161415*y + 0.0334540*z
	bb := 0.0719453*x - 0.2289914*y + 1.4052427*z
	return rgbColor(gamma(r)*255, gamma(g)*255, gamma(bb)*255)
}

// oklabColor converts OKLab to sRGB.
func oklabColor(l, a, b float64) render.Color {
	lp := l + 0.3963377774*a + 0.2158037573*b
	mp := l - 0.1055613458*a - 0.0638541728*b
	sp := l - 0.0894841775*a - 1.2914855480*b
	l3, m3, s3 := lp*lp*lp, mp*mp*mp, sp*sp*sp

	r := 4.0767416621*l3 - 3.3077115913*m3 + 0.2309699292*s3
	g := -1.2684380046*l3 + 2.6097574011*m3 - 0.3413193965*s3
	bb := -0.0041960863*l3 - 0.7034186147*m3 + 1.7076147010*s3
	return rgbColor(gamma(r)*255, gamma(g)*255, gamma(bb)*255)
}

// gamma encodes one linear-light sRGB channel.
func gamma(u float64) float64 {
	if u <= 0.0031308 {
		return 12.92 * u
	}
	return 1.055*math.Pow(u, 1/2.4) - 0.055
}

// rgbColor clamps float channels into a truecolor render.Color.
func rgbColor(r, g, b float64) render.Color {
	return render.Color{IsRGB: true, R: clamp255(r), G: clamp255(g), B: clamp255(b)}
}

func clamp255(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}
