package lsp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteFramesWithContentLength(t *testing.T) {
	// Given
	var buf bytes.Buffer
	c := NewConn(&buf, &buf)

	// When
	err := c.Write(&Message{Method: "ping"})

	// Then
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "Content-Length: ") {
		t.Errorf("output missing Content-Length header: %q", out)
	}
	if !strings.Contains(out, "\r\n\r\n") {
		t.Errorf("output missing header/body separator: %q", out)
	}
	if !strings.Contains(out, `"jsonrpc":"2.0"`) {
		t.Errorf("Write did not stamp jsonrpc version: %q", out)
	}
}

func TestReadWriteRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		msg  Message
	}{
		{"request", Message{ID: json.RawMessage(`1`), Method: "initialize", Params: json.RawMessage(`{"a":1}`)}},
		{"notification", Message{Method: "textDocument/didOpen", Params: json.RawMessage(`{}`)}},
		{"response", Message{ID: json.RawMessage(`2`), Result: json.RawMessage(`{"ok":true}`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var buf bytes.Buffer
			c := NewConn(&buf, &buf)

			// When
			if err := c.Write(&tc.msg); err != nil {
				t.Fatalf("Write: %v", err)
			}
			got, err := c.Read()

			// Then
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if got.Method != tc.msg.Method {
				t.Errorf("Method = %q, want %q", got.Method, tc.msg.Method)
			}
			if string(got.Params) != string(tc.msg.Params) {
				t.Errorf("Params = %q, want %q", got.Params, tc.msg.Params)
			}
			if string(got.Result) != string(tc.msg.Result) {
				t.Errorf("Result = %q, want %q", got.Result, tc.msg.Result)
			}
		})
	}
}

func TestReadMissingContentLengthErrors(t *testing.T) {
	// Given: a header block with no Content-Length
	in := strings.NewReader("X-Other: 1\r\n\r\n")
	c := NewConn(in, &bytes.Buffer{})

	// When
	_, err := c.Read()

	// Then
	if err == nil {
		t.Fatal("expected error for missing Content-Length")
	}
}
