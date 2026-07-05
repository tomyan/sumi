package term

import (
	"os"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// DetectColorDepth infers the terminal's colour capability from the
// environment: NO_COLOR wins, then COLORTERM (truecolor/24bit), then TERM
// hints (dumb → mono, *256color* → 256), defaulting to the basic palette.
func DetectColorDepth() render.ColorDepth {
	return detectColorDepth(os.Getenv)
}

func detectColorDepth(getenv func(string) string) render.ColorDepth {
	if getenv("NO_COLOR") != "" {
		return render.DepthMono
	}
	colorterm := strings.ToLower(getenv("COLORTERM"))
	if colorterm == "truecolor" || colorterm == "24bit" {
		return render.DepthTrueColor
	}
	term := strings.ToLower(getenv("TERM"))
	switch {
	case term == "dumb":
		return render.DepthMono
	case strings.Contains(term, "256color"):
		return render.Depth256
	}
	return render.Depth16
}
