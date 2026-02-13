package sumitest

import (
	"strings"
	"testing"
)

func counterScenario() Scenario {
	return Scenario{
		Name:   "counter-basics",
		Width:  20,
		Height: 3,
		NewApp: createCounterApp,
		Steps: []Step{
			{Name: "initial"},
			{Name: "after-first-key", Action: func(h *Harness) {
				h.Step(KeyEvent('x'))
			}},
			{Name: "after-second-key", Action: func(h *Harness) {
				h.Step(KeyEvent('y'))
			}},
		},
	}
}

func TestRunScenarioReturnsCorrectFrameCount(t *testing.T) {
	// Given
	s := counterScenario()

	// When
	frames := RunScenario(s)

	// Then
	if len(frames) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(frames))
	}
}

func TestRunScenarioFrameNames(t *testing.T) {
	// Given
	s := counterScenario()

	// When
	frames := RunScenario(s)

	// Then
	expected := []string{"initial", "after-first-key", "after-second-key"}
	for i, name := range expected {
		if frames[i].Name != name {
			t.Errorf("frame %d: expected name %q, got %q", i, name, frames[i].Name)
		}
	}
}

func TestRunScenarioInitialFrameContent(t *testing.T) {
	// Given
	s := counterScenario()

	// When
	frames := RunScenario(s)

	// Then
	if !strings.Contains(frames[0].StyledText, "Count: 0") {
		t.Errorf("initial frame should contain 'Count: 0', got:\n%s", frames[0].StyledText)
	}
}

func TestRunScenarioFrameContentProgresses(t *testing.T) {
	// Given
	s := counterScenario()

	// When
	frames := RunScenario(s)

	// Then
	if !strings.Contains(frames[1].StyledText, "Count: 1") {
		t.Errorf("frame 1 should contain 'Count: 1', got:\n%s", frames[1].StyledText)
	}
	if !strings.Contains(frames[2].StyledText, "Count: 2") {
		t.Errorf("frame 2 should contain 'Count: 2', got:\n%s", frames[2].StyledText)
	}
}

func TestRunScenarioStyledTextIsNotEmpty(t *testing.T) {
	// Given
	s := counterScenario()

	// When
	frames := RunScenario(s)

	// Then
	for i, f := range frames {
		if f.StyledText == "" {
			t.Errorf("frame %d has empty StyledText", i)
		}
	}
}
