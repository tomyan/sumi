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
		Width:  46,
		Height: 10,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "typed-into-first", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.KeyEvent('a'))
			}},
			{Name: "after-tab", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
			}},
			{Name: "typed-into-second", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.KeyEvent('b'))
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

	// Then — focus starts on the first field and follows Tab / Shift+Tab
	assertFrameContains(t, frames[0].StyledText, "initial", "First (focused):", "Second (blurred):")
	assertFrameContains(t, frames[2].StyledText, "after-tab", "First (blurred):", "Second (focused):")
	assertFrameContains(t, frames[4].StyledText, "after-shift-tab", "First (focused):", "Second (blurred):")
}

func TestTypingTargetsTheFocusedField(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(focusScenario())

	// Then — 'a' lands in the first field, 'b' in the second
	assertFrameContains(t, frames[1].StyledText, "typed-into-first", "First (focused): a", "Second (blurred):")
	assertFrameContains(t, frames[3].StyledText, "typed-into-second", "First (blurred): a", "Second (focused): b")
}

func TestConsumedKeysDoNotReachRootHandler(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(focusScenario())

	// Then — both typed keys were consumed by field handlers
	last := frames[len(frames)-1].StyledText
	if !strings.Contains(last, "Root saw 0 unconsumed keys") {
		t.Errorf("expected root handler to see 0 keys:\n%s", last)
	}
}

func assertFrameContains(t *testing.T, frame, step string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(frame, want) {
			t.Errorf("%s frame missing %q:\n%s", step, want, frame)
		}
	}
}
