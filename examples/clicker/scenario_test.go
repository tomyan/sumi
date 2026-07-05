package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func clickerScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "clicker-basics",
		Width:  30,
		Height: 6,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "after-click", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.ClickEvent(1, 3)) // inside the button border
			}},
			{Name: "after-miss", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.ClickEvent(5, 28)) // empty area below/right
			}},
		},
	}
}

func TestClickerSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, clickerScenario())
}

func TestClickIncrementsOnlyInsideButton(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(clickerScenario())

	// Then — the click inside the button increments; the miss does not
	if !strings.Contains(frames[0].StyledText, "Count: 0") {
		t.Errorf("initial frame missing Count: 0:\n%s", frames[0].StyledText)
	}
	if !strings.Contains(frames[1].StyledText, "Count: 1") {
		t.Errorf("after-click frame missing Count: 1:\n%s", frames[1].StyledText)
	}
	if !strings.Contains(frames[2].StyledText, "Count: 1") {
		t.Errorf("after-miss frame should still show Count: 1:\n%s", frames[2].StyledText)
	}
}
