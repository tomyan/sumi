package template

import (
	"strings"
	"unicode"
)

// TagTier classifies a tag name into one of the four component tiers.
type TagTier int

const (
	TierPrimitive   TagTier = iota // box, text, title
	TierFundamental                // textedit, scrollbar — lowercase, no prefix
	TierStdlib                     // sumi:TextInput — sumi: prefix
	TierUser                       // myui:Button — other prefix
)

// primitives is the set of parser-level element names.
var primitives = map[string]bool{
	"box":   true,
	"text":  true,
	"title": true,
}

// ClassifyTag returns the tier for a given tag name.
func ClassifyTag(name string) TagTier {
	if primitives[name] {
		return TierPrimitive
	}
	if prefix, _, ok := SplitPrefix(name); ok {
		if prefix == "sumi" {
			return TierStdlib
		}
		return TierUser
	}
	return TierFundamental
}

// TagRegistryKey returns the lowercase registry key for a tag name.
// Fundamental tags pass through unchanged. Prefixed tags are lowercased.
func TagRegistryKey(name string) string {
	if prefix, local, ok := SplitPrefix(name); ok {
		return prefix + ":" + strings.ToLower(local)
	}
	return name
}

// TagComponentFile returns the .sumi filename for a tag name.
// Fundamental tags: "textedit" → "textedit.sumi".
// Prefixed tags: "sumi:TextInput" → "text-input.sumi" (PascalCase to kebab-case).
func TagComponentFile(name string) string {
	if _, local, ok := SplitPrefix(name); ok {
		return pascalToKebab(local) + ".sumi"
	}
	return name + ".sumi"
}

// SplitPrefix splits "prefix:Local" into ("prefix", "Local", true).
// Returns ("", "", false) if there is no colon.
func SplitPrefix(name string) (prefix, local string, ok bool) {
	idx := strings.IndexByte(name, ':')
	if idx < 0 {
		return "", "", false
	}
	return name[:idx], name[idx+1:], true
}

// pascalToKebab converts PascalCase to kebab-case.
// "TextInput" → "text-input", "Select" → "select".
func pascalToKebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
