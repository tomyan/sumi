package script

import (
	"fmt"
	"strings"
	"unicode"
)

// Script represents the parsed contents of a <script> block.
type Script struct {
	StateDecls []StateDecl
	FuncDecls  []FuncDecl
}

// StateDecl represents a reactive state declaration: name := $state(initExpr)
type StateDecl struct {
	Name     string // variable name
	InitExpr string // initial value expression, e.g. "0", `"hello"`, `[]string{"a","b"}`
}

// FuncDecl represents a function declaration within the script block.
type FuncDecl struct {
	Name             string            // function name
	Params           string            // parameter list, e.g. "" or "key string"
	Body             string            // function body (raw Go code between braces)
	StateAssignments []StateAssignment // assignments to state variables within the body
}

// StateAssignment records an assignment to a known state variable within a function body.
type StateAssignment struct {
	VarName string // which state var is being assigned
	Line    string // the full assignment line (trimmed)
}

// Parse parses a script block containing $state declarations and function definitions.
func Parse(input string) (*Script, error) {
	p := &parser{input: input, pos: 0}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() (*Script, error) {
	s := &Script{}

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		// Try to parse a $state declaration: name := $state(expr)
		if decl, ok, err := p.tryParseStateDecl(); err != nil {
			return nil, err
		} else if ok {
			s.StateDecls = append(s.StateDecls, decl)
			continue
		}

		// Try to parse a function declaration: func name(params) { body }
		if fdecl, ok, err := p.tryParseFuncDecl(); err != nil {
			return nil, err
		} else if ok {
			s.FuncDecls = append(s.FuncDecls, fdecl)
			continue
		}

		// Skip unrecognized lines
		p.skipLine()
	}

	// Detect state assignments in function bodies
	stateNames := make(map[string]bool)
	for _, sd := range s.StateDecls {
		stateNames[sd.Name] = true
	}
	for i := range s.FuncDecls {
		s.FuncDecls[i].StateAssignments = findStateAssignments(s.FuncDecls[i].Body, stateNames)
	}

	return s, nil
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func (p *parser) skipLine() {
	for p.pos < len(p.input) && p.input[p.pos] != '\n' {
		p.pos++
	}
	if p.pos < len(p.input) {
		p.pos++ // skip the newline
	}
}

// tryParseStateDecl tries to parse: name := $state(expr)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParseStateDecl() (StateDecl, bool, error) {
	saved := p.pos

	// Read identifier
	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return StateDecl{}, false, nil
	}

	p.skipInlineWhitespace()

	// Check for :=
	if !p.matchString(":=") {
		p.pos = saved
		return StateDecl{}, false, nil
	}

	p.skipInlineWhitespace()

	// Check for $state(
	if !p.matchString("$state(") {
		p.pos = saved
		return StateDecl{}, false, nil
	}

	// Read the init expression, matching parens
	expr, err := p.readBalancedParenContents()
	if err != nil {
		return StateDecl{}, false, fmt.Errorf("unterminated $state expression for %q: %w", name, err)
	}

	return StateDecl{Name: name, InitExpr: expr}, true, nil
}

// tryParseFuncDecl tries to parse: func name(params) { body }
func (p *parser) tryParseFuncDecl() (FuncDecl, bool, error) {
	saved := p.pos

	// Check for "func" keyword
	if !p.matchString("func") {
		p.pos = saved
		return FuncDecl{}, false, nil
	}

	// Must be followed by whitespace (not part of a larger identifier)
	if p.pos < len(p.input) && isIdentChar(p.input[p.pos]) {
		p.pos = saved
		return FuncDecl{}, false, nil
	}

	p.skipInlineWhitespace()

	// Read function name
	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return FuncDecl{}, false, nil
	}

	p.skipInlineWhitespace()

	// Read params between parens
	if p.pos >= len(p.input) || p.input[p.pos] != '(' {
		p.pos = saved
		return FuncDecl{}, false, nil
	}
	p.pos++ // skip (
	params, err := p.readUntilByte(')')
	if err != nil {
		return FuncDecl{}, false, fmt.Errorf("unterminated function parameter list for %q", name)
	}
	p.pos++ // skip )

	p.skipInlineWhitespace()

	// Read body between braces
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		p.pos = saved
		return FuncDecl{}, false, nil
	}

	body, err := p.readBalancedBraceContents()
	if err != nil {
		return FuncDecl{}, false, fmt.Errorf("unterminated function body for %q: %w", name, err)
	}

	return FuncDecl{Name: name, Params: params, Body: body}, true, nil
}

