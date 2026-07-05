package style

import (
	"fmt"
	"strings"
)

// Stylesheet represents the parsed contents of a <style> block.
type Stylesheet struct {
	Rules     []Rule
	Keyframes []Keyframe
}

// Rule represents a single CSS-like rule with one selector and properties.
// Selector lists (`a, b { }`) are exploded into one Rule per selector.
type Rule struct {
	Selector   string            // raw selector text without the subject pseudo
	Pseudo     string            // subject pseudo-class ("hover", ...); "" for none
	Parsed     ComplexSelector   // structured form used for matching
	Media      string            // enclosing @media query text; "" for none
	Properties map[string]string // e.g. {"color": "green", "font-weight": "bold"}
}

// Keyframe represents an @keyframes animation definition.
type Keyframe struct {
	Name  string         // animation name
	Stops []KeyframeStop // percentage stops with properties
}

// KeyframeStop is a single stop in a keyframe animation.
type KeyframeStop struct {
	Percent    float64 // 0.0 to 1.0
	Properties map[string]string
}

// Parse parses CSS-like style rules from the content of a <style> block.
func Parse(input string) (*Stylesheet, error) {
	p := &parser{input: input, pos: 0}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() (*Stylesheet, error) {
	s := &Stylesheet{}

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}

		// Check for comment that wasn't fully consumed (unterminated)
		if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '*' {
			return nil, fmt.Errorf("unterminated comment")
		}

		// @keyframes block
		if strings.HasPrefix(p.input[p.pos:], "@keyframes ") {
			kf, err := p.parseKeyframes()
			if err != nil {
				return nil, err
			}
			s.Keyframes = append(s.Keyframes, kf)
			continue
		}

		// @media blocks: rules parse normally, tagged with the query.
		if strings.HasPrefix(p.input[p.pos:], "@media") {
			rules, err := p.parseMediaBlock()
			if err != nil {
				return nil, err
			}
			s.Rules = append(s.Rules, rules...)
			continue
		}

		// Unknown at-rules parse and drop silently (graceful-drop policy).
		if p.input[p.pos] == '@' {
			if err := p.skipAtRule(); err != nil {
				return nil, err
			}
			continue
		}

		rules, err := p.parseRules()
		if err != nil {
			return nil, err
		}
		s.Rules = append(s.Rules, rules...)
	}

	return s, nil
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.pos++
	}
}

func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos+1 < len(p.input) && p.input[p.pos] == '/' && p.input[p.pos+1] == '*' {
			if err := p.skipComment(); err != nil {
				return // leave position at the comment start for error reporting
			}
		} else {
			return
		}
	}
}

func (p *parser) skipComment() error {
	p.pos += 2 // skip /*
	for p.pos+1 < len(p.input) {
		if p.input[p.pos] == '*' && p.input[p.pos+1] == '/' {
			p.pos += 2
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated comment")
}

// skipAtRule consumes an unrecognised at-rule: either a statement form
// terminated by ';' (@import "x";) or a block form with balanced braces
// (@media ... { ... }). The content is discarded.
func (p *parser) skipAtRule() error {
	start := p.pos
	// Scan the prelude up to '{' or ';'.
	for p.pos < len(p.input) && p.input[p.pos] != '{' && p.input[p.pos] != ';' {
		p.pos++
	}
	if p.pos >= len(p.input) {
		return fmt.Errorf("unterminated at-rule at position %d", start)
	}
	if p.input[p.pos] == ';' {
		p.pos++ // statement form consumed
		return nil
	}
	// Block form: skip balanced braces.
	depth := 0
	for p.pos < len(p.input) {
		switch p.input[p.pos] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				p.pos++
				return nil
			}
		}
		p.pos++
	}
	return fmt.Errorf("unterminated at-rule block at position %d", start)
}

// parseMediaBlock parses `@media <query> { rules... }`, tagging every
// contained rule with the query text.
func (p *parser) parseMediaBlock() ([]Rule, error) {
	p.pos += len("@media")
	query := strings.TrimSpace(p.readSelectorText())
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, fmt.Errorf("@media %s: expected '{'", query)
	}
	p.pos++ // consume outer {

	var rules []Rule
	for {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("@media %s: unterminated block", query)
		}
		if p.input[p.pos] == '}' {
			p.pos++ // consume outer }
			return rules, nil
		}
		inner, err := p.parseRules()
		if err != nil {
			return nil, err
		}
		for i := range inner {
			inner[i].Media = query
		}
		rules = append(rules, inner...)
	}
}

