package lsp

import "testing"

// fixture is a component exercising every section for context tests.
const contextFixture = "<script>\nfunc greet() {}\n</script>\n" +
	"<style>\n.a { color: red }\n</style>\n" +
	"<div id=\"root\">\n<span>Hi</span>\n</div>\n"

func TestClassifyContext(t *testing.T) {
	// Given: a fixture with script, style and template sections
	cases := []struct {
		name   string
		cursor Position
		want   Context
	}{
		{"in script body", Position{Line: 1, Character: 5}, ContextScript},
		{"in style body", Position{Line: 4, Character: 5}, ContextStyle},
		{"after open angle", Position{Line: 6, Character: 1}, ContextTagName},
		{"partway through tag", Position{Line: 6, Character: 3}, ContextTagName},
		{"after tag whitespace", Position{Line: 6, Character: 5}, ContextAttrName},
		{"inside text body", Position{Line: 7, Character: 7}, ContextTemplate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// When
			got := ClassifyContext(contextFixture, tc.cursor)

			// Then
			if got != tc.want {
				t.Errorf("ClassifyContext(%v) = %v, want %v", tc.cursor, got, tc.want)
			}
		})
	}
}

func TestClassifyClosingTagName(t *testing.T) {
	// Given: a cursor right after a "</" closing sequence
	text := "<div></div>"

	// When: cursor sits after the slash of the closing tag
	got := ClassifyContext(text, Position{Line: 0, Character: 7})

	// Then
	if got != ContextTagName {
		t.Errorf("closing tag context = %v, want ContextTagName", got)
	}
}
