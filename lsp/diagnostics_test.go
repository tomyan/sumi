package lsp

import (
	"strings"
	"testing"
)

func TestDiagnosticsCleanTemplateHasNone(t *testing.T) {
	// Given
	text := "<div>hello</div>"

	// When
	diags := Diagnostics(text)

	// Then
	if len(diags) != 0 {
		t.Fatalf("got %d diagnostics, want 0: %+v", len(diags), diags)
	}
}

func TestDiagnosticsTemplateErrorHasPreciseRange(t *testing.T) {
	// Given: a stray character on the second line
	text := "<div></div>\n@oops"

	// When
	diags := Diagnostics(text)

	// Then
	if len(diags) != 1 {
		t.Fatalf("got %d diagnostics, want 1: %+v", len(diags), diags)
	}
	d := diags[0]
	if d.Range.Start.Line != 1 || d.Range.Start.Character != 0 {
		t.Errorf("range start = %+v, want {Line:1 Character:0}", d.Range.Start)
	}
	if !strings.Contains(d.Message, "unexpected character") {
		t.Errorf("message = %q, want it to mention the unexpected character", d.Message)
	}
	if d.Source != "sumi" {
		t.Errorf("source = %q, want sumi", d.Source)
	}
}

func TestDiagnosticsTemplateOffsetAccountsForScriptSection(t *testing.T) {
	// Given: a script section pushes the template error down the file
	text := "<script>\nx := sumi.New(0)\n</script>\n\n<div></div>\n@oops"

	// When
	diags := Diagnostics(text)

	// Then: the diagnostic lands on the line holding '@oops'
	if len(diags) != 1 {
		t.Fatalf("got %d diagnostics, want 1: %+v", len(diags), diags)
	}
	wantLine := strings.Count(text[:strings.Index(text, "@oops")], "\n")
	if diags[0].Range.Start.Line != wantLine {
		t.Errorf("range start line = %d, want %d", diags[0].Range.Start.Line, wantLine)
	}
}

func TestDiagnosticsNeverPanicsOnGarbage(t *testing.T) {
	// Given: assorted malformed and truncated inputs
	inputs := []string{
		"",
		"<",
		"<div",
		"<div>{if}",
		"{for",
		"<script>\n</script>\n<style>\n.a{",
		"\x00\x01\xff garbage",
		strings.Repeat("<div>", 100),
		"<slot:",
		"{snippet x(",
	}

	// When / Then: must not panic and must return a (possibly empty) slice
	for _, in := range inputs {
		diags := Diagnostics(in)
		if diags == nil {
			t.Errorf("Diagnostics(%q) = nil, want non-nil slice", in)
		}
	}
}
