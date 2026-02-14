package sumitest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Preview runs a scenario interactively, showing each frame with ANSI rendering
// and the styled text snapshot below it. Opens /dev/tty directly so output and
// input work even under go test's pipes. Press Enter to advance between frames.
func Preview(s Scenario) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "preview: cannot open /dev/tty: %v\n", err)
		return
	}
	defer tty.Close()

	frames := RunScenario(s)
	reader := bufio.NewReader(tty)
	for i := range frames {
		writePreviewFrame(tty, s, frames, i)
		promptRow := previewHeight(s, frames[i])
		fmt.Fprintf(tty, "\x1b[%d;1H", promptRow)
		if i < len(frames)-1 {
			fmt.Fprint(tty, "Press Enter for next frame...")
		} else {
			fmt.Fprint(tty, "Last frame. Press Enter to finish.")
		}
		reader.ReadString('\n')
	}
	// Clear screen on exit
	fmt.Fprint(tty, "\x1b[2J\x1b[1;1H")
}

const headerHeight = 3

// previewHeight returns the total rows needed for a preview frame.
func previewHeight(s Scenario, frame Frame) int {
	snapshotLines := strings.Count(frame.StyledText, "\n") + 1
	// header + rendered component + separator + "Snapshot:" label + snapshot lines + blank
	h := headerHeight + s.Height + 1 + 1 + snapshotLines + 1
	if s.SourceFile != "" {
		h += 1 + sourceLineCount(s.SourceFile) + 1 // separator + source + blank
	}
	return h
}

// writePreviewFrame renders a single preview frame to a writer.
// Shows the ANSI-rendered component followed by the styled text snapshot.
func writePreviewFrame(w io.Writer, s Scenario, frames []Frame, index int) {
	frame := frames[index]

	// Clear screen and move to top-left
	fmt.Fprint(w, "\x1b[2J\x1b[1;1H")

	// Header
	fmt.Fprintf(w, "Scenario: %s  |  Frame %d/%d  |  Step: %s\n",
		s.Name, index+1, len(frames), frame.Name)
	fmt.Fprintf(w, "%s\n", repeatChar('─', 60))

	// Rendered component
	h := replayToFrame(s, index)
	if buf := h.Buffer(); buf != nil {
		buf.RenderToOffset(w, headerHeight-1, 0)
	}

	// Snapshot section below the rendered component
	snapshotRow := headerHeight + s.Height
	fmt.Fprintf(w, "\x1b[%d;1H", snapshotRow)
	fmt.Fprintf(w, "\x1b[0m%s\n", repeatChar('─', 60))
	fmt.Fprintf(w, "Snapshot:\n")
	for _, line := range strings.Split(frame.StyledText, "\n") {
		fmt.Fprintf(w, "  %s\n", line)
	}

	// Source code section
	if s.SourceFile != "" {
		writeSourceSection(w, s.SourceFile)
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

// writeSourceSection reads and displays a .sumi source file.
func writeSourceSection(w io.Writer, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(w, "\x1b[2mSource: %v\x1b[0m\n", err)
		return
	}
	fmt.Fprintf(w, "%s\n", repeatChar('─', 60))
	fmt.Fprintf(w, "Source: %s\n", path)
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		fmt.Fprintf(w, "\x1b[2m  %s\x1b[0m\n", line)
	}
}

// sourceLineCount returns the number of lines in a file, or 0 on error.
func sourceLineCount(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return strings.Count(string(data), "\n") + 1
}

func repeatChar(ch rune, count int) string {
	result := make([]rune, count)
	for i := range result {
		result[i] = ch
	}
	return string(result)
}
