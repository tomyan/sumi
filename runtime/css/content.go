package css

import "strings"

// ParseContent evaluates a CSS content: value for a pseudo-element:
// quoted strings, attr(name) against the element's attributes, and
// space-separated concatenation. `none` and empty values report false.
func ParseContent(value string, attrs map[string]string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "none" {
		return "", false
	}
	var b strings.Builder
	rest := value
	for rest != "" {
		rest = strings.TrimLeft(rest, " ")
		if rest == "" {
			break
		}
		switch {
		case rest[0] == '"' || rest[0] == '\'':
			quote := rest[0]
			end := strings.IndexByte(rest[1:], quote)
			if end < 0 {
				return "", false
			}
			b.WriteString(rest[1 : 1+end])
			rest = rest[end+2:]
		case strings.HasPrefix(rest, "attr("):
			end := strings.IndexByte(rest, ')')
			if end < 0 {
				return "", false
			}
			b.WriteString(attrs[strings.TrimSpace(rest[len("attr("):end])])
			rest = rest[end+1:]
		default:
			return "", false
		}
	}
	return b.String(), true
}
