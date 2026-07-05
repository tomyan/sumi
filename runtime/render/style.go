package render

// Color represents a terminal color — either a named ANSI color or 24-bit RGB.
type Color struct {
	Name     string // "red", "green", "cyan", "yellow", "blue", "magenta", "white", "black", "" (default)
	IsRGB    bool
	R, G, B  uint8
	Is256    bool  // colour is an xterm-256 palette index (set by depth quantization)
	Index256 uint8 // palette index when Is256
}

// Style represents visual attributes for a terminal cell.
type Style struct {
	FG            Color
	BG            Color
	Bold          bool
	Dim           bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Inverse       bool
}

// IsZero returns true if the style has no attributes set.
func (s Style) IsZero() bool {
	return s == Style{}
}

// Inherit returns a style that merges inheritable properties from parent
// into child. Only properties where the child has no value set are inherited.
// FG colour inherits when child has no FG. Boolean properties (bold, dim, etc.)
// inherit only when the child has a completely zero style (no properties set).
// BG and Inverse do NOT inherit (matching CSS behaviour).
func (child Style) Inherit(parent Style) Style {
	s := child
	// FG color inherits if child has none.
	if s.FG == (Color{}) {
		s.FG = parent.FG
	}
	// Boolean properties inherit only if the child has no style at all.
	// This prevents parent's dim/bold from overriding a child that sets
	// its own colour or other properties.
	if child.IsZero() {
		s.Bold = parent.Bold
		s.Dim = parent.Dim
		s.Italic = parent.Italic
		s.Underline = parent.Underline
		s.Strikethrough = parent.Strikethrough
	}
	return s
}

var fgCodes = map[string]int{
	"black":   30,
	"red":     31,
	"green":   32,
	"yellow":  33,
	"blue":    34,
	"magenta": 35,
	"cyan":    36,
	"white":   37,
}

var bgCodes = map[string]int{
	"black":   40,
	"red":     41,
	"green":   42,
	"yellow":  43,
	"blue":    44,
	"magenta": 45,
	"cyan":    46,
	"white":   47,
}

// colorToFGCode maps a color name to its ANSI foreground code.
func colorToFGCode(name string) (int, bool) {
	code, ok := fgCodes[name]
	return code, ok
}

// colorToBGCode maps a color name to its ANSI background code.
func colorToBGCode(name string) (int, bool) {
	code, ok := bgCodes[name]
	return code, ok
}
