package button

import (
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

// Scenario returns the test-preview scenario for the Button component.
func Scenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:         "button-basics",
		Width:        30,
		Height:       3,
		NewApp:       func(w, h int) *tui.App { return CreateApp(w, h) },
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "focus", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
			}},
			{Name: "click", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.EnterEvent())
			}},
		},
	}
}
