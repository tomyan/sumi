package lsp

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tomyan/sumi/parser/template"
	"github.com/tomyan/sumi/runtime/css"
)

// globalAttrs is the attribute vocabulary offered in attr-name context. On*
// handlers are open-ended in codegen; this is the curated recognised set.
var globalAttrs = []string{
	"id", "class", "focusable", "onkey", "onclick", "onkeydown",
	"onfocus", "onblur", "onchange", "onclose", "bind:value",
	"disabled", "checked", "open",
}

// Completions returns the completion items for the cursor at pos in a document
// with the given text and URI. The URI locates sibling component files.
func Completions(text string, pos Position, uri string) []CompletionItem {
	switch ClassifyContext(text, pos) {
	case ContextStyle:
		return propertyItems()
	case ContextTagName:
		return tagItems(uri)
	case ContextAttrName:
		return attrItems()
	default:
		return nil
	}
}

// propertyItems lists supported CSS property names.
func propertyItems() []CompletionItem {
	var items []CompletionItem
	for _, name := range css.SupportedProperties() {
		items = append(items, CompletionItem{Label: name, Kind: KindProperty})
	}
	return items
}

// attrItems lists the recognised global attributes.
func attrItems() []CompletionItem {
	var items []CompletionItem
	for _, name := range globalAttrs {
		items = append(items, CompletionItem{Label: name, Kind: KindField})
	}
	return items
}

// tagItems lists HTML element tags and sibling component names.
func tagItems(uri string) []CompletionItem {
	var items []CompletionItem
	for _, name := range template.HTMLTagNames() {
		items = append(items, CompletionItem{Label: name, Kind: KindKeyword})
	}
	for _, name := range siblingComponents(uri) {
		items = append(items, CompletionItem{Label: name, Kind: KindClass})
	}
	return items
}

// siblingComponents lists the PascalCase names of other .sumi files in the
// same directory as the document. An unreadable directory yields none.
func siblingComponents(uri string) []string {
	dir := filepath.Dir(uriToPath(uri))
	self := filepath.Base(uriToPath(uri))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || e.Name() == self || !strings.HasSuffix(e.Name(), ".sumi") {
			continue
		}
		names = append(names, template.ExportedComponentName(template.ComponentName(e.Name())))
	}
	return names
}

// uriToPath strips a file:// scheme from a document URI.
func uriToPath(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}
