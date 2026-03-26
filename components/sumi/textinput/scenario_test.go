package textinput

import (
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func textInputScenario() sumitest.Scenario {
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

func TestTextInputSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, textInputScenario())
}
