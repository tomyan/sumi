package css

import (
	"strings"

	"github.com/tomyan/sumi/parser/style"
)

// matchStructural evaluates one structural pseudo-class against an element.
// Unknown pseudo-classes never match (graceful-drop: the rule goes inert
// rather than erroring).
func matchStructural(raw string, e Element) bool {
	name, arg := splitPseudoArg(raw)
	idx, count := siblingPosition(e)
	switch name {
	case "root":
		return e.Tag == "root"
	case "empty":
		return e.Empty
	case "first-child":
		return idx == 0
	case "last-child":
		return idx == count-1
	case "only-child":
		return count == 1
	case "nth-child":
		return nthMatches(arg, idx+1)
	case "nth-last-child":
		return nthMatches(arg, count-idx)
	}
	tIdx, tCount := typePosition(e)
	switch name {
	case "first-of-type":
		return tIdx == 0
	case "last-of-type":
		return tIdx == tCount-1
	case "only-of-type":
		return tCount == 1
	case "nth-of-type":
		return nthMatches(arg, tIdx+1)
	case "nth-last-of-type":
		return nthMatches(arg, tCount-tIdx)
	}
	return false
}

// splitPseudoArg splits "nth-child(2n+1)" into ("nth-child", "2n+1").
func splitPseudoArg(raw string) (string, string) {
	open := strings.IndexByte(raw, '(')
	if open < 0 {
		return raw, ""
	}
	close := strings.LastIndexByte(raw, ')')
	if close < open {
		return raw, ""
	}
	return raw[:open], raw[open+1 : close]
}

func nthMatches(arg string, index int) bool {
	nth, err := style.ParseNth(arg)
	if err != nil {
		return false
	}
	return nth.Matches(index)
}

// siblingPosition returns the element's 0-based index and the sibling count.
// An element without sibling context counts as an only child.
func siblingPosition(e Element) (int, int) {
	if e.Siblings == nil {
		return 0, 1
	}
	return e.Index, len(e.Siblings)
}

// typePosition returns the element's 0-based index and count among siblings
// with the same tag.
func typePosition(e Element) (int, int) {
	if e.Siblings == nil {
		return 0, 1
	}
	idx, count := 0, 0
	for i, sib := range e.Siblings {
		if sib.Tag != e.Tag {
			continue
		}
		if i == e.Index {
			idx = count
		}
		count++
	}
	return idx, count
}
