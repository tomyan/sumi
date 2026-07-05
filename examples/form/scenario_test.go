package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func formScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "form-basics",
		Width:  40,
		Height: 8,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "toggle-notifications", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.KeyEvent(' '))
			}},
			{Name: "select-large", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
				h.Step(sumitest.TabEvent())
				h.Step(sumitest.KeyEvent(' '))
			}},
		},
	}
}

func TestFormSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, formScenario())
}

func TestCheckboxAndRadioDriveSignals(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(formScenario())

	// Then
	if !strings.Contains(frames[0].StyledText, "Notifications (off)") ||
		!strings.Contains(frames[0].StyledText, "Size: small") {
		t.Errorf("initial frame:\n%s", frames[0].StyledText)
	}
	if !strings.Contains(frames[1].StyledText, "Notifications (on)") {
		t.Errorf("after toggle:\n%s", frames[1].StyledText)
	}
	if !strings.Contains(frames[2].StyledText, "Size: large") {
		t.Errorf("after selecting large:\n%s", frames[2].StyledText)
	}
	// The radio group moved: small unchecked, large checked.
	if !strings.Contains(frames[2].StyledText, "( )") || !strings.Contains(frames[2].StyledText, "(•)") {
		t.Errorf("radio glyphs after selection:\n%s", frames[2].StyledText)
	}
}
