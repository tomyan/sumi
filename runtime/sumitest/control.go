package sumitest

import (
	"encoding/json"
	"fmt"
	"net"
)

// Request is a command sent from the parent to the serve subprocess.
type Request struct {
	Cmd   string `json:"cmd"`             // "info", "step", "quit"
	Index int    `json:"index,omitempty"` // step index for "step" cmd
}

// InfoResponse is the response to an "info" command.
type InfoResponse struct {
	Name       string   `json:"name"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Steps      []string `json:"steps"`
	SourceFile string   `json:"source_file,omitempty"`
}

// StepResponse is the response to a "step" command.
type StepResponse struct {
	Name       string `json:"name"`
	StyledText string `json:"styled_text"`
}

// Client connects to a serve subprocess over a Unix socket.
type Client struct {
	conn net.Conn
	enc  *json.Encoder
	dec  *json.Decoder
}

// Connect dials the Unix socket at the given path.
func Connect(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return &Client{
		conn: conn,
		enc:  json.NewEncoder(conn),
		dec:  json.NewDecoder(conn),
	}, nil
}

// Info sends an "info" command and returns the scenario metadata.
func (c *Client) Info() (*InfoResponse, error) {
	if err := c.enc.Encode(Request{Cmd: "info"}); err != nil {
		return nil, fmt.Errorf("send info: %w", err)
	}
	var resp InfoResponse
	if err := c.dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("read info: %w", err)
	}
	return &resp, nil
}

// Step sends a "step" command and returns the step metadata.
// The ANSI output is written to stdout (the PTY), not the socket.
func (c *Client) Step(index int) (*StepResponse, error) {
	if err := c.enc.Encode(Request{Cmd: "step", Index: index}); err != nil {
		return nil, fmt.Errorf("send step: %w", err)
	}
	var resp StepResponse
	if err := c.dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("read step: %w", err)
	}
	return &resp, nil
}

// Quit sends a "quit" command to the subprocess.
func (c *Client) Quit() error {
	if err := c.enc.Encode(Request{Cmd: "quit"}); err != nil {
		return fmt.Errorf("send quit: %w", err)
	}
	return nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
