package render

import (
	"encoding/base64"
	"fmt"
	"io"
)

// CopyToClipboard writes an OSC 52 escape sequence to copy text to the system clipboard.
func CopyToClipboard(w io.Writer, text string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	fmt.Fprintf(w, "\x1b]52;c;%s\x07", encoded)
}
