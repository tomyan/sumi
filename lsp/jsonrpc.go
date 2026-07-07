package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Message is a raw JSON-RPC 2.0 envelope. Params and Result stay as
// json.RawMessage so the server can decode them into method-specific types.
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError is the error object of a JSON-RPC response.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Conn reads and writes LSP messages using stdio Content-Length framing.
type Conn struct {
	r *bufio.Reader
	w io.Writer
}

// NewConn wraps a reader/writer pair as an LSP connection.
func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{r: bufio.NewReader(r), w: w}
}

// Read reads one framed message, returning io.EOF when the stream closes
// cleanly between messages.
func (c *Conn) Read() (*Message, error) {
	length, err := readContentLength(c.r)
	if err != nil {
		return nil, err
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, err
	}
	var m Message
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Write frames and writes one message with a Content-Length header.
func (c *Conn) Write(m *Message) error {
	m.JSONRPC = "2.0"
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err = c.w.Write(body)
	return err
}

// readContentLength consumes the header block and returns the body length.
func readContentLength(r *bufio.Reader) (int, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if length == -1 {
				return 0, fmt.Errorf("missing Content-Length header")
			}
			return length, nil
		}
		if v, ok := parseContentLength(line); ok {
			length = v
		}
	}
}

// parseContentLength extracts the value of a Content-Length header line.
func parseContentLength(line string) (int, bool) {
	const prefix = "Content-Length:"
	if !strings.HasPrefix(line, prefix) {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimSpace(line[len(prefix):]))
	if err != nil {
		return 0, false
	}
	return v, true
}
