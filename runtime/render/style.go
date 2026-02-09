package render

// Color represents a terminal color by ANSI name.
type Color struct {
	Name string // "red", "green", "cyan", "yellow", "blue", "magenta", "white", "black", "" (default)
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