// parseRules parses one rule block; a selector list yields one Rule per
// selector, all sharing the block's properties.
func (p *parser) parseRules() ([]Rule, error) {
	selectorText := p.readSelectorText()
	if strings.TrimSpace(selectorText) == "" {
		return nil, fmt.Errorf("expected selector at position %d", p.pos)
	}

	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return nil, fmt.Errorf("expected '{' after selector %q", selectorText)
	}
	p.pos++ // skip {

	props, err := p.parseProperties()
	if err != nil {
		return nil, err
	}

	selectors, err := ParseSelectorList(selectorText)
	if err != nil {
		return nil, err
	}

	items := SplitSelectorList(selectorText)
	rules := make([]Rule, len(selectors))
	for i, sel := range selectors {
		pseudo := sel.Parts[len(sel.Parts)-1].Pseudo
		raw := strings.TrimSpace(items[i])
		if pseudo != "" {
			raw = strings.TrimSuffix(raw, ":"+pseudo)
		}
		rules[i] = Rule{Selector: raw, Pseudo: pseudo, Parsed: sel, Properties: props}
	}
	return rules, nil
}

// readSelectorText reads raw selector text up to the opening brace.
func (p *parser) readSelectorText() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '{' {
		p.pos++
	}
	return strings.TrimSpace(p.input[start:p.pos])
}

func (p *parser) parseProperties() (map[string]string, error) {
	props := make(map[string]string)

	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unterminated rule block, expected '}'")
		}

		// Check for closing brace
		if p.input[p.pos] == '}' {
			p.pos++ // skip }
			return props, nil
		}

		// Parse property name
		name := p.readPropertyName()
		if name == "" {
			return nil, fmt.Errorf("expected property name at position %d", p.pos)
		}

		p.skipWhitespace()

		// Expect colon
		if p.pos >= len(p.input) || p.input[p.pos] != ':' {
			return nil, fmt.Errorf("expected ':' after property %q", name)
		}
		p.pos++ // skip :

		p.skipWhitespace()

		// Read value (up to ; or })
		value := p.readPropertyValue()

		props[name] = value
	}

	return nil, fmt.Errorf("unterminated rule block, expected '}'")
}

func (p *parser) readPropertyName() string {
	start := p.pos
	for p.pos < len(p.input) && isPropertyNameChar(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) readPropertyValue() string {
	var b strings.Builder
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if ch == ';' {
			p.pos++ // consume semicolon
			break
		}
		if ch == '}' {
			// Don't consume }, leave it for the caller
			break
		}
		b.WriteByte(ch)
		p.pos++
	}
	return strings.TrimSpace(b.String())
}

// parseKeyframes parses an @keyframes block:
// @keyframes name { 0% { color: red; } 100% { color: blue; } }
func (p *parser) parseKeyframes() (Keyframe, error) {
	// Skip "@keyframes "
	p.pos += len("@keyframes ")
	p.skipWhitespace()

	// Read name
	start := p.pos
	for p.pos < len(p.input) && !isWhitespace(p.input[p.pos]) && p.input[p.pos] != '{' {
		p.pos++
	}
	name := strings.TrimSpace(p.input[start:p.pos])
	if name == "" {
		return Keyframe{}, fmt.Errorf("@keyframes missing name")
	}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return Keyframe{}, fmt.Errorf("@keyframes %s: expected '{'", name)
	}
	p.pos++ // consume outer {

	var stops []KeyframeStop
	for p.pos < len(p.input) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++ // consume outer }
			return Keyframe{Name: name, Stops: stops}, nil
		}

		stop, err := p.parseKeyframeStop()
		if err != nil {
			return Keyframe{}, fmt.Errorf("@keyframes %s: %w", name, err)
		}
		stops = append(stops, stop)
	}
	return Keyframe{}, fmt.Errorf("@keyframes %s: unterminated block", name)
}

func (p *parser) parseKeyframeStop() (KeyframeStop, error) {
	p.skipWhitespace()

	// Read the selector: "0%", "50%", "100%", "from", "to"
	start := p.pos
	for p.pos < len(p.input) && !isWhitespace(p.input[p.pos]) && p.input[p.pos] != '{' {
		p.pos++
	}
	selector := strings.TrimSpace(p.input[start:p.pos])

	var percent float64
	switch selector {
	case "from":
		percent = 0
	case "to":
		percent = 1
	default:
		if !strings.HasSuffix(selector, "%") {
			return KeyframeStop{}, fmt.Errorf("invalid keyframe selector: %q", selector)
		}
		val := strings.TrimSuffix(selector, "%")
		n, err := fmt.Sscanf(val, "%f", &percent)
		if err != nil || n != 1 {
			return KeyframeStop{}, fmt.Errorf("invalid keyframe percentage: %q", selector)
		}
		percent /= 100
	}

	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return KeyframeStop{}, fmt.Errorf("expected '{' after %q", selector)
	}
	p.pos++ // consume {

	props := make(map[string]string)
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		if p.input[p.pos] == '}' {
			p.pos++ // consume }
			return KeyframeStop{Percent: percent, Properties: props}, nil
		}
		name := p.readPropertyName()
		if name == "" {
			return KeyframeStop{}, fmt.Errorf("empty property name in keyframe")
		}
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == ':' {
			p.pos++ // consume :
		}
		p.skipWhitespace()
		value := p.readPropertyValue()
		props[name] = value
	}
	return KeyframeStop{}, fmt.Errorf("unterminated keyframe stop")
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isPropertyNameChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == '_'
}
