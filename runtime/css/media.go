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

// reducedMotion holds the prefers-reduced-motion preference. Terminals
// have no standard probe; RunOptions and the SUMI_REDUCED_MOTION env var
// drive it.
var reducedMotion bool

// SetReducedMotion sets the prefers-reduced-motion media preference.
func SetReducedMotion(reduced bool) { reducedMotion = reduced }

// ReducedMotion reports the current prefers-reduced-motion preference.
func ReducedMotion() bool { return reducedMotion }

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
	case "prefers-reduced-motion":
		if reducedMotion {
			return value == "reduce"
		}
		return value == "no-preference"
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

// supportedProperties are the property names sumi consumes; @supports
// property-name checks test membership. Custom properties always pass.
var supportedProperties = map[string]bool{
	"color": true, "background": true, "background-color": true,
	"border-color": true, "font-weight": true, "font-style": true,
	"text-decoration": true, "opacity": true, "inverse": true,
	"flex-direction": true, "width": true, "height": true, "gap": true,
	"flex-grow": true, "min-width": true, "justify-content": true,
	"align-items": true, "padding": true, "display": true,
	"overflow": true, "position": true, "top": true, "left": true,
	"right": true, "bottom": true, "z-index": true, "border": true,
	"border-top": true, "border-bottom": true, "border-title": true,
	"border-collapse": true, "transition": true, "animation": true,
}

// supportsMatches evaluates an @supports condition: (property: value)
// name checks joined by `and`. Unknown forms fail closed.
func supportsMatches(cond string) bool {
	for _, c := range strings.Split(cond, " and ") {
		c = strings.TrimSpace(c)
		c = strings.TrimPrefix(c, "(")
		c = strings.TrimSuffix(c, ")")
		name, _, found := strings.Cut(c, ":")
		if !found {
			return false
		}
		name = strings.TrimSpace(name)
		if !supportedProperties[name] && !strings.HasPrefix(name, "--") {
			return false
		}
	}
	return true
}

// containerMatches evaluates an @container size condition against the
// nearest laid-out ancestor's dimensions. Zero dimensions (not yet laid
// out) fail closed; a second resolve pass runs after layout.
func containerMatches(cond string, w, h int) bool {
	for _, c := range strings.Split(cond, " and ") {
		if !containerCondition(strings.TrimSpace(c), w, h) {
			return false
		}
	}
	return true
}

func containerCondition(cond string, w, h int) bool {
	cond = strings.TrimPrefix(cond, "(")
	cond = strings.TrimSuffix(cond, ")")
	name, value, found := strings.Cut(cond, ":")
	if !found {
		return false
	}
	switch strings.TrimSpace(name) {
	case "min-width":
		return w > 0 && w >= cellLength(strings.TrimSpace(value))
	case "max-width":
		return w > 0 && w <= cellLength(strings.TrimSpace(value))
	case "min-height":
		return h > 0 && h >= cellLength(strings.TrimSpace(value))
	case "max-height":
		return h > 0 && h <= cellLength(strings.TrimSpace(value))
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
