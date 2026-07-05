package css

import (
	"strconv"
	"strings"

	"github.com/tomyan/sumi/runtime/render"
)

// viewport is the terminal size media queries evaluate against, set by the
// runtime before each resolution pass. Zero means unknown: size conditions
// fail closed.
var viewportW, viewportH int

// SetViewport sets the viewport for media evaluation (cells).
func SetViewport(w, h int) { viewportW, viewportH = w, h }

// mediaMatches evaluates a media query against the current context:
// display-mode is terminal, prefers-color-scheme follows the active render
// scheme, and min/max-width/height compare against the viewport in cells.
// Unknown conditions fail the block (the stylesheet stays valid CSS).
func mediaMatches(query string) bool {
	for _, cond := range strings.Split(query, " and ") {
		if !conditionMatches(strings.TrimSpace(cond)) {
			return false
		}
	}
	return true
}

func conditionMatches(cond string) bool {
	cond = strings.TrimPrefix(cond, "(")
	cond = strings.TrimSuffix(cond, ")")
	name, value, found := strings.Cut(cond, ":")
	if !found {
		return false
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	switch name {
	case "display-mode":
		return value == "terminal"
	case "prefers-color-scheme":
		if render.GetColorScheme() == render.SchemeLight {
			return value == "light"
		}
		return value == "dark"
	case "min-width":
		return viewportW > 0 && viewportW >= cellLength(value)
	case "max-width":
		return viewportW > 0 && viewportW <= cellLength(value)
	case "min-height":
		return viewportH > 0 && viewportH >= cellLength(value)
	case "max-height":
		return viewportH > 0 && viewportH <= cellLength(value)
	}
	return false
}

// cellLength parses a media length (bare cells, cell, or ch units).
func cellLength(v string) int {
	v = strings.TrimSuffix(strings.TrimSuffix(v, "cell"), "ch")
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return -1
	}
	return n
}
