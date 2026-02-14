package textinput

import (
	"testing"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func textInputScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:       "textinput-basics",
		Width:      40,
		Height:     5,
		NewApp:     func(w, h int) *tui.App { return CreateApp(w, h) },
		SourceFile: "../text-input.sumi",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "focus", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
			}},
			{Name: "type-hello", Action: func(h *sumitest.Harness) {
				for _, ch := range "Hello" {
					h.Step(sumitest.KeyEvent(ch))
				}
			}},
			{Name: "type-world", Action: func(h *sumitest.Harness) {
				for _, ch := range " World" {
					h.Step(sumitest.KeyEvent(ch))
				}
			}},
			{Name: "backspace", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.BackspaceEvent())
			}},
			{Name: "overflow-right", Action: func(h *sumitest.Harness) {
				// Type enough to overflow: "Hello Worl" (10) + " is a long sentence!" (20) = 30 chars
				for _, ch := range " is a long sentence!" {
					h.Step(sumitest.KeyEvent(ch))
				}
			}},
			{Name: "home", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.SpecialEvent(input.KeyHome))
			}},
			{Name: "end", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.SpecialEvent(input.KeyEnd))
			}},
		},
	}
}

func TestTextInputSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, textInputScenario())
}

func TestTextInputPreview(t *testing.T) {
	if !sumitest.PreviewMode() {
		t.Skip("run with -preview to see interactive preview")
	}
	sumitest.Preview(textInputScenario())
}
