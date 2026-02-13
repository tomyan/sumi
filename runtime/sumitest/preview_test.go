package sumitest

import (
	"bytes"
	"strings"
	"testing"
)

func TestPreviewToWriterRendersHeader(t *testing.T) {
	// Given
	s := counterScenario()
	frames := RunScenario(s)

	// When
	var out bytes.Buffer
	writePreviewFrame(&out, s, frames, 0)

	// Then — header should contain scenario name, frame number, and step name
	result := out.String()
	if !strings.Contains(result, "counter-basics") {
		t.Errorf("expected scenario name in header, got:\n%s", result)
	}
	if !strings.Contains(result, "1/3") {
		t.Errorf("expected frame number in header, got:\n%s", result)
	}
	if !strings.Contains(result, "initial") {
		t.Errorf("expected step name in header, got:\n%s", result)
	}
}

func TestPreviewToWriterRendersMultipleFrames(t *testing.T) {
	// Given
	s := counterScenario()
	frames := RunScenario(s)

	// When — render each frame
	for i := range frames {
		var out bytes.Buffer
		writePreviewFrame(&out, s, frames, i)
		result := out.String()

		// Then — should contain ANSI escape sequences (cursor addressing)
		if !strings.Contains(result, "\x1b[") {
			t.Errorf("frame %d: expected ANSI sequences in output", i)
		}
	}
}

func TestPreviewToWriterContainsClearScreen(t *testing.T) {
	// Given
	s := counterScenario()
	frames := RunScenario(s)

	// When
	var out bytes.Buffer
	writePreviewFrame(&out, s, frames, 0)

	// Then — should start with clear screen sequence
	result := out.String()
	if !strings.Contains(result, "\x1b[2J") {
		t.Errorf("expected clear screen in output, got:\n%q", result[:min(len(result), 100)])
	}
}

func TestHarnessBufferReturnsBuffer(t *testing.T) {
	// Given
	app := createCounterApp(20, 3)
	h := New(app)

	// When
	buf := h.Buffer()

	// Then
	if buf == nil {
		t.Fatal("Buffer() returned nil")
	}
	if buf.Width() != 20 {
		t.Errorf("expected width 20, got %d", buf.Width())
	}
	if buf.Height() != 3 {
		t.Errorf("expected height 3, got %d", buf.Height())
	}
}
