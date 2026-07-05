package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func dialogScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "dialog-basics",
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
			{Name: "opened", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.EnterEvent())
			}},
			{Name: "chose-no", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
				h.Step(sumitest.EnterEvent())
			}},
			{Name: "reopened", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.EnterEvent())
			}},
			{Name: "escaped", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.EscapeEvent())
			}},
		},
	}
}

func TestDialogSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, dialogScenario())
}

func TestDialogFlow(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(dialogScenario())

	// Then — closed at first, opens on Enter, No keeps, Escape closes
	if strings.Contains(frames[0].StyledText, "Really delete") {
		t.Errorf("dialog visible before opening:\n%s", frames[0].StyledText)
	}
	if !strings.Contains(frames[1].StyledText, "Really delete everything?") {
		t.Errorf("dialog should be open:\n%s", frames[1].StyledText)
	}
	if strings.Contains(frames[2].StyledText, "Really delete") ||
		!strings.Contains(frames[2].StyledText, "Status: kept") {
		t.Errorf("after choosing No:\n%s", frames[2].StyledText)
	}
	if !strings.Contains(frames[3].StyledText, "Really delete everything?") {
		t.Errorf("dialog should reopen:\n%s", frames[3].StyledText)
	}
	if strings.Contains(frames[4].StyledText, "Really delete") ||
		!strings.Contains(frames[4].StyledText, "Status: kept") {
		t.Errorf("after Escape:\n%s", frames[4].StyledText)
	}
}
