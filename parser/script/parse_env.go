package script

import "fmt"

// tryParseEnvDecl tries to parse: name := $env(key)
// Returns the decl, whether it matched, and any error.
func (p *parser) tryParseEnvDecl() (EnvDecl, bool, error) {
	saved := p.pos

	name := p.readIdentifier()
	if name == "" {
		p.pos = saved
		return EnvDecl{}, false, nil
	}

	if !p.matchInlineSequence(":=", "$env(") {
		p.pos = saved
		return EnvDecl{}, false, nil
	}

	key, err := p.readBalancedParenContents()
	if err != nil {
		return EnvDecl{}, false, fmt.Errorf("unterminated $env expression for %q: %w", name, err)
	}

	return EnvDecl{Name: name, Key: key}, true, nil
}
