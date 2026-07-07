package lsp

import (
	"strings"

	"github.com/tomyan/sumi/parser/section"
)

// Context classifies where a cursor sits within a .sumi document. Completion
// and hover use it to decide which vocabulary to offer.
type Context int

const (
	// ContextTemplate is anywhere in the template that is not a tag name or
	// attribute name (text bodies, expressions, whitespace between elements).
	ContextTemplate Context = iota
	// ContextScript is inside the <script> section.
	ContextScript
	// ContextStyle is inside the <style> section.
	ContextStyle
	// ContextTagName is immediately after `<` or `</`, within a tag name.
	ContextTagName
	// ContextAttrName is inside an open tag, after whitespace following the
	// tag name.
	ContextAttrName
)

// ClassifyContext reports which section and, within the template, which token
// context the cursor occupies.
func ClassifyContext(text string, pos Position) Context {
	offset := positionToOffset(text, pos)
	sections, err := section.Parse(text)
	if err != nil {
		return ContextTemplate
	}
	if within(offset, sections.ScriptStart, len(sections.Script)) {
		return ContextScript
	}
	if within(offset, sections.StyleStart, len(sections.Style)) {
		return ContextStyle
	}
	return classifyTemplate(text, offset)
}

// within reports whether offset falls in the half-open region [start, start+n)
// with a valid (non-negative) start.
func within(offset, start, n int) bool {
	return start >= 0 && offset >= start && offset < start+n
}

// classifyTemplate distinguishes tag-name, attribute-name, and plain template
// contexts by scanning back to the nearest tag delimiter.
func classifyTemplate(text string, offset int) Context {
	open := nearestOpenTag(text, offset)
	if open < 0 {
		return ContextTemplate
	}
	inner := strings.TrimPrefix(text[open+1:offset], "/")
	if strings.ContainsAny(inner, " \t\r\n") {
		return ContextAttrName
	}
	return ContextTagName
}

// nearestOpenTag returns the index of the `<` that opens the tag enclosing
// offset, or -1 when the cursor is not inside a tag (a `>` is seen first, or
// no `<` precedes it).
func nearestOpenTag(text string, offset int) int {
	for i := offset - 1; i >= 0; i-- {
		switch text[i] {
		case '>':
			return -1
		case '<':
			return i
		}
	}
	return -1
}
