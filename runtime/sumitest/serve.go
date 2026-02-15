package sumitest

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/tui"
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

	handleConnectionTo(conn, s, os.Stdout)
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

// serveState holds persistent app/harness state across step commands.
type serveState struct {
	scenario Scenario
	stdout   *os.File
	app      *tui.App
	harness  *Harness
	current  int // last applied step index, -1 = pristine
}

func newServeState(s Scenario, stdout *os.File) *serveState {
	return &serveState{scenario: s, stdout: stdout, current: -1}
}

// stepTo applies the scenario up to the given index and returns the response.
// Forward steps reuse the existing app; backward steps reset from scratch.
func (st *serveState) stepTo(index int) StepResponse {
	limit := index
	if limit >= len(st.scenario.Steps) {
		limit = len(st.scenario.Steps) - 1
	}

	if st.app == nil || index <= st.current {
		// First call or going backward: create fresh app.
		st.app = st.scenario.NewApp(st.scenario.Width, st.scenario.Height)
		st.harness = New(st.app)
		st.current = -1
	}

	// Apply steps from current+1 to limit.
	for i := st.current + 1; i <= limit; i++ {
		if st.scenario.Steps[i].Action != nil {
			st.scenario.Steps[i].Action(st.harness)
		}
	}
	st.current = limit

	if buf := st.harness.Buffer(); buf != nil {
		buf.RenderTo(st.stdout)
	}
	fmt.Fprint(st.stdout, sentinel)

	return StepResponse{
		Name:       st.scenario.Steps[limit].Name,
		StyledText: st.harness.StyledText(),
	}
}

// dispatchInput sends an event to the live app and returns the updated state.
func (st *serveState) dispatchInput(evt *input.Event) StepResponse {
	if evt == nil || st.harness == nil {
		return StepResponse{Name: "input"}
	}

	st.harness.Step(*evt)

	if buf := st.harness.Buffer(); buf != nil {
		buf.RenderTo(st.stdout)
	}
	fmt.Fprint(st.stdout, sentinel)

	return StepResponse{
		Name:       "input",
		StyledText: st.harness.StyledText(),
	}
}

func handleConnectionTo(conn net.Conn, s Scenario, stdout *os.File) {
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	st := newServeState(s, stdout)

	for {
		var req Request
		if err := dec.Decode(&req); err != nil {
			return // connection closed
		}

		switch req.Cmd {
		case "info":
			enc.Encode(buildInfoResponse(s))
		case "step":
			resp := st.stepTo(req.Index)
			enc.Encode(resp)
		case "input":
			resp := st.dispatchInput(req.Event)
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
