package sumitest

import (
	"strings"
	"testing"
)

// AssertText checks that the harness plain text output matches expected exactly.
func AssertText(t testing.TB, h *Harness, expected string) {
	t.Helper()
	got := h.Text()
	if got != expected {
		t.Errorf("text mismatch:\ngot:  %q\nwant: %q", got, expected)
	}
}

// AssertStyledText checks that the harness styled text output matches expected exactly.
func AssertStyledText(t testing.TB, h *Harness, expected string) {
	t.Helper()
	got := h.StyledText()
	if got != expected {
		t.Errorf("styled text mismatch:\ngot:  %q\nwant: %q", got, expected)
	}
}

// AssertContains checks that the harness plain text output contains the substring.
func AssertContains(t testing.TB, h *Harness, substring string) {
	t.Helper()
	got := h.Text()
	if !strings.Contains(got, substring) {
		t.Errorf("text does not contain %q:\ngot: %q", substring, got)
	}
}

// AssertStyledContains checks that the harness styled text output contains the substring.
func AssertStyledContains(t testing.TB, h *Harness, substring string) {
	t.Helper()
	got := h.StyledText()
	if !strings.Contains(got, substring) {
		t.Errorf("styled text does not contain %q:\ngot: %q", substring, got)
	}
}
