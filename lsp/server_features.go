package lsp

import "encoding/json"

// completion answers textDocument/completion with context-appropriate items.
func (s *Server) completion(id, params json.RawMessage) error {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	text := s.documents[p.TextDocument.URI]
	items := Completions(text, p.Position, p.TextDocument.URI)
	if items == nil {
		items = []CompletionItem{}
	}
	return s.reply(id, items)
}

// hover answers textDocument/hover, replying null when there is nothing to
// describe.
func (s *Server) hover(id, params json.RawMessage) error {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	text := s.documents[p.TextDocument.URI]
	return s.reply(id, HoverAt(text, p.Position))
}

// documentSymbol answers textDocument/documentSymbol with a flat symbol list.
func (s *Server) documentSymbol(id, params json.RawMessage) error {
	var p DocumentSymbolParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	text := s.documents[p.TextDocument.URI]
	syms := DocumentSymbols(text, p.TextDocument.URI)
	if syms == nil {
		syms = []SymbolInformation{}
	}
	return s.reply(id, syms)
}

// definition answers textDocument/definition, replying null when unresolved.
func (s *Server) definition(id, params json.RawMessage) error {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	text := s.documents[p.TextDocument.URI]
	return s.reply(id, Definition(text, p.Position, p.TextDocument.URI))
}
