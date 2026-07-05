package main

import (
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func linksScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:   "links-basics",
		Width:  40,
		Height: 5,
		NewApp: func(w, h int) *tui.App {
			comp := NewApp(AppProps{})
			return tui.TestApp(comp, w, h)
		},
		SourceFile:   "app.sumi",
		ScenarioFile: "scenario_test.go",
		Steps: []sumitest.Step{
			{Name: "initial"},
			{Name: "tab-to-blog", Action: func(h *sumitest.Harness) {
				h.Step(sumitest.TabEvent())
			}},
		},
	}
}

func TestLinksSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, linksScenario())
}

func TestEnterOpensTheFocusedLink(t *testing.T) {
	// Given — capture opens instead of shelling out
	var opened []string
	prev := tui.OpenURL
	tui.OpenURL = func(href string) error { opened = append(opened, href); return nil }
	t.Cleanup(func() { tui.OpenURL = prev })

	comp := NewApp(AppProps{})
	app := tui.TestApp(comp, 40, 5)

	// When — Enter on the first link, Tab, Enter on the second
	app.Step(sumitest.EnterEvent())
	app.Step(sumitest.TabEvent())
	app.Step(sumitest.EnterEvent())

	// Then
	want := []string{"https://example.com/docs", "https://example.com/blog"}
	if len(opened) != 2 || opened[0] != want[0] || opened[1] != want[1] {
		t.Errorf("opened = %v, want %v", opened, want)
	}
}

func TestFocusedLinkRendersInverse(t *testing.T) {
	// Given
	frames := sumitest.RunScenario(linksScenario())

	// Then — the a:focus rule highlights the focused link
	if !strings.Contains(frames[0].StyledText, "<<inverse>>Documentation<</>>") {
		t.Errorf("initial frame should show Documentation focused:\n%s", frames[0].StyledText)
	}
	if !strings.Contains(frames[1].StyledText, "<<inverse>>Blog<</>>") {
		t.Errorf("after Tab the Blog link should be focused:\n%s", frames[1].StyledText)
	}
}
