package template

import "strings"

// controlFlowTokens are the brace-openers that end a loose text run:
// control-flow blocks, block terminators, and {else}. Any other brace
// starts an {expression} that belongs to the text.
var controlFlowTokens = []string{
	"{if ", "{for ", "{render ", "{snippet ", "{slot", "{/", "{else}",
}

// controlFlowStart reports whether the input at pos begins a
// control-flow token rather than a text expression.
func (p *parser) controlFlowStart() bool {
	rest := p.input[p.pos:]
	for _, tok := range controlFlowTokens {
		if strings.HasPrefix(rest, tok) {
			return true
		}
	}
	return false
}

// parseLooseText consumes a run of loose text (including {expr} parts)
// inside a container body, stopping at the next element tag or
// control-flow token. Whitespace-only runs follow the JSX newline rule:
// a gap containing a newline is source formatting and yields nil; a
// single-line gap yields one space text node.
func (p *parser) parseLooseText() Node {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == '<' {
			break
		}
		if c == '{' {
			if p.controlFlowStart() {
				break
			}
			p.consumeExpression()
			continue
		}
		p.pos++
	}
	return looseTextNode(p.input[start:p.pos])
}

// consumeExpression advances past a brace-delimited {expression}.
// Braces do not nest (matching parseTextParts).
func (p *parser) consumeExpression() {
	for p.pos < len(p.input) && p.input[p.pos] != '}' {
		p.pos++
	}
	if p.pos < len(p.input) {
		p.pos++ // consume '}'
	}
}

// looseTextNode builds the text node for a raw loose-text run, applying
// the JSX newline rule to whitespace-only runs.
func looseTextNode(raw string) Node {
	if strings.TrimSpace(raw) != "" {
		return &TextElement{Attributes: map[string]string{}, Parts: parseTextParts(raw)}
	}
	if strings.Contains(raw, "\n") || raw == "" {
		return nil // formatting gap
	}
	return &TextElement{Attributes: map[string]string{}, Parts: []Part{&StringPart{Value: " "}}}
}
