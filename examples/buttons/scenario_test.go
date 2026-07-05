package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func buttonsScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "buttons-basics",
		Width:  44,
		Height: 10,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "enter-presses-save", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.EnterEvent())
			}},
			{Name: "tab-then-enter-presses-cancel", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
				h.Step(sumitest.EnterEvent())
			}},
			{Name: "click-presses-save-again", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.ClickEvent(2, 5)) // inside the Save button
			}},
		},
	}
}

func TestButtonsSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, buttonsScenario())
}

func TestEnterAndClickActivateButtons(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(buttonsScenario())

	// Then
	steps := []struct {
		frame int
		want  string
	}{
		{0, "Saved 0 times, cancelled 0 times"},
		{1, "Saved 1 times, cancelled 0 times"},
		{2, "Saved 1 times, cancelled 1 times"},
		{3, "Saved 2 times, cancelled 1 times"},
	}
	for _, s := range steps {
		if !strings.Contains(frames[s.frame].StyledText, s.want) {
			t.Errorf("frame %d (%s) missing %q:\n%s",
				s.frame, frames[s.frame].Name, s.want, frames[s.frame].StyledText)
		}
	}
}

func TestClickMovesFocusToClickedButton(t *testing.T) {
	// Given — after tabbing to Cancel, click Save
	frames := sumitest.RunScenario(buttonsScenario())

	// Then — the final frame shows focus (cyan border) back on Save
	last := frames[3].StyledText
	lines := strings.Split(last, "\n")
	var saveBorder string
	for _, l := range lines {
		if strings.Contains(l, "┌") {
			saveBorder = l
			break
		}
	}
	if !strings.Contains(saveBorder, "<<cyan>>") {
		t.Errorf("expected Save button border focused (cyan) after click:\n%s", last)
	}
}
