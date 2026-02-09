package script

// Script represents the parsed contents of a <script> block.
type Script struct {
	StateDecls []StateDecl
	PropDecls  []PropDecl
	FuncDecls  []FuncDecl
}

// StateDecl represents a reactive state declaration: name := $state(initExpr)
type StateDecl struct {
	Name     string // variable name
	InitExpr string // initial value expression, e.g. "0", `"hello"`, `[]string{"a","b"}`
}

// PropDecl represents a component prop declaration: name := $prop(defaultExpr)
type PropDecl struct {
	Name        string // variable name
	DefaultExpr string // default value expression
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

		if decl, ok, err := p.tryParseStateDecl(); err != nil {
			return nil, err
		} else if ok {
			s.StateDecls = append(s.StateDecls, decl)
			continue
		}

		if pdecl, ok, err := p.tryParsePropDecl(); err != nil {
			return nil, err
		} else if ok {
			s.PropDecls = append(s.PropDecls, pdecl)
			continue
		}

		if fdecl, ok, err := p.tryParseFuncDecl(); err != nil {
			return nil, err
		} else if ok {
			s.FuncDecls = append(s.FuncDecls, fdecl)
			continue
		}

		p.skipLine()
	}

	resolveStateAssignments(s)
	return s, nil
}

// resolveStateAssignments detects state assignments in all function bodies.
// Both state and prop variables are reactive, so assignments to either are tracked.
func resolveStateAssignments(s *Script) {
	stateNames := make(map[string]bool)
	for _, stateDecl := range s.StateDecls {
		stateNames[stateDecl.Name] = true
	}
	for _, propDecl := range s.PropDecls {
		stateNames[propDecl.Name] = true
	}
	for i := range s.FuncDecls {
		s.FuncDecls[i].StateAssignments = findStateAssignments(s.FuncDecls[i].Body, stateNames)
	}
}
