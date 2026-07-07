package lsp

// Position is a zero-based line and UTF-16 character offset within a line.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a half-open span between two positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Diagnostic severities as defined by the LSP specification.
const (
	SeverityError   = 1
	SeverityWarning = 2
)

// Diagnostic is a single problem reported for a document.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

// PublishDiagnosticsParams is the payload of textDocument/publishDiagnostics.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// InitializeResult is returned from the initialize request.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities advertises what the server supports. The completion,
// hover, documentSymbol and definition fields are placeholders that later
// slices populate; they stay omitted until then.
type ServerCapabilities struct {
	TextDocumentSync       int                `json:"textDocumentSync"`
	CompletionProvider     *CompletionOptions `json:"completionProvider,omitempty"`
	HoverProvider          bool               `json:"hoverProvider,omitempty"`
	DocumentSymbolProvider bool               `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider     bool               `json:"definitionProvider,omitempty"`
}

// CompletionOptions is the completion capability object (unused until the
// completion slice fills it).
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// TextDocumentItem is a document announced by textDocument/didOpen.
type TextDocumentItem struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

// DidOpenTextDocumentParams is the payload of textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// VersionedTextDocumentIdentifier references a document by URI.
type VersionedTextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// TextDocumentContentChangeEvent is one change. With full sync the Text
// field holds the entire new document.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// DidChangeTextDocumentParams is the payload of textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}
