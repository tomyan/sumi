package render

// ColorDepth is the colour capability the renderer emits for.
type ColorDepth int

const (
	DepthAuto      ColorDepth = iota // detect from the environment at startup
	DepthMono                        // no colour (attributes only)
	Depth16                          // basic ANSI palette
	Depth256                         // xterm 256-colour palette
	DepthTrueColor                   // 24-bit RGB
)

// colorDepth is the active emission depth. Truecolor by default so tests and
// embedding are unaffected until an app opts into detection.
var colorDepth = DepthTrueColor

// SetColorDepth sets the emission depth for all subsequent rendering.
func SetColorDepth(d ColorDepth) {
	if d == DepthAuto {
		return
	}
	colorDepth = d
}

// GetColorDepth reports the active emission depth.
func GetColorDepth() ColorDepth { return colorDepth }

// quantize degrades a colour to the active depth. The result is either the
// zero Color (emit nothing), a Name (16-colour code), an Index256, or RGB.
func quantize(c Color) Color {
	c = resolveScheme(c)
	switch colorDepth {
	case DepthMono:
		return Color{}
	case Depth16:
		if c.IsRGB {
			return Color{Name: nearestANSI(c.R, c.G, c.B)}
		}
		return c
	case Depth256:
		if c.IsRGB {
			return Color{Is256: true, Index256: rgbTo256(c.R, c.G, c.B)}
		}
		return c
	}
	return c
}

// ansiRGB is the canonical sRGB value of each basic ANSI palette entry used
// for nearest-colour degradation.
var ansiRGB = []struct {
	name    string
	r, g, b int
}{
	{"black", 0, 0, 0}, {"red", 205, 49, 49}, {"green", 13, 188, 121},
	{"yellow", 229, 229, 16}, {"blue", 36, 114, 200}, {"magenta", 188, 63, 188},
	{"cyan", 17, 168, 205}, {"white", 229, 229, 229},
}

func nearestANSI(r, g, b uint8) string {
	best, bestDist := "white", 1<<62
	for _, e := range ansiRGB {
		dr, dg, db := int(r)-e.r, int(g)-e.g, int(b)-e.b
		if d := dr*dr + dg*dg + db*db; d < bestDist {
			best, bestDist = e.name, d
		}
	}
	return best
}

// rgbTo256 maps RGB onto the xterm palette: the 6x6x6 colour cube (16–231)
// or the grayscale ramp (232–255), whichever is nearer.
func rgbTo256(r, g, b uint8) uint8 {
	cr, cg, cb := cubeLevel(r), cubeLevel(g), cubeLevel(b)
	cubeIdx := 16 + 36*cr + 6*cg + cb
	cubeDist := dist3(r, g, b, cubeValue(cr), cubeValue(cg), cubeValue(cb))

	grayIdx, grayVal := nearestGray(r, g, b)
	grayDist := dist3(r, g, b, grayVal, grayVal, grayVal)

	if grayDist < cubeDist {
		return grayIdx
	}
	return uint8(cubeIdx)
}

// cubeLevel maps a channel to the nearest of the six cube levels
// (0, 95, 135, 175, 215, 255).
func cubeLevel(v uint8) int {
	if v < 48 {
		return 0
	}
	if v < 115 {
		return 1
	}
	return int(v-35) / 40
}

func cubeValue(level int) uint8 {
	if level == 0 {
		return 0
	}
	return uint8(55 + level*40)
}

// nearestGray returns the closest grayscale-ramp entry (232–255, values
// 8, 18, ..., 238) for the average luminance.
func nearestGray(r, g, b uint8) (uint8, uint8) {
	avg := (int(r) + int(g) + int(b)) / 3
	level := (avg - 8 + 5) / 10
	if level < 0 {
		level = 0
	}
	if level > 23 {
		level = 23
	}
	return uint8(232 + level), uint8(8 + level*10)
}

func dist3(r, g, b, r2, g2, b2 uint8) int {
	dr, dg, db := int(r)-int(r2), int(g)-int(g2), int(b)-int(b2)
	return dr*dr + dg*dg + db*db
}
