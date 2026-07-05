package script

// Script represents the parsed contents of a <script> block.
type Script struct {
	FuncDecls     []FuncDecl
	SignalDecls   []SignalDecl   // signal.New() declarations (new reactive model)
	ComputedDecls []ComputedDecl // signal.From() declarations (new reactive model)
}

// SignalDecl represents a signal declaration: name := sumi.New(expr) or signal.New(expr)
type SignalDecl struct {
	Name     string // variable name
	InitExpr string // initial value expression
}

// ComputedDecl represents a computed signal: name := sumi.From(func() T { expr }) or signal.From(...)
type ComputedDecl struct {
	Name string // variable name
	Expr string // the full function literal expression
}

// FuncDecl represents a function declaration within the script block.
type FuncDecl struct {
	Name       string // function name
	Params     string // parameter list, e.g. "" or "key string"
	ReturnType string // return type, e.g. "string" or "" if void
	Body       string // function body (raw Go code between braces)
}

// Parse parses a script block containing function definitions.
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

		if fdecl, ok, err := p.tryParseFuncDecl(); err != nil {
			return nil, err
		} else if ok {
			s.FuncDecls = append(s.FuncDecls, fdecl)
			continue
		}

		p.skipLine()
	}

	return s, nil
}
