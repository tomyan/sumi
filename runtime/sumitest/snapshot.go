package sumitest

import (
	"fmt"
	"os"
	"strings"
)

const frameSeparatorPrefix = "=== Frame: "
const frameSeparatorSuffix = " ==="

// WriteSnapshot writes frames to a snapshot file.
// Format: each frame has a separator line followed by styled text and a blank line.
func WriteSnapshot(path string, frames []Frame) error {
	var b strings.Builder
	for _, f := range frames {
		fmt.Fprintf(&b, "%s%s%s\n", frameSeparatorPrefix, f.Name, frameSeparatorSuffix)
		b.WriteString(f.StyledText)
		b.WriteByte('\n')
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

// ReadSnapshot reads frames from a snapshot file.
func ReadSnapshot(path string) ([]Frame, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseSnapshot(string(data))
}

// parseSnapshot parses the snapshot file format into frames.
func parseSnapshot(content string) ([]Frame, error) {
	var frames []Frame
	lines := strings.Split(content, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		if !strings.HasPrefix(line, frameSeparatorPrefix) {
			i++
			continue
		}
		name := strings.TrimPrefix(line, frameSeparatorPrefix)
		name = strings.TrimSuffix(name, frameSeparatorSuffix)
		i++

		// Collect text lines until next separator or end of file
		var textLines []string
		for i < len(lines) {
			if strings.HasPrefix(lines[i], frameSeparatorPrefix) {
				break
			}
			textLines = append(textLines, lines[i])
			i++
		}

		// Trim trailing empty lines (the blank line separator)
		for len(textLines) > 0 && textLines[len(textLines)-1] == "" {
			textLines = textLines[:len(textLines)-1]
		}

		frames = append(frames, Frame{
			Name:       name,
			StyledText: strings.Join(textLines, "\n"),
		})
	}
	return frames, nil
}
