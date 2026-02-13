package textinput

import (
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func textInputScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "textinput-basics",
		Width:  40,
		Height: 5,
		NewApp: func(w, h int) *tui.App { return CreateApp(w, h) },
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
