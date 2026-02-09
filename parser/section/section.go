package section

import "strings"

// Sections holds the three optional parts of a .sumi file.
type Sections struct {
	Script   string
	Style    string
	Template string
}

// Parse splits a .sumi file's content into its three optional sections.
// Script is content between <script> and </script> tags.
// Style is content between <style> and </style> tags.
// Template is everything else, with surrounding whitespace trimmed.
func Parse(input string) (Sections, error) {
	var s Sections
	remaining := input

	remaining, s.Script = extractSection(remaining, "<script>", "</script>")
	remaining, s.Style = extractSection(remaining, "<style>", "</style>")

	s.Template = strings.TrimSpace(remaining)
	return s, nil
}

// extractSection finds and removes a section delimited by open/close tags.
// Returns the remaining input with the section removed, and the content between the tags.
func extractSection(input, openTag, closeTag string) (remaining, content string) {
	openIdx := strings.Index(input, openTag)
	if openIdx == -1 {
		return input, ""
	}
	closeIdx := strings.Index(input, closeTag)
	if closeIdx == -1 {
		return input, ""
	}

	content = input[openIdx+len(openTag) : closeIdx]
	remaining = input[:openIdx] + input[closeIdx+len(closeTag):]
	return remaining, content
}
