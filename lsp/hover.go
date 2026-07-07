package lsp

import (
	"strings"

	"github.com/tomyan/sumi/runtime/css"
)

// cssNotes holds sumi-specific guidance for properties whose terminal
// behaviour differs from the browser. Properties absent here get the generic
// "supported terminal CSS property" line only.
var cssNotes = map[string]string{
	"border":          "Terminal border style: single | double | rounded | heavy | ascii.",
	"border-title":    "Text drawn into the top edge of a bordered box.",
	"opacity":         "Values below 1 dim the cell — terminals have no true alpha channel.",
	"border-collapse": "Merges adjacent cell borders, drawing junction characters at joins.",
}

// tagNotes holds one-line UA behaviour notes for interactive elements. Other
// tags get a generic element note.
var tagNotes = map[string]string{
	"button":  "Focusable; Enter or Space activates it (synthesizes a click).",
	"input":   "Focusable control; edits on keypress or toggles when checkable.",
	"select":  "Focusable; cycles through its options and shows the selected label.",
	"dialog":  "When open, traps focus as a modal scope.",
	"details": "Toggles its body open or closed via its <summary>.",
	"a":       "Focusable link; activation follows its href.",
}

// HoverAt returns hover content for the cursor at pos, or nil when there is
// nothing to describe.
func HoverAt(text string, pos Position) *Hover {
	offset := positionToOffset(text, pos)
	word := wordAt(text, offset)
	if word == "" {
		return nil
	}
	switch ClassifyContext(text, pos) {
	case ContextStyle:
		return propertyHover(word)
	case ContextTagName:
		return tagHover(word)
	default:
		return nil
	}
}

// propertyHover describes a CSS property, or nil when the word is not one.
// A trailing colon (from "prop:") is trimmed before lookup.
func propertyHover(prop string) *Hover {
	prop = strings.TrimRight(prop, ":")
	if !supported(prop) {
		return nil
	}
	value := "`" + prop + "` — supported terminal CSS property."
	if note, ok := cssNotes[prop]; ok {
		value += "\n\n" + note
	}
	return markdownHover(value)
}

// tagHover describes an HTML tag's terminal UA behaviour.
func tagHover(tag string) *Hover {
	note, ok := tagNotes[tag]
	if !ok {
		note = "HTML element."
	}
	return markdownHover("`<" + tag + ">` — " + note)
}

// markdownHover wraps a markdown string as a hover result.
func markdownHover(value string) *Hover {
	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: value}}
}

// supported reports whether name is a property sumi consumes.
func supported(name string) bool {
	for _, p := range css.SupportedProperties() {
		if p == name {
			return true
		}
	}
	return false
}

// wordAt returns the identifier-like token surrounding offset: a run of
// letters, digits, '-', ':' or '_'. Empty when offset is not on such a token.
func wordAt(text string, offset int) string {
	start, end := offset, offset
	for start > 0 && isWordByte(text[start-1]) {
		start--
	}
	for end < len(text) && isWordByte(text[end]) {
		end++
	}
	return text[start:end]
}

// isWordByte reports whether b can be part of a hover word token.
func isWordByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	case b == '-', b == ':', b == '_':
		return true
	}
	return false
}
