package layout

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tomyan/sumi/runtime/render"
)

func TestApplyChangesTextContentUpdate(t *testing.T) {
	// Given
	var buf bytes.Buffer
	changes := []Change{
		{
			Old: &Box{X: 0, Y: 0, Width: 5, Height: 1, Content: "hello"},
			New: &Box{X: 0, Y: 0, Width: 3, Height: 1, Content: "bye"},
		},
	}

	// When
	ApplyChanges(&buf, changes)

	// Then — should contain cursor positioning and new text
	got := buf.String()
	if !strings.Contains(got, "bye") {
		t.Errorf("output should contain 'bye': %q", got)
	}
}

func TestApplyChangesTextWithStyle(t *testing.T) {
	// Given
	var buf bytes.Buffer
	style := render.Style{FG: render.Color{Name: "red"}}
	changes := []Change{
		{
			Old: &Box{X: 0, Y: 0, Width: 3, Height: 1, Content: "old"},
			New: &Box{X: 0, Y: 0, Width: 3, Height: 1, Content: "new", Style: style},
		},
	}

	// When
	ApplyChanges(&buf, changes)

	// Then — should contain SGR for red
	got := buf.String()
	if !strings.Contains(got, "\x1b[31m") {
		t.Errorf("output should contain red SGR: %q", got)
	}
}

func TestApplyChangesBorderUpdate(t *testing.T) {
	// Given
	var buf bytes.Buffer
	changes := []Change{
		{
			Old: &Box{X: 0, Y: 0, Width: 5, Height: 3, Border: "single"},
			New: &Box{X: 0, Y: 0, Width: 5, Height: 3, Border: "single",
				Style: render.Style{FG: render.Color{Name: "cyan"}}},
		},
	}

	// When
	ApplyChanges(&buf, changes)

	// Then — should contain border characters
	got := buf.String()
	if !strings.Contains(got, "┌") {
		t.Errorf("output should contain border characters: %q", got)
	}
}

func TestApplyChangesAddition(t *testing.T) {
	// Given — new node, no old
	var buf bytes.Buffer
	changes := []Change{
		{
			Old: nil,
			New: &Box{X: 0, Y: 0, Width: 5, Height: 1, Content: "added"},
		},
	}

	// When
	ApplyChanges(&buf, changes)

	// Then
	got := buf.String()
	if !strings.Contains(got, "added") {
		t.Errorf("output should contain 'added': %q", got)
	}
}

func TestApplyChangesRemoval(t *testing.T) {
	// Given — old node, no new
	var buf bytes.Buffer
	changes := []Change{
		{
			Old: &Box{X: 0, Y: 0, Width: 5, Height: 1, Content: "gone"},
			New: nil,
		},
	}

	// When
	ApplyChanges(&buf, changes)

	// Then — should clear the old region with spaces
	got := buf.String()
	if !strings.Contains(got, "     ") {
		t.Errorf("output should contain spaces to clear region: %q", got)
	}
}

func TestApplyChangesNoChanges(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	ApplyChanges(&buf, nil)

	// Then
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}
