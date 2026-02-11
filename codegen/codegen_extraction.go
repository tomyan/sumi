package codegen

import (
	"bytes"
	"fmt"
)

// extractedNode represents a text node extracted from the tree for sync patching.
type extractedNode struct {
	varName  string // e.g., "node0"
	syncExpr string // e.g., `fmt.Sprintf("Count: %v", count)`
}

// extractionCtx tracks expression nodes extracted during tree building.
// When passed to tree-writing functions, text nodes with expressions are
// emitted as named variables rather than inline literals, enabling the
// build-once / sync pattern.
type extractionCtx struct {
	nodes   []extractedNode
	count   int
	prefix  string // namespace prefix, e.g., "counter0_" (empty for root)
	declBuf bytes.Buffer
}

// newExtractionCtx creates an extraction context with the given namespace prefix.
func newExtractionCtx(prefix string) *extractionCtx {
	return &extractionCtx{prefix: prefix}
}

// nextName returns the next unique variable name for an extracted node.
func (e *extractionCtx) nextName() string {
	name := fmt.Sprintf("%snode%d", e.prefix, e.count)
	e.count++
	return name
}

// hasSyncNodes returns true if any nodes were extracted that need syncing.
func (e *extractionCtx) hasSyncNodes() bool {
	return len(e.nodes) > 0
}
