package template

import (
	"fmt"
	"strings"
)

// parseSlotElement parses <slot:name /> or <slot:name>default content</slot:name>.
func (p *parser) parseSlotElement(tagName string) (Node, error) {
	slotName := strings.TrimPrefix(tagName, "slot:")
	attrs, err := p.parseAttributes()
	if err != nil {
		return nil, err
	}

	// Self-closing: <slot:name />
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		p.pos++ // consume /
		if p.pos < len(p.input) && p.input[p.pos] == '>' {
			p.pos++ // consume >
		}
		return &SlotElement{Name: slotName, Attributes: attrs}, nil
	}

	// Consume >
	if p.pos < len(p.input) && p.input[p.pos] == '>' {
		p.pos++
	}

	// Parse default content using the existing child parsing infrastructure.
	closingTag := fmt.Sprintf("</slot:%s>", slotName)
	var children []Node
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, p.errorf("missing closing %s", closingTag)
		}
		if strings.HasPrefix(p.input[p.pos:], closingTag) {
			p.pos += len(closingTag)
			return &SlotElement{Name: slotName, Attributes: attrs, Default: children}, nil
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
			return nil, p.errorf("unexpected character %q inside <slot:%s> at position %d", p.input[p.pos], slotName, p.pos)
		}
	}
}

// parseSlotDefBlock parses {slot name}...{/slot} content definition.
// Called from parseControlFlow after the "slot" keyword is consumed.
func (p *parser) parseSlotDefBlock() (Node, error) {
	p.skipWhitespace()

	// Read name and optional params until }
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '}' {
		p.pos++
	}
	nameAndParams := strings.TrimSpace(p.input[start:p.pos])
	if p.pos < len(p.input) {
		p.pos++ // consume }
	}

	name := nameAndParams
	params := ""
	if idx := strings.Index(name, "("); idx >= 0 {
		params = name[idx:]
		name = strings.TrimSpace(name[:idx])
	}

	// Parse children until {/slot}
	var children []Node
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, p.errorf("missing {/slot} for slot %q", name)
		}
		if strings.HasPrefix(p.input[p.pos:], "{/slot}") {
			p.pos += len("{/slot}")
			return &SlotDefNode{Name: name, Params: params, Children: children}, nil
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
			return nil, p.errorf("unexpected character %q inside {slot %s} at position %d", p.input[p.pos], name, p.pos)
		}
	}
}
