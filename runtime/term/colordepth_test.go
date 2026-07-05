package term

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

// A8: colour depth detection from the environment.

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDetectColorDepth(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want render.ColorDepth
	}{
		{"NO_COLOR wins", map[string]string{"NO_COLOR": "1", "COLORTERM": "truecolor"}, render.DepthMono},
		{"COLORTERM truecolor", map[string]string{"COLORTERM": "truecolor", "TERM": "xterm-256color"}, render.DepthTrueColor},
		{"COLORTERM 24bit", map[string]string{"COLORTERM": "24bit"}, render.DepthTrueColor},
		{"TERM 256color", map[string]string{"TERM": "xterm-256color"}, render.Depth256},
		{"TERM dumb", map[string]string{"TERM": "dumb"}, render.DepthMono},
		{"plain xterm", map[string]string{"TERM": "xterm"}, render.Depth16},
		{"empty env", map[string]string{}, render.Depth16},
	}
	for _, c := range cases {
		if got := detectColorDepth(envOf(c.env)); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