func (p *parser) readIdentifier() string {
	start := p.pos
	if p.pos < len(p.input) && (unicode.IsLetter(rune(p.input[p.pos])) || p.input[p.pos] == '_') {
		p.pos++
	} else {
		return ""
	}
	for p.pos < len(p.input) && isIdentChar(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func (p *parser) skipInlineWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t') {
		p.pos++
	}
}

func (p *parser) matchString(s string) bool {
	if p.pos+len(s) <= len(p.input) && p.input[p.pos:p.pos+len(s)] == s {
		p.pos += len(s)
		return true
	}
	return false
}

// readBalancedParenContents reads content inside parens, handling nested parens and strings.
// Assumes the opening '(' has already been consumed. Consumes the closing ')'.
func (p *parser) readBalancedParenContents() (string, error) {
	depth := 1
	start := p.pos

	for p.pos < len(p.input) && depth > 0 {
		ch := p.input[p.pos]
		switch ch {
		case '(':
			depth++
			p.pos++
		case ')':
			depth--
			if depth == 0 {
				result := p.input[start:p.pos]
				p.pos++ // consume closing )
				return result, nil
			}
			p.pos++
		case '"':
			p.pos++
			if err := p.skipDoubleQuotedString(); err != nil {
				return "", err
			}
		case '`':
			p.pos++
			if err := p.skipBacktickString(); err != nil {
				return "", err
			}
		case '\'':
			p.pos++
			if err := p.skipSingleQuotedChar(); err != nil {
				return "", err
			}
		default:
			p.pos++
		}
	}

	return "", fmt.Errorf("unexpected end of input, expected ')'")
}

// readBalancedBraceContents reads content inside braces, handling nested braces and strings.
// Assumes the opening '{' has already been consumed. Consumes the closing '}'.
func (p *parser) readBalancedBraceContents() (string, error) {
	depth := 1
	start := p.pos
	p.pos++ // skip past opening {

	for p.pos < len(p.input) && depth > 0 {
		ch := p.input[p.pos]
		switch ch {
		case '{':
			depth++
			p.pos++
		case '}':
			depth--
			if depth == 0 {
				result := p.input[start+1 : p.pos]
				p.pos++ // consume closing }
				return result, nil
			}
			p.pos++
		case '"':
			p.pos++
			if err := p.skipDoubleQuotedString(); err != nil {
				return "", err
			}
		case '`':
			p.pos++
			if err := p.skipBacktickString(); err != nil {
				return "", err
			}
		case '\'':
			p.pos++
			if err := p.skipSingleQuotedChar(); err != nil {
				return "", err
			}
		default:
			p.pos++
		}
	}

	return "", fmt.Errorf("unexpected end of input, expected '}'")
}

func (p *parser) skipDoubleQuotedString() error {
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' {
			p.pos += 2 // skip escape sequence
			continue
		}
		if p.input[p.pos] == '"' {
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated string literal")
}

func (p *parser) skipBacktickString() error {
	for p.pos < len(p.input) {
		if p.input[p.pos] == '`' {
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated raw string literal")
}

func (p *parser) skipSingleQuotedChar() error {
	for p.pos < len(p.input) {
		if p.input[p.pos] == '\\' {
			p.pos += 2
			continue
		}
		if p.input[p.pos] == '\'' {
			p.pos++
			return nil
		}
		p.pos++
	}
	return fmt.Errorf("unterminated character literal")
}

func (p *parser) readUntilByte(b byte) (string, error) {
	start := p.pos
	for p.pos < len(p.input) {
		if p.input[p.pos] == b {
			return p.input[start:p.pos], nil
		}
		p.pos++
	}
	return "", fmt.Errorf("unexpected end of input, expected %q", string(b))
}

// findStateAssignments scans function body lines for assignments to known state variables.
// Looks for "stateVar = expr" patterns (plain = not :=).
func findStateAssignments(body string, stateNames map[string]bool) []StateAssignment {
	if len(stateNames) == 0 {
		return nil
	}

	var assignments []StateAssignment
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Look for "identifier = " pattern (but not ":=" or "==")
		for name := range stateNames {
			if !strings.HasPrefix(trimmed, name) {
				continue
			}
			rest := trimmed[len(name):]
			rest = strings.TrimLeft(rest, " \t")
			// Must start with = but not := or ==
			if len(rest) > 0 && rest[0] == '=' && (len(rest) < 2 || (rest[1] != '=' && rest[1] != ':')) {
				// Check that what precedes the name is not part of a larger identifier
				// Since we checked HasPrefix on trimmed, the name is at the start
				// But also verify the char after name isn't an ident char
				afterName := trimmed[len(name):]
				if len(afterName) > 0 && isIdentChar(afterName[0]) {
					continue
				}
				assignments = append(assignments, StateAssignment{
					VarName: name,
					Line:    trimmed,
				})
			}
		}
	}

	return assignments
}
