package lsp

import (
	"regexp"

	"github.com/tomyan/sumi/parser/script"
	"github.com/tomyan/sumi/parser/section"
)

// stateDeclRe matches a signal/computed declaration: name := sumi.New(...)
// or signal.From(...). The parser does not yet model these, so the LSP
// extracts them directly.
var stateDeclRe = regexp.MustCompile(`(\w+)\s*:=\s*(?:sumi|signal)\.(?:New|From)\b`)

// DocumentSymbols returns a flat symbol list for a document: script functions
// and state declarations, template elements carrying an id, and component
// usages.
func DocumentSymbols(text, uri string) []SymbolInformation {
	sections, err := section.Parse(text)
	if err != nil {
		return nil
	}
	var syms []SymbolInformation
	syms = append(syms, scriptSymbols(text, uri, sections)...)
	syms = append(syms, templateSymbols(text, uri, sections)...)
	return syms
}

// scriptSymbols collects function and state-declaration symbols.
func scriptSymbols(text, uri string, s section.Sections) []SymbolInformation {
	if s.ScriptStart < 0 {
		return nil
	}
	var syms []SymbolInformation
	syms = append(syms, funcSymbols(text, uri, s)...)
	syms = append(syms, stateSymbols(text, uri, s)...)
	return syms
}

// funcSymbols locates each declared function's name in the script section.
func funcSymbols(text, uri string, s section.Sections) []SymbolInformation {
	sc, err := script.Parse(s.Script)
	if err != nil {
		return nil
	}
	var syms []SymbolInformation
	for _, fn := range sc.FuncDecls {
		re := regexp.MustCompile(`func\s+(` + regexp.QuoteMeta(fn.Name) + `)\b`)
		loc := re.FindStringSubmatchIndex(s.Script)
		if loc == nil {
			continue
		}
		syms = append(syms, symbolAt(text, uri, s.ScriptStart+loc[2], fn.Name, SymbolFunction))
	}
	return syms
}

// stateSymbols locates signal/computed declarations in the script section.
func stateSymbols(text, uri string, s section.Sections) []SymbolInformation {
	var syms []SymbolInformation
	for _, m := range stateDeclRe.FindAllStringSubmatchIndex(s.Script, -1) {
		name := s.Script[m[2]:m[3]]
		syms = append(syms, symbolAt(text, uri, s.ScriptStart+m[2], name, SymbolVariable))
	}
	return syms
}

// symbolAt builds a symbol whose range spans name starting at the given byte
// offset in the original text.
func symbolAt(text, uri string, offset int, name string, kind int) SymbolInformation {
	start := offsetToPosition(text, offset)
	end := offsetToPosition(text, offset+len(name))
	return SymbolInformation{
		Name:     name,
		Kind:     kind,
		Location: Location{URI: uri, Range: Range{Start: start, End: end}},
	}
}
