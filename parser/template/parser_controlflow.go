package template

import (
	"fmt"
	"strings"
)

// parseControlFlow dispatches {if ...} and {for ...} blocks.
func (p *parser) parseControlFlow() (Node, error) {
	p.pos++ // consume '{'
	p.skipWhitespace()
	keyword := p.readWhile(func(b byte) bool {
		return b != ' ' && b != '\t' && b != '}'
	})
	switch keyword {
	case "if":
		return p.parseIfNode()
	case "for":
		return p.parseForNode()
	case "slot":
		return p.parseSlotDefBlock()
	case "snippet":
		return p.parseSnippetBlock()
	case "render":
		return p.parseRenderCall()
	default:
		return nil, fmt.Errorf("unexpected control flow keyword %q at position %d", keyword, p.pos)
	}
}

// parseIfNode parses the condition and body of an {if ...}...{/if} block.
func (p *parser) parseIfNode() (Node, error) {
	p.skipWhitespace()
	condition := strings.TrimSpace(p.readUntil('}'))
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unterminated {if} tag")
	}
	p.pos++ // consume '}'

	then, hitElse, err := p.parseControlFlowChildren("if")
	if err != nil {
		return nil, err
	}

	var elseChildren []Node
	if hitElse {
		elseChildren, _, err = p.parseControlFlowChildren("if")
		if err != nil {
			return nil, err
		}
	}

	return &IfNode{
		Condition: condition,
		Then:      then,
		Else:      elseChildren,
	}, nil
}

// parseForNode parses the clause and body of a {for ...}...{/for} block.
func (p *parser) parseForNode() (Node, error) {
	p.skipWhitespace()
	raw := strings.TrimSpace(p.readUntil('}'))
	if p.pos >= len(p.input) {
		return nil, fmt.Errorf("unterminated {for} tag")
	}
	p.pos++ // consume '}'

	clause, key := splitForKey(raw)

	children, _, err := p.parseControlFlowChildren("for")
	if err != nil {
		return nil, err
	}

	return &ForNode{
		Clause:   clause,
		Key:      key,
		Children: children,
	}, nil
}

// splitForKey splits a for clause on the last " key=" to separate the Go
// for-clause from the key expression. Uses the last occurrence to avoid
// conflicts with variable names containing "key".
func splitForKey(raw string) (clause, key string) {
	idx := strings.LastIndex(raw, " key=")
	if idx < 0 {
		return raw, ""
	}
	return strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+5:])
}

// parseControlFlowChildren parses children until {/keyword} or {else} is found.
// Returns the children, whether {else} was encountered, and any error.
func (p *parser) parseControlFlowChildren(keyword string) ([]Node, bool, error) {
	closeTag := "{/" + keyword + "}"
	elseTag := "{else}"
	var children []Node

	for {
		if p.pos >= len(p.input) {
			return nil, false, fmt.Errorf("missing closing {/%s}", keyword)
		}
		if strings.HasPrefix(p.input[p.pos:], closeTag) {
			p.pos += len(closeTag)
			return children, false, nil
		}
		if keyword == "if" && strings.HasPrefix(p.input[p.pos:], elseTag) {
			p.pos += len(elseTag)
			return children, true, nil
		}
		if p.input[p.pos] == '{' && p.controlFlowStart() {
			child, err := p.parseControlFlow()
			if err != nil {
				return nil, false, err
			}
			children = append(children, child)
			continue
		}
		if p.input[p.pos] == '<' {
			child, err := p.parseElement()
			if err != nil {
				return nil, false, err
			}
			children = append(children, child)
			continue
		}
		if node := p.parseLooseText(); node != nil {
			children = append(children, node)
		}
	}
}
