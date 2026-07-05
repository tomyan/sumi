package render

// ColorScheme is the terminal's light/dark preference.
type ColorScheme int

const (
	SchemeDark ColorScheme = iota // terminals are dark by default
	SchemeLight
)

var colorScheme = SchemeDark

// SetColorScheme sets the active scheme used to resolve light-dark() pairs.
func SetColorScheme(s ColorScheme) { colorScheme = s }

// GetColorScheme reports the active scheme.
func GetColorScheme() ColorScheme { return colorScheme }

// ColorPair holds the two arms of a CSS light-dark() value.
type ColorPair struct {
	Light, Dark Color
}

// resolveScheme collapses a light-dark pair to the active scheme's colour.
// Colours without a pair pass through unchanged.
func resolveScheme(c Color) Color {
	if c.Pair == nil {
		return c
	}
	if colorScheme == SchemeLight {
		return c.Pair.Light
	}
	return c.Pair.Dark
}
