package preview

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func nvimAvailable(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("nvim"); err != nil {
		t.Skip("nvim not found, skipping integration test")
	}
}

func TestEditorStartsAndStops(t *testing.T) {
	nvimAvailable(t)

	// Given — a temporary file to edit
	tmp, err := os.CreateTemp("", "editor-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("hello\n")
	tmp.Close()
	defer os.Remove(tmp.Name())

	// When — create editor
	ed, err := NewEditor(tmp.Name(), 10, 40, nil)
	if err != nil {
		t.Fatalf("NewEditor: %v", err)
	}
	defer ed.Close()

	// Then — screen has correct dimensions
	if ed.Screen().Width() != 40 {
		t.Errorf("Width = %d, want 40", ed.Screen().Width())
	}
	if ed.Screen().Height() != 10 {
		t.Errorf("Height = %d, want 10", ed.Screen().Height())
	}
}

func TestEditorRendersFileContent(t *testing.T) {
	nvimAvailable(t)

	// Given
	tmp, err := os.CreateTemp("", "editor-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.WriteString("Hello World\n")
	tmp.Close()
	defer os.Remove(tmp.Name())

	// When — create editor and wait for nvim to render
	ed, err := NewEditor(tmp.Name(), 10, 40, nil)
	if err != nil {
		t.Fatalf("NewEditor: %v", err)
	}
	defer ed.Close()

	// Then — wait for screen to contain the file content
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		if screenContains(ed, "Hello World") {
			found = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !found {
		t.Error("screen does not contain 'Hello World' after timeout")
	}
}

func TestEditorClose(t *testing.T) {
	nvimAvailable(t)

	// Given
	tmp, err := os.CreateTemp("", "editor-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	ed, err := NewEditor(tmp.Name(), 10, 40, nil)
	if err != nil {
		t.Fatalf("NewEditor: %v", err)
	}

	// When
	ed.Close()

	// Then — process should be terminated
	select {
	case <-ed.done:
		// success
	case <-time.After(3 * time.Second):
		t.Error("editor did not shut down within timeout")
	}
}

func TestEditorResize(t *testing.T) {
	nvimAvailable(t)

	// Given
	tmp, err := os.CreateTemp("", "editor-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	ed, err := NewEditor(tmp.Name(), 10, 40, nil)
	if err != nil {
		t.Fatalf("NewEditor: %v", err)
	}
	defer ed.Close()

	// When
	ed.Resize(20, 80)

	// Then
	if ed.Screen().Width() != 80 {
		t.Errorf("Width = %d, want 80", ed.Screen().Width())
	}
	if ed.Screen().Height() != 20 {
		t.Errorf("Height = %d, want 20", ed.Screen().Height())
	}
}

// screenContains checks if any row of the screen contains the given text.
func screenContains(ed *Editor, text string) bool {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	scr := ed.screen
	for row := 0; row < scr.Height(); row++ {
		var line []rune
		for col := 0; col < scr.Width(); col++ {
			ch := scr.Cell(row, col).Ch
			if ch == 0 {
				ch = ' '
			}
			line = append(line, ch)
		}
		if containsSubstring(string(line), text) {
			return true
		}
	}
	return false
}

func containsSubstring(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && findSubstring(s, sub)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
