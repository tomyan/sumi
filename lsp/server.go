package lsp

import (
	"encoding/json"
	"errors"
	"io"
)

// Server is a minimal LSP server that publishes diagnostics for .sumi files.
// Documents are held in memory with full text synchronisation.
type Server struct {
	conn      *Conn
	documents map[string]string
	exited    bool
}

// NewServer builds a server reading requests from r and writing responses
// and notifications to w.
func NewServer(r io.Reader, w io.Writer) *Server {
	return &Server{conn: NewConn(r, w), documents: map[string]string{}}
}

// Run processes messages until the stream closes or an exit is requested.
func (s *Server) Run() error {
	for {
		msg, err := s.conn.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := s.handle(msg); err != nil {
			return err
		}
		if s.exited {
			return nil
		}
	}
}

// handle dispatches one message by method.
func (s *Server) handle(msg *Message) error {
	switch msg.Method {
	case "initialize":
		return s.reply(msg.ID, InitializeResult{Capabilities: capabilities()})
	case "initialized":
		return nil
	case "shutdown":
		return s.reply(msg.ID, nil)
	case "exit":
		s.exited = true
		return nil
	case "textDocument/didOpen":
		return s.didOpen(msg.Params)
	case "textDocument/didChange":
		return s.didChange(msg.Params)
	case "textDocument/completion":
		return s.completion(msg.ID, msg.Params)
	case "textDocument/hover":
		return s.hover(msg.ID, msg.Params)
	case "textDocument/documentSymbol":
		return s.documentSymbol(msg.ID, msg.Params)
	case "textDocument/definition":
		return s.definition(msg.ID, msg.Params)
	default:
		return nil
	}
}

// capabilities advertises full-text sync plus the language features the
// server implements.
func capabilities() ServerCapabilities {
	return ServerCapabilities{
		TextDocumentSync:       1,
		CompletionProvider:     &CompletionOptions{TriggerCharacters: []string{"<", " "}},
		HoverProvider:          true,
		DocumentSymbolProvider: true,
		DefinitionProvider:     true,
	}
}

// didOpen tracks a newly opened document and publishes its diagnostics.
func (s *Server) didOpen(params json.RawMessage) error {
	var p DidOpenTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	s.documents[p.TextDocument.URI] = p.TextDocument.Text
	return s.publish(p.TextDocument.URI)
}

// didChange applies a full-sync change and republishes diagnostics.
func (s *Server) didChange(params json.RawMessage) error {
	var p DidChangeTextDocumentParams
	if err := json.Unmarshal(params, &p); err != nil {
		return err
	}
	if len(p.ContentChanges) == 0 {
		return nil
	}
	uri := p.TextDocument.URI
	s.documents[uri] = p.ContentChanges[len(p.ContentChanges)-1].Text
	return s.publish(uri)
}

// publish computes and sends diagnostics for a document.
func (s *Server) publish(uri string) error {
	diags := Diagnostics(s.documents[uri])
	return s.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	})
}

// reply sends a JSON-RPC response to a request.
func (s *Server) reply(id json.RawMessage, result any) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return s.conn.Write(&Message{ID: id, Result: raw})
}

// notify sends a JSON-RPC notification.
func (s *Server) notify(method string, params any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return s.conn.Write(&Message{Method: method, Params: raw})
}
