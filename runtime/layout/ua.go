package layout

import "github.com/tomyan/sumi/parser/style"

// uaCSS is the user-agent stylesheet: browser-like defaults for the HTML
// element vocabulary, in cells. Author rules layer on top (they come later
// in the merged rule list, so equal-specificity author declarations win).
const uaCSS = `
h1, h2, h3, h4, h5, h6 { font-weight: bold; margin: 1 0; }
p { margin: 1 0; }
ul, ol { margin: 1 0; padding: 0 0 0 2; }
li::before { content: "• "; }
blockquote { margin: 1 2; padding: 0 0 0 1; opacity: dim; }
pre { white-space: pre; margin: 1 0; }
hr { height: 1; width: 100%; border-top: single; margin: 1 0; }
strong, b { font-weight: bold; }
em, i, var { font-style: italic; }
u, a { text-decoration: underline; }
s, del { text-decoration: line-through; }
mark { inverse: true; }
`

var uaStylesheet = mustParseUA()

func mustParseUA() *style.Stylesheet {
	ss, err := style.Parse(uaCSS)
	if err != nil {
		panic("sumi: UA stylesheet failed to parse: " + err.Error())
	}
	return ss
}

// mergedWithUA caches author stylesheets merged onto the UA layer.
var mergedWithUA = map[*style.Stylesheet]*style.Stylesheet{}

// withUA layers an author stylesheet over the UA defaults. A nil author
// sheet still gets UA styling for HTML elements.
func withUA(author *style.Stylesheet) *style.Stylesheet {
	if author == nil {
		return uaStylesheet
	}
	if merged, ok := mergedWithUA[author]; ok {
		return merged
	}
	merged := &style.Stylesheet{
		Rules:     append(append([]style.Rule{}, uaStylesheet.Rules...), author.Rules...),
		Keyframes: author.Keyframes,
	}
	mergedWithUA[author] = merged
	return merged
}
