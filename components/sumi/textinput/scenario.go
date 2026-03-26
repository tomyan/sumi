package textinput

import (
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

// Scenario returns the test-preview scenario for the TextInput component.
func Scenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "textinput-basics",
		Width:  40,
		Height: 5,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile: "../text-input.sumi",
		Steps: []sumitest.Step{
			{Name: "initial"},
		},
	}
}
