package css

import (
	"math"
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// ansiNames are the 8 basic colour keywords that map to themeable terminal
// palette entries rather than fixed RGB values.
var ansiNames = map[string]bool{
	"black": true, "red": true, "green": true, "yellow": true,
	"blue": true, "magenta": true, "cyan": true, "white": true,
}

// ParseColorValue parses a CSS colour value: the 8 ANSI keywords (kept as
// palette names), the CSS named colours, hex (3/4/6/8 digit), and the
// rgb()/hsl()/hwb()/lab()/lch()/oklab()/oklch() functions in legacy or
// modern syntax. `transparent` yields the zero Color (no colour set).
// Alpha components parse but are ignored (compositing lands with F2).
// Reports false for unrecognised values (graceful drop).
func ParseColorValue(v string) (render.Color, bool) {
	v = strings.TrimSpace(strings.ToLower(v))
	switch {
	case v == "":
		return render.Color{}, false
	case v == "transparent":
		return render.Color{}, true
	case ansiNames[v]:
		return render.Color{Name: v}, true
	}
	if hex, ok := namedColors[v]; ok {
		return hexColor(hex)
	}
	if strings.HasPrefix(v, "#") {
		return hexColor(v)
	}
	if open := strings.IndexByte(v, '('); open > 0 && strings.HasSuffix(v, ")") {
		return functionColor(v[:open], v[open+1:len(v)-1])
	}
	return render.Color{}, false
}

func hexColor(v string) (render.Color, bool) {
	h := strings.TrimPrefix(v, "#")
	switch len(h) {
	case 3, 4: // #rgb / #rgba
		h = expandShortHex(h[:3])
	case 6: // #rrggbb
	case 8: // #rrggbbaa — alpha ignored
		h = h[:6]
	default:
		return render.Color{}, false
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return render.Color{}, false
	}
	return rgbColor(float64(n>>16&0xff), float64(n>>8&0xff), float64(n&0xff)), true
}

func expandShortHex(h string) string {
	var b strings.Builder
	for i := 0; i < 3; i++ {
		b.WriteByte(h[i])
		b.WriteByte(h[i])
	}
	return b.String()
}

func functionColor(name, args string) (render.Color, bool) {
	nums, ok := colorArgs(args)
	if !ok || len(nums) < 3 {
		return render.Color{}, false
	}
	c := nums[:3]
	switch name {
	case "rgb", "rgba":
		return rgbColor(c[0].value(255), c[1].value(255), c[2].value(255)), true
	case "hsl", "hsla":
		r, g, b := hslToRgb(c[0].angle(), c[1].value(1), c[2].value(1))
		return rgbColor(r, g, b), true
	case "hwb":
		r, g, b := hwbToRgb(c[0].angle(), c[1].value(1), c[2].value(1))
		return rgbColor(r, g, b), true
	case "lab":
		return labColor(c[0].value(100), c[1].scaled(125), c[2].scaled(125)), true
	case "lch":
		a, b := chToAB(c[1].scaled(150), c[2].angle())
		return labColor(c[0].value(100), a, b), true
	case "oklab":
		return oklabColor(c[0].value(1), c[1].scaled(0.4), c[2].scaled(0.4)), true
	case "oklch":
		a, b := chToAB(c[1].scaled(0.4), c[2].angle())
		return oklabColor(c[0].value(1), a, b), true
	}
	return render.Color{}, false
}

// component is one numeric colour-function argument.
type component struct {
	num   float64
	isPct bool
	unit  string // angle unit for hue components
}

// value resolves the component against the scale a percentage maps onto
// (100% → scale). Plain numbers pass through unchanged.
func (c component) value(scale float64) float64 {
	if c.isPct {
		return c.num / 100 * scale
	}
	return c.num
}

// scaled resolves a ±100% = ±scale component (lab/lch a, b, chroma).
func (c component) scaled(scale float64) float64 {
	if c.isPct {
		return c.num / 100 * scale
	}
	return c.num
}

// angle resolves a hue to degrees.
func (c component) angle() float64 {
	switch c.unit {
	case "rad":
		return c.num * 180 / math.Pi
	case "grad":
		return c.num * 0.9
	case "turn":
		return c.num * 360
	}
	return c.num
}

// colorArgs tokenizes function arguments in legacy (comma) or modern (space,
// slash-alpha) syntax. The alpha component, if present, is dropped.
func colorArgs(args string) ([]component, bool) {
	args = strings.ReplaceAll(args, ",", " ")
	if i := strings.IndexByte(args, '/'); i >= 0 {
		args = args[:i]
	}
	fields := strings.Fields(args)
	out := make([]component, 0, len(fields))
	for _, f := range fields {
		c, ok := parseComponent(f)
		if !ok {
			return nil, false
		}
		out = append(out, c)
	}
	return out, true
}

func parseComponent(f string) (component, bool) {
	var c component
	if strings.HasSuffix(f, "%") {
		c.isPct = true
		f = strings.TrimSuffix(f, "%")
	}
	for _, unit := range []string{"deg", "grad", "rad", "turn"} {
		if strings.HasSuffix(f, unit) {
			c.unit = unit
			f = strings.TrimSuffix(f, unit)
			break
		}
	}
	if f == "none" {
		return component{}, true
	}
	n, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return component{}, false
	}
	c.num = n
	return c, true
}

// hsl values: s, l in 0..1 when from percentages (value(1)); numbers 0..1.
func hslToRgb(h, s, l float64) (float64, float64, float64) {
	h = math.Mod(math.Mod(h, 360)+360, 360) / 360
	if s == 0 {
		return l * 255, l * 255, l * 255
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := hueToRgb(p, q, h+1.0/3)
	g := hueToRgb(p, q, h)
	b := hueToRgb(p, q, h-1.0/3)
	return r * 255, g * 255, b * 255
}

func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6:
		return p + (q-p)*6*t
	case t < 1.0/2:
		return q
	case t < 2.0/3:
		return p + (q-p)*(2.0/3-t)*6
	}
	return p
}

func hwbToRgb(h, w, bl float64) (float64, float64, float64) {
	if w+bl >= 1 {
		g := w / (w + bl) * 255
		return g, g, g
	}
	r, g, b := hslToRgb(h, 1, 0.5)
	scale := 1 - w - bl
	r = (r/255*scale + w) * 255
	g = (g/255*scale + w) * 255
	b = (b/255*scale + w) * 255
	return r, g, b
}
