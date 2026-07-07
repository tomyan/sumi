package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

// frame encodes a JSON-RPC message with Content-Length framing.
func frame(t *testing.T, id any, method string, params any) string {
	t.Helper()
	m := map[string]any{"jsonrpc": "2.0"}
	if id != nil {
		m["id"] = id
	}
	if method != "" {
		m["method"] = method
	}
	if params != nil {
		m["params"] = params
	}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

// readAll decodes every framed message written to buf.
func readAll(t *testing.T, buf *bytes.Buffer) []*Message {
	t.Helper()
	conn := NewConn(bytes.NewReader(buf.Bytes()), io.Discard)
	var msgs []*Message
	for {
		m, err := conn.Read()
		if err != nil {
			return msgs
		}
		msgs = append(msgs, m)
	}
}

// diagnosticsFor finds the publishDiagnostics notification for a URI.
func diagnosticsFor(t *testing.T, msgs []*Message, uri string) ([]Diagnostic, bool) {
	t.Helper()
	var last []Diagnostic
	found := false
	for _, m := range msgs {
		if m.Method != "textDocument/publishDiagnostics" {
			continue
		}
		var p PublishDiagnosticsParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal diagnostics: %v", err)
		}
		if p.URI == uri {
			last = p.Diagnostics
			found = true
		}
	}
	return last, found
}

func TestServerFullSession(t *testing.T) {
	// Given: a session that opens a broken file then fixes it
	uri := "file:///test.sumi"
	broken := "<div></div>\n@oops"
	fixed := "<div>ok</div>"
	var in bytes.Buffer
	in.WriteString(frame(t, 1, "initialize", map[string]any{}))
	in.WriteString(frame(t, nil, "initialized", map[string]any{}))
	in.WriteString(frame(t, nil, "textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{"uri": uri, "text": broken},
	}))
	in.WriteString(frame(t, nil, "textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri},
		"contentChanges": []map[string]any{{"text": fixed}},
	}))
	in.WriteString(frame(t, 2, "shutdown", nil))
	in.WriteString(frame(t, nil, "exit", nil))

	// When
	var out bytes.Buffer
	if err := NewServer(&in, &out).Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Then
	msgs := readAll(t, &out)
	assertInitializeResult(t, msgs)
	assertDiagnosticsProgression(t, msgs, uri)
}

func assertInitializeResult(t *testing.T, msgs []*Message) {
	t.Helper()
	for _, m := range msgs {
		if len(m.ID) == 0 || string(m.ID) != "1" {
			continue
		}
		var res InitializeResult
		if err := json.Unmarshal(m.Result, &res); err != nil {
			t.Fatalf("unmarshal initialize result: %v", err)
		}
		if res.Capabilities.TextDocumentSync != 1 {
			t.Errorf("textDocumentSync = %d, want 1", res.Capabilities.TextDocumentSync)
		}
		return
	}
	t.Fatal("no initialize response found")
}

func assertDiagnosticsProgression(t *testing.T, msgs []*Message, uri string) {
	t.Helper()
	// The final published diagnostics (after didChange) must be empty.
	diags, ok := diagnosticsFor(t, msgs, uri)
	if !ok {
		t.Fatal("no diagnostics published for uri")
	}
	if len(diags) != 0 {
		t.Errorf("final diagnostics = %+v, want empty after fix", diags)
	}

	// The first publish (from didOpen of the broken file) must be non-empty
	// and point at the second line.
	first, firstOK := firstDiagnostics(t, msgs, uri)
	if !firstOK {
		t.Fatal("no first diagnostics found")
	}
	if len(first) == 0 {
		t.Fatal("first diagnostics empty, want an error for the broken file")
	}
	if first[0].Range.Start.Line != 1 {
		t.Errorf("first diagnostic line = %d, want 1", first[0].Range.Start.Line)
	}
}

func firstDiagnostics(t *testing.T, msgs []*Message, uri string) ([]Diagnostic, bool) {
	t.Helper()
	for _, m := range msgs {
		if m.Method != "textDocument/publishDiagnostics" {
			continue
		}
		var p PublishDiagnosticsParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal diagnostics: %v", err)
		}
		if p.URI == uri {
			return p.Diagnostics, true
		}
	}
	return nil, false
}
