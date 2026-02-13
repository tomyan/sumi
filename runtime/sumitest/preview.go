package sumitest

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// Preview runs a scenario interactively, showing each frame with ANSI rendering.
// Press Enter to advance between frames. Only active when -preview flag is set.
func Preview(s Scenario) {
	frames := RunScenario(s)
	reader := bufio.NewReader(os.Stdin)
	for i := range frames {
		writePreviewFrame(os.Stdout, s, frames, i)
		fmt.Print("\n\nPress Enter to continue...")
		reader.ReadString('\n')
	}
}

const headerHeight = 3

// writePreviewFrame renders a single preview frame to a writer.
func writePreviewFrame(w io.Writer, s Scenario, frames []Frame, index int) {
	frame := frames[index]

	// Clear screen and move to top-left
	fmt.Fprint(w, "\x1b[2J\x1b[1;1H")

	// Header
	fmt.Fprintf(w, "Scenario: %s  |  Frame %d/%d  |  Step: %s\n",
		s.Name, index+1, len(frames), frame.Name)
	fmt.Fprintf(w, "%s\n", repeatChar('─', 60))

	// Replay scenario to this frame to get the buffer
	h := replayToFrame(s, index)
	if buf := h.Buffer(); buf != nil {
		buf.RenderToOffset(w, headerHeight-1, 0)
	}
}

// replayToFrame creates a fresh app and replays steps up to the given index.
func replayToFrame(s Scenario, index int) *Harness {
	app := s.NewApp(s.Width, s.Height)
	h := New(app)
	for i := 0; i <= index; i++ {
		if s.Steps[i].Action != nil {
			s.Steps[i].Action(h)
		}
	}
	return h
}

func repeatChar(ch rune, count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = ch
	}
	return string(result)
}
