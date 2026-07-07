package lsp

import (
	"bytes"
	"encoding/json"
	"testing"
)

// responseFor finds the response result for a request id.
func responseFor(t *testing.T, msgs []*Message, id string) json.RawMessage {
	t.Helper()
	for _, m := range msgs {
		if string(m.ID) == id && m.Result != nil {
			return m.Result
		}
	}
	t.Fatalf("no response for id %s", id)
	return nil
}

func TestServerLanguageFeatures(t *testing.T) {
	// Given: a document opened, then feature requests against it
	uri := "file:///feat.sumi"
	doc := "<script>\nfunc go() {}\n</script>\n" +
		"<style>\n.a { color: red }\n</style>\n<button onclick={go}>x</button>"
	var in bytes.Buffer
	in.WriteString(frame(t, 1, "initialize", map[string]any{}))
	in.WriteString(frame(t, nil, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "text": doc},
	}))
	in.WriteString(frame(t, 2, "textDocument/completion", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 4, "character": 6},
	}))
	in.WriteString(frame(t, 3, "textDocument/hover", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 4, "character": 6},
	}))
	in.WriteString(frame(t, 4, "textDocument/documentSymbol", map[string]any{
		"textDocument": map[string]any{"uri": uri},
	}))
	in.WriteString(frame(t, 5, "textDocument/definition", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"position":     map[string]any{"line": 6, "character": 17},
	}))
	in.WriteString(frame(t, 6, "shutdown", nil))
	in.WriteString(frame(t, nil, "exit", nil))

	// When
	var out bytes.Buffer
	if err := NewServer(&in, &out).Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then
	msgs := readAll(t, &out)
	assertCompletionHasCSS(t, msgs)
	assertHoverHasContent(t, msgs)
	assertSymbolsHaveFunc(t, msgs)
	assertDefinitionInFile(t, msgs, uri)
}

func assertCompletionHasCSS(t *testing.T, msgs []*Message) {
	t.Helper()
	var items []CompletionItem
	if err := json.Unmarshal(responseFor(t, msgs, "2"), &items); err != nil {
		t.Fatalf("unmarshal completion: %v", err)
	}
	for _, it := range items {
		if it.Label == "color" {
			return
		}
	}
	t.Errorf("completion did not include color: %+v", items)
}

func assertHoverHasContent(t *testing.T, msgs []*Message) {
	t.Helper()
	var h Hover
	if err := json.Unmarshal(responseFor(t, msgs, "3"), &h); err != nil {
		t.Fatalf("unmarshal hover: %v", err)
	}
	if h.Contents.Value == "" {
		t.Error("hover content empty")
	}
}

func assertSymbolsHaveFunc(t *testing.T, msgs []*Message) {
	t.Helper()
	var syms []SymbolInformation
	if err := json.Unmarshal(responseFor(t, msgs, "4"), &syms); err != nil {
		t.Fatalf("unmarshal symbols: %v", err)
	}
	for _, s := range syms {
		if s.Name == "go" && s.Kind == SymbolFunction {
			return
		}
	}
	t.Errorf("documentSymbol missing func go: %+v", syms)
}

func assertDefinitionInFile(t *testing.T, msgs []*Message, uri string) {
	t.Helper()
	var loc Location
	if err := json.Unmarshal(responseFor(t, msgs, "5"), &loc); err != nil {
		t.Fatalf("unmarshal definition: %v", err)
	}
	if loc.URI != uri {
		t.Errorf("definition URI = %q, want %q", loc.URI, uri)
	}
	if loc.Range.Start.Line != 1 {
		t.Errorf("definition line = %d, want 1", loc.Range.Start.Line)
	}
}
