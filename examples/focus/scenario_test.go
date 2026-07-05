package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func focusScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "focus-basics",
		Width:  40,
		Height: 10,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "after-tab", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
			}},
			{Name: "after-shift-tab", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.ShiftTabEvent())
			}},
		},
	}
}

func TestFocusSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, focusScenario())
}

func TestTabMovesFocusBetweenFields(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(focusScenario())

	// Then — first field starts focused, Tab hands focus to the second,
	// Shift+Tab hands it back.
	assertFieldStates(t, frames[0].StyledText, "initial", "First field: focused", "Second field: blurred")
	assertFieldStates(t, frames[1].StyledText, "after-tab", "First field: blurred", "Second field: focused")
	assertFieldStates(t, frames[2].StyledText, "after-shift-tab", "First field: focused", "Second field: blurred")
}

func assertFieldStates(t *testing.T, frame, step string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(frame, want) {
			t.Errorf("%s frame missing %q:\n%s", step, want, frame)
		}
	}
}
