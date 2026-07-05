package tui

import (
	"math"
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/layout"
)

// eighthBlocks are the left-partial block glyphs, indexed by eighths 1-7.
var eighthBlocks = []string{"", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}

// syncBarElement projects a progress/meter bar into the implicit child:
// full blocks, an eighth-block partial, then track. No value renders an
// indeterminate all-track bar.
func syncBarElement(n *layout.Input) {
	width := inputContentWidth(n)
	if width <= 0 {
		return
	}
	child := ensureValueChild(n)
	frac, ok := barFraction(n)
	if !ok {
		child.Content = strings.Repeat("░", width)
		return
	}
	eighths := int(math.Round(frac * float64(width) * 8))
	full := eighths / 8
	partial := eighthBlocks[eighths%8]
	content := strings.Repeat("█", full) + partial
	track := width - full - len([]rune(partial))
	child.Content = content + strings.Repeat("░", track)
}

// barFraction computes value's position in [min, max], clamped to [0, 1].
// Defaults follow HTML: max 1, min 0.
func barFraction(n *layout.Input) (float64, bool) {
	value, err := strconv.ParseFloat(n.Attrs["value"], 64)
	if err != nil {
		return 0, false
	}
	max := 1.0
	if v, err := strconv.ParseFloat(n.Attrs["max"], 64); err == nil {
		max = v
	}
	min := 0.0
	if v, err := strconv.ParseFloat(n.Attrs["min"], 64); err == nil {
		min = v
	}
	if max <= min {
		return 0, false
	}
	frac := (value - min) / (max - min)
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	return frac, true
}
