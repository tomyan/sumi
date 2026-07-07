package section

import "strings"

// Sections holds the optional parts of a .sumi file.
type Sections struct {
	Imports  string
	Script   string
	Style    string
	Template string

	// ScriptStart, StyleStart and TemplateStart are the byte offsets of each
	// section's content within the original input, or -1 when the section is
	// absent. Editor diagnostics use them to map parse errors back to the
	// source file. Offsets assume the conventional document order
	// (imports, script, style, template).
	ScriptStart   int
	StyleStart    int
	TemplateStart int
}

// Parse splits a .sumi file's content into its optional sections.
// Script is content between <script> and </script> tags.
// Style is content between <style> and </style> tags.
// Template is everything else, with surrounding whitespace trimmed.
func Parse(input string) (Sections, error) {
	s := Sections{ScriptStart: -1, StyleStart: -1, TemplateStart: -1}
	remaining := input
	removed := 0

	var start int
	remaining, s.Imports, start = extractSection(remaining, "<sumi:imports>", "</sumi:imports>")
	removed = advance(removed, start, "<sumi:imports>", s.Imports, "</sumi:imports>")

	remaining, s.Script, start = extractSection(remaining, "<script>", "</script>")
	s.ScriptStart = originalStart(start, removed)
	removed = advance(removed, start, "<script>", s.Script, "</script>")

	remaining, s.Style, start = extractSection(remaining, "<style>", "</style>")
	s.StyleStart = originalStart(start, removed)
	removed = advance(removed, start, "<style>", s.Style, "</style>")

	s.Template = strings.TrimSpace(remaining)
	s.TemplateStart = templateStart(remaining, s.Template, removed)
	return s, nil
}

// extractSection finds and removes a section delimited by open/close tags.
// Returns the remaining input with the section removed, the content between
// the tags, and the content's start offset within the given input (-1 when
// the section is absent).
func extractSection(input, openTag, closeTag string) (remaining, content string, start int) {
	openIdx := strings.Index(input, openTag)
	if openIdx == -1 {
		return input, "", -1
	}
	closeIdx := strings.Index(input, closeTag)
	if closeIdx == -1 {
		return input, "", -1
	}
	start = openIdx + len(openTag)
	content = input[start:closeIdx]
	remaining = input[:openIdx] + input[closeIdx+len(closeTag):]
	return remaining, content, start
}

// advance grows the running count of bytes removed from earlier in the input
// once a section has been extracted, so later offsets can be mapped back to
// the original input.
func advance(removed, start int, openTag, content, closeTag string) int {
	if start == -1 {
		return removed
	}
	return removed + len(openTag) + len(content) + len(closeTag)
}

// originalStart maps a content start offset in the trimmed input back to the
// original input, or returns -1 when the section is absent.
func originalStart(start, removed int) int {
	if start == -1 {
		return -1
	}
	return start + removed
}

// templateStart returns the original-input offset of the trimmed template
// content, or -1 when the template is empty.
func templateStart(remaining, template string, removed int) int {
	if template == "" {
		return -1
	}
	return strings.Index(remaining, template) + removed
}
