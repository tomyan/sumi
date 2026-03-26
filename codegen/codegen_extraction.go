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
// build-once / sync pattern. Dynamic-children parents (boxes with {if}/{for})
// have their Children rebuilt in sync.
type extractionCtx struct {
	nodes          []extractedNode // text nodes needing Content sync
	count          int
	boxCount       int
	prefix         string // namespace prefix, e.g., "counter0_" (empty for root)
	hasCursor      bool   // true when any box has dynamic cursor attributes
	extractRootBox bool   // true when the first box should be extracted (for $self wiring)
	focusablesSeen int    // counts focusable boxes seen during tree walk (for cursor-focus correlation)
	signals        map[string]bool // signal variable names (for auto-unwrapping in expressions)
	declBuf        bytes.Buffer
	preSyncBuf     bytes.Buffer // code that must run before node content updates (e.g., focused state)
	syncBuf        bytes.Buffer // multi-line sync code (for dynamic children IIFE)
}

// newExtractionCtx creates an extraction context with the given namespace prefix.
func newExtractionCtx(prefix string) *extractionCtx {
	return &extractionCtx{prefix: prefix}
}

// nextNodeName returns the next unique variable name for an extracted text node.
func (e *extractionCtx) nextNodeName() string {
	name := fmt.Sprintf("%snode%d", e.prefix, e.count)
	e.count++
	return name
}

// nextBoxName returns the next unique variable name for an extracted box.
func (e *extractionCtx) nextBoxName() string {
	name := fmt.Sprintf("%sbox%d", e.prefix, e.boxCount)
	e.boxCount++
	return name
}

// hasSyncContent returns true if any sync entries exist.
func (e *extractionCtx) hasSyncContent() bool {
	return len(e.nodes) > 0 || e.preSyncBuf.Len() > 0 || e.syncBuf.Len() > 0
}

// firstBoxName returns the name of the first extracted box, or "" if none.
func (e *extractionCtx) firstBoxName() string {
	if e.boxCount == 0 {
		return ""
	}
	return fmt.Sprintf("%sbox0", e.prefix)
}
