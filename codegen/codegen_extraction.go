package codegen

import (
	"bytes"
	"fmt"
)

// extractedNode represents a text node extracted from the tree for sync patching.
type extractedNode struct {
	varName  string // e.g., "node0"
	syncExpr string // e.g., `sumi.Sprintf("Count: %v", count)`
}

// extractionCtx tracks expression nodes extracted during tree building.
// When passed to tree-writing functions, text nodes with expressions are
// emitted as named variables rather than inline literals, enabling the
// build-once / sync pattern. Dynamic-children parents (boxes with {if}/{for})
// have their Children rebuilt in sync.
type extractionCtx struct {
	nodes             []extractedNode              // text nodes needing Content sync
	count             int
	boxCount          int
	signals           map[string]bool              // signal variable names (for auto-unwrapping in expressions)
	eventFuncs        map[string]bool              // declared funcs with params — emitted as direct DOM handler refs
	componentChildren map[string]ComponentChildInfo // child components for template
	declBuf           bytes.Buffer
	syncBuf           bytes.Buffer // dynamic children rebuild code (for {if}/{for} inside sumi.Effect)
	inDynamic         bool         // true inside IIFE — skip extraction, only do signal unwrapping
	componentIdx      int          // counter for child component variable names (matches writeChildComponentInstances order)
}

// newExtractionCtx creates an extraction context.
func newExtractionCtx() *extractionCtx {
	return &extractionCtx{}
}

// nextNodeName returns the next unique variable name for an extracted text node.
func (e *extractionCtx) nextNodeName() string {
	name := fmt.Sprintf("node%d", e.count)
	e.count++
	return name
}

// nextBoxName returns the next unique variable name for an extracted box.
func (e *extractionCtx) nextBoxName() string {
	name := fmt.Sprintf("box%d", e.boxCount)
	e.boxCount++
	return name
}

// hasSyncContent returns true if any sync entries exist.
func (e *extractionCtx) hasSyncContent() bool {
	return len(e.nodes) > 0 || e.syncBuf.Len() > 0
}

