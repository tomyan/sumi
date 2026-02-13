package render

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCopyToClipboard(t *testing.T) {
	// Given
	var buf bytes.Buffer
	text := "hello world"

	// When
	CopyToClipboard(&buf, text)

	// Then
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	expected := "\x1b]52;c;" + encoded + "\x07"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}

func TestCopyToClipboardEmpty(t *testing.T) {
	// Given
	var buf bytes.Buffer

	// When
	CopyToClipboard(&buf, "")

	// Then
	encoded := base64.StdEncoding.EncodeToString([]byte(""))
	expected := "\x1b]52;c;" + encoded + "\x07"
	if buf.String() != expected {
		t.Errorf("got %q, want %q", buf.String(), expected)
	}
}
