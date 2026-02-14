package sumitest

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// sentinel is the OSC sequence written after each rendered frame to signal
// that the parent has received all ANSI output for this step.
const sentinel = "\x1b]999;done\x07"

// Serve enters serve mode for the given scenario, listening on the Unix socket
// specified by SUMI_CONTROL_SOCKET. It replays steps on demand, writing ANSI
// output to stdout and responding with metadata on the socket.
func Serve(s Scenario) {
	socketPath := os.Getenv("SUMI_CONTROL_SOCKET")
	if socketPath == "" {
		fmt.Fprintln(os.Stderr, "serve: SUMI_CONTROL_SOCKET not set")
		os.Exit(1)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: listen %s: %v\n", socketPath, err)
		os.Exit(1)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Fprintf(os.Stderr, "serve: accept: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	handleConnection(conn, s)
}

// ServeOnListener is like Serve but uses a provided listener, for testing.
func ServeOnListener(listener net.Listener, s Scenario, stdout *os.File) {
	conn, err := listener.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	handleConnectionTo(conn, s, stdout)
}

func handleConnection(conn net.Conn, s Scenario) {
	handleConnectionTo(conn, s, os.Stdout)
}

func handleConnectionTo(conn net.Conn, s Scenario, stdout *os.File) {
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			return // connection closed
		}

		switch req.Cmd {
		case "info":
			enc.Encode(buildInfoResponse(s))
		case "step":
			resp := executeStep(s, req.Index, stdout)
			enc.Encode(resp)
		case "quit":
			return
		}
	}
}

func buildInfoResponse(s Scenario) InfoResponse {
	stepNames := make([]string, len(s.Steps))
	for i, step := range s.Steps {
		stepNames[i] = step.Name
	}
	return InfoResponse{
		Name:       s.Name,
		Width:      s.Width,
		Height:     s.Height,
		Steps:      stepNames,
		SourceFile: s.SourceFile,
	}
}

func executeStep(s Scenario, index int, stdout *os.File) StepResponse {
	app := s.NewApp(s.Width, s.Height)
	h := New(app)

	// Replay all steps up to and including the requested index.
	limit := index
	if limit >= len(s.Steps) {
		limit = len(s.Steps) - 1
	}
	for i := 0; i <= limit; i++ {
		if s.Steps[i].Action != nil {
			s.Steps[i].Action(h)
		}
	}

	// Write ANSI rendering to stdout (the PTY), followed by sentinel.
	if buf := h.Buffer(); buf != nil {
		buf.RenderTo(stdout)
	}
	fmt.Fprint(stdout, sentinel)

	return StepResponse{
		Name:       s.Steps[limit].Name,
		StyledText: h.StyledText(),
	}
}
