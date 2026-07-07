package codegen

import (
	"fmt"
	"strings"
)

// numberedLines renders source with 1-based line numbers for error reports.
func numberedLines(src []byte) string {
	var b strings.Builder
	for i, line := range strings.Split(string(src), "\n") {
		fmt.Fprintf(&b, "%4d\t%s\n", i+1, line)
	}
	return b.String()
}
