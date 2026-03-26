package splitpanel

import (
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

// Scenario returns the test-preview scenario for the SplitPanel component.
func Scenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:       "splitpanel-basics",
		Width:      40,
		Height:     10,
		NewApp:     func(w, h int) *tui.App { return CreateApp(w, h) },
		SourceFile: "../split-panel.sumi",
		Steps: []sumitest.Step{
			{Name: "initial"},
		},
	}
}
