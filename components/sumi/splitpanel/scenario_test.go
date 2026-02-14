package splitpanel

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/sumitest"
	"github.com/tomyan/sumi/runtime/tui"
)

func splitPanelScenario() sumitest.Scenario {
	return sumitest.Scenario{
		Name:       "splitpanel-basics",
		Width:      40,
		Height:     10,
		NewApp:     func(w, h int) *tui.App { return CreateApp(w, h) },
		SourceFile: "../split-panel.sumi",
		Steps: []sumitest.Step{
			{Name: "initial"},
		},
	}
}

func TestSplitPanelSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, splitPanelScenario())
}

func TestSplitPanelRendersWithTitles(t *testing.T) {
	// Given
	app := CreateApp(40, 10)

	// Then — panels should have titled borders
	row0 := rowText(app.TestBuffer, 0)
	if !containsStr(row0, "Actual") {
		t.Errorf("expected 'Actual' in top border, got: %q", row0)
	}
	if !containsStr(row0, "Expected") {
		t.Errorf("expected 'Expected' in top border, got: %q", row0)
	}
}

func TestSplitPanelEqualWidth(t *testing.T) {
	// Given
	app := CreateApp(40, 10)

	// Then — both panels should roughly split the width
	// The junction character should be near column 20 (middle)
	junctionIdx := -1
	for col := 0; col < 40; col++ {
		ch := app.TestBuffer.Cell(0, col).Ch
		if ch == '┬' || ch == '┐' {
			if col > 10 && col < 30 {
				junctionIdx = col
				break
			}
		}
	}
	if junctionIdx < 0 {
		t.Errorf("expected junction near middle of row 0, got: %q", rowText(app.TestBuffer, 0))
	}
}

func TestSplitPanelFixedHeight(t *testing.T) {
	// Given — panelHeight=5 in app.sumi
	app := CreateApp(40, 10)

	// Then — panels should be 5 rows tall
	// Row 4 should be the bottom border
	hasBottomBorder := false
	for col := 0; col < 40; col++ {
		ch := app.TestBuffer.Cell(4, col).Ch
		if ch == '└' || ch == '┘' || ch == '┴' || ch == '─' {
			hasBottomBorder = true
			break
		}
	}
	if !hasBottomBorder {
		t.Errorf("expected bottom border at row 4 for 5-row panels, got: %q", rowText(app.TestBuffer, 4))
	}
}

// rowText extracts the text content of a buffer row.
func rowText(buf *render.Buffer, row int) string {
	var s []rune
	for col := 0; col < buf.Width(); col++ {
		ch := buf.Cell(row, col).Ch
		if ch == 0 {
			ch = ' '
		}
		s = append(s, ch)
	}
	return string(s)
}

// containsStr checks if a string contains a substring.
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
