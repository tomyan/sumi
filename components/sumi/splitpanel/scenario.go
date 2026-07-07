package splitpanel

import (
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

// createApp builds a test-mode app from the generated component constructor.
func createApp(w, h int) *tui.App {
	return tui.TestApp(NewApp(AppProps{}), w, h)
}

// Scenario returns the test-preview scenario for the SplitPanel component.
func Scenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:       "splitpanel-basics",
		Width:      40,
		Height:     10,
		NewApp:     createApp,
		SourceFile: "../split-panel.sumi",
		Steps: []sumitest.Step{
			{Name: "initial"},
		},
	}
}
