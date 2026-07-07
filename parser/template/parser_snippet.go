package template

import "strings"

// parseSnippetBlock parses {snippet name(params)}...{/snippet}.
func (p *parser) parseSnippetBlock() (Node, error) {
	p.skipWhitespace()

	// Read name
	name := p.readWhile(func(b byte) bool {
		return b != '(' && b != '}' && !isWhitespace(b)
	})

	// Read params (including parentheses)
	params := ""
	if p.pos < len(p.input) && p.input[p.pos] == '(' {
		start := p.pos
		depth := 0
		for p.pos < len(p.input) {
			if p.input[p.pos] == '(' {
				depth++
			}
			if p.input[p.pos] == ')' {
				depth--
				if depth == 0 {
					p.pos++
					params = p.input[start:p.pos]
					break
				}
			}
			p.pos++
		}
	}

	// Consume closing }
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '}' {
		p.pos++
	}

	// Parse children until {/snippet}
	var children []Node
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, p.errorf("missing {/snippet} for snippet %q", name)
		}
		if strings.HasPrefix(p.input[p.pos:], "{/snippet}") {
			p.pos += len("{/snippet}")
			return &SnippetNode{Name: name, Params: params, Children: children}, nil
		}
		if p.input[p.pos] == '{' {
			child, err := p.parseControlFlow()
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		} else if p.input[p.pos] == '<' {
			child, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		} else {
			return nil, p.errorf("unexpected character %q inside {snippet %s} at position %d", p.input[p.pos], name, p.pos)
		}
	}
}

// parseRenderCall parses {render name(args)}.
func (p *parser) parseRenderCall() (Node, error) {
	p.skipWhitespace()

	// Read name
	name := p.readWhile(func(b byte) bool {
		return b != '(' && b != '}' && !isWhitespace(b)
	})

	// Read args
	args := ""
	if p.pos < len(p.input) && p.input[p.pos] == '(' {
		p.pos++ // consume (
		start := p.pos
		depth := 1
		for p.pos < len(p.input) {
			if p.input[p.pos] == '(' {
				depth++
			}
			if p.input[p.pos] == ')' {
				depth--
				if depth == 0 {
					args = p.input[start:p.pos]
					p.pos++ // consume )
					break
				}
			}
			p.pos++
		}
	}

	// Consume closing }
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '}' {
		p.pos++
	}

	return &RenderNode{Name: name, Args: args}, nil
}
