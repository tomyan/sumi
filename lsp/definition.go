package lsp

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tomyan/sumi/parser/section"
	"github.com/tomyan/sumi/parser/template"
)

// Definition resolves a go-to-definition request: a component tag jumps to
// the component's file; a handler name inside an on* attribute jumps to its
// function declaration. It returns nil when there is nothing to resolve.
func Definition(text string, pos Position, uri string) *Location {
	offset := positionToOffset(text, pos)
	switch ClassifyContext(text, pos) {
	case ContextTagName:
		return componentDefinition(wordAt(text, offset), uri)
	case ContextAttrName:
		return handlerDefinition(text, offset, uri)
	default:
		return nil
	}
}

// componentDefinition locates the .sumi file for a component tag. HTML tag
// names and unknown components resolve to nil.
func componentDefinition(tag, uri string) *Location {
	tag = strings.TrimPrefix(tag, "/")
	if tag == "" || isKnownHTMLTag(tag) {
		return nil
	}
	dir := filepath.Dir(uriToPath(uri))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sumi") {
			continue
		}
		if template.ExportedComponentName(template.ComponentName(e.Name())) == tag {
			return fileStart("file://" + filepath.Join(dir, e.Name()))
		}
	}
	return nil
}

// isKnownHTMLTag reports whether tag is a recognised HTML element name.
func isKnownHTMLTag(tag string) bool {
	for _, name := range template.HTMLTagNames() {
		if name == tag {
			return true
		}
	}
	return false
}

// handlerDefinition locates the function referenced by an on* attribute value
// under the cursor.
func handlerDefinition(text string, offset int, uri string) *Location {
	name, ok := handlerRef(text, offset)
	if !ok {
		return nil
	}
	sections, err := section.Parse(text)
	if err != nil || sections.ScriptStart < 0 {
		return nil
	}
	re := regexp.MustCompile(`func\s+(` + regexp.QuoteMeta(name) + `)\b`)
	loc := re.FindStringSubmatchIndex(sections.Script)
	if loc == nil {
		return nil
	}
	start := offsetToPosition(text, sections.ScriptStart+loc[2])
	return &Location{URI: uri, Range: Range{Start: start, End: start}}
}

// handlerRef returns the identifier under the cursor when it is the value of
// an on* attribute expression: on…={name}.
func handlerRef(text string, offset int) (string, bool) {
	name := wordAt(text, offset)
	if name == "" {
		return "", false
	}
	start := offset
	for start > 0 && isWordByte(text[start-1]) {
		start--
	}
	i := start - 1
	if i < 0 || text[i] != '{' {
		return "", false
	}
	i--
	if i < 0 || text[i] != '=' {
		return "", false
	}
	if attr := attrNameBefore(text, i); strings.HasPrefix(attr, "on") && len(attr) > 2 {
		return name, true
	}
	return "", false
}

// attrNameBefore reads the attribute-name run ending just before index i.
func attrNameBefore(text string, i int) string {
	j := i
	for j > 0 && isAttrNameByte(text[j-1]) {
		j--
	}
	return text[j:i]
}

// isAttrNameByte reports whether b can appear in an attribute name.
func isAttrNameByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z':
		return true
	case b == '-', b == ':':
		return true
	}
	return false
}

// fileStart returns a location at the start (0:0) of a file URI.
func fileStart(uri string) *Location {
	return &Location{URI: uri, Range: Range{}}
}
