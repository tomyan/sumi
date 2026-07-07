package lsp

import (
	"strings"

	"github.com/tomyan/sumi/parser/section"
	"github.com/tomyan/sumi/parser/template"
)

// templateEntry is a template symbol awaiting position resolution. anchor is
// the substring whose occurrence marks the symbol's location.
type templateEntry struct {
	name   string
	anchor string
	kind   int
}

// templateSymbols collects id-bearing elements and component usages from the
// template, resolving each to a location by scanning the template text in
// document order.
func templateSymbols(text, uri string, s section.Sections) []SymbolInformation {
	if s.TemplateStart < 0 {
		return nil
	}
	doc, err := template.Parse(s.Template)
	if err != nil {
		return nil
	}
	entries := collectEntries(doc.Children)
	return resolveEntries(text, uri, s, entries)
}

// collectEntries walks nodes in pre-order gathering template symbols.
func collectEntries(nodes []template.Node) []templateEntry {
	var entries []templateEntry
	for _, n := range nodes {
		entries = append(entries, entryFor(n)...)
		entries = append(entries, collectEntries(childrenOf(n))...)
	}
	return entries
}

// entryFor returns the symbol(s) contributed by a single node.
func entryFor(n template.Node) []templateEntry {
	switch node := n.(type) {
	case *template.ComponentElement:
		return []templateEntry{{name: node.Name, anchor: "<" + node.Name, kind: SymbolObject}}
	case *template.BoxElement:
		return idEntry(node.Attributes)
	case *template.TextElement:
		return idEntry(node.Attributes)
	}
	return nil
}

// idEntry returns an id symbol when the attributes carry a non-empty id.
func idEntry(attrs map[string]string) []templateEntry {
	id, ok := attrs["id"]
	if !ok || id == "" {
		return nil
	}
	return []templateEntry{{name: id, anchor: "id=", kind: SymbolObject}}
}

// childrenOf returns the child nodes of a container node.
func childrenOf(n template.Node) []template.Node {
	switch node := n.(type) {
	case *template.BoxElement:
		return node.Children
	case *template.IfNode:
		return append(append([]template.Node{}, node.Then...), node.Else...)
	case *template.ForNode:
		return node.Children
	case *template.SnippetNode:
		return node.Children
	case *template.SlotDefNode:
		return node.Children
	case *template.SlotElement:
		return node.Default
	}
	return nil
}

// resolveEntries assigns each entry a location by advancing through the
// template text, so repeated names map to distinct occurrences.
func resolveEntries(text, uri string, s section.Sections, entries []templateEntry) []SymbolInformation {
	var syms []SymbolInformation
	cursor := 0
	for _, e := range entries {
		i := strings.Index(s.Template[cursor:], e.anchor)
		if i < 0 {
			continue
		}
		local := cursor + i
		cursor = local + len(e.anchor)
		syms = append(syms, symbolAt(text, uri, s.TemplateStart+local, e.name, e.kind))
	}
	return syms
}
