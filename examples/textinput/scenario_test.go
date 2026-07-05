package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func textInputScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "textinput-element",
		Width:  40,
		Height: 7,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "typed-name", Action: func(h *sumitest.Harness) {
				for _, r := range "Ada" {
					h.Step(sumitest.KeyEvent(r))
				}
			}},
			{Name: "edited-in-middle", Action: func(h *sumitest.Harness) {
				// Home, then fix the casing: "Ada" → "ada" wouldn't read well,
				// so insert a prefix instead: "Ada" → "Dr Ada".
				h.Step(sumitest.SpecialEvent("home"))
				h.Step(sumitest.PasteEvent("Dr "))
			}},
		},
	}
}

func TestTextInputSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, textInputScenario())
}

func TestTypingFlowsThroughInputEvents(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(textInputScenario())

	// Then — the oninput handler drives the greeting signal
	if !strings.Contains(frames[0].StyledText, "Hello, !") {
		t.Errorf("initial frame:\n%s", frames[0].StyledText)
	}
	if !strings.Contains(frames[1].StyledText, "Hello, Ada!") {
		t.Errorf("typed frame should greet Ada:\n%s", frames[1].StyledText)
	}
	if !strings.Contains(frames[2].StyledText, "Hello, Dr Ada!") {
		t.Errorf("home+paste frame should greet Dr Ada:\n%s", frames[2].StyledText)
	}
}
