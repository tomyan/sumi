package script

import (
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/printer"
	"go/token"
	"strings"
)

// ScriptInfo contains information extracted from a Go AST analysis of a script block.
type ScriptInfo struct {
	Signals map[string]bool // variable names declared with signal.New or signal.From
	Props   []PropInfo      // var declarations (component props)
	Funcs   []FuncInfo      // function declarations
	Source  string          // the original source
}

// PropInfo describes a prop declared with var in the script block.
type PropInfo struct {
	Name    string // variable name
	TypeStr string // Go type as string (e.g. "string", "*signal.Signal[string]")
	Default string // default value expression, or "" if none
}

// FuncInfo describes a function declared in the script block.
type FuncInfo struct {
	Name   string // function name
	Params string // parameter list source (e.g. "evt input.Event")
	Body   string // function body source (between braces)
	Source string // full function source
}

// ParseGoAST parses a script block using Go's AST parser.
// Identifies signal declarations, prop declarations (var), and functions.
func ParseGoAST(src string) (*ScriptInfo, error) {
	// Strip function declarations (sumi syntax, not valid Go inside a function body).
	// Parse them separately with the existing script parser.
	strippedSrc := stripFuncDecls(src)

	// Wrap in a function to make it valid Go for the parser.
	wrapped := "package p\nfunc _script() {\n" + strippedSrc + "\n}\n"
	fset := token.NewFileSet()
	file, err := goparser.ParseFile(fset, "script.go", wrapped, goparser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse script: %w", err)
	}

	info := &ScriptInfo{
		Signals: make(map[string]bool),
		Source:  src,
	}

	// Also parse top-level var declarations (props).
	// These need to be parsed differently — as package-level declarations.
	propWrapped := "package p\n" + extractVarDecls(src)
	propFile, propErr := goparser.ParseFile(fset, "props.go", propWrapped, goparser.AllErrors)
	if propErr == nil {
		for _, decl := range propFile.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range vs.Names {
					pi := PropInfo{Name: name.Name}
					if vs.Type != nil {
						pi.TypeStr = exprToString(fset, vs.Type)
					}
					if i < len(vs.Values) {
						pi.Default = exprToString(fset, vs.Values[i])
					}
					info.Props = append(info.Props, pi)
				}
			}
		}
	}

	// Find the _script function body.
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "_script" {
			continue
		}
		for _, stmt := range fd.Body.List {
			switch s := stmt.(type) {
			case *ast.AssignStmt:
				if s.Tok == token.DEFINE {
					analyzeShortDecl(info, s, fset)
				}
			}
		}
	}

	// Extract functions from the source (go/ast wraps them as closures inside _script,
	// so we use the existing script parser for function extraction).
	info.Funcs = extractFuncs(src)

	return info, nil
}

// analyzeShortDecl checks if a := declaration is a signal.New or signal.From call.
func analyzeShortDecl(info *ScriptInfo, s *ast.AssignStmt, fset *token.FileSet) {
	if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
		return
	}
	ident, ok := s.Lhs[0].(*ast.Ident)
	if !ok {
		return
	}
	call, ok := s.Rhs[0].(*ast.CallExpr)
	if !ok {
		return
	}
	funcName := callFuncName(call)
	if funcName == "signal.New" || funcName == "sumi.New" ||
		funcName == "signal.From" || funcName == "sumi.From" ||
		funcName == "tui.Env" || funcName == "sumi.Env" {
		info.Signals[ident.Name] = true
	}
}

// callFuncName returns the dotted name of a call expression (e.g. "signal.New").
func callFuncName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if x, ok := fn.X.(*ast.Ident); ok {
			return x.Name + "." + fn.Sel.Name
		}
	case *ast.IndexExpr:
		// Generic call: signal.New[int](0) → the function is in fn.X
		if sel, ok := fn.X.(*ast.SelectorExpr); ok {
			if x, ok := sel.X.(*ast.Ident); ok {
				return x.Name + "." + sel.Sel.Name
			}
		}
	}
	return ""
}

// stripFuncDecls removes func declarations from the source so go/ast can parse the rest.
// Uses the existing script parser to find function boundaries.
func stripFuncDecls(src string) string {
	p := &parser{input: src, pos: 0}
	// Find all function spans to exclude.
	type span struct{ start, end int }
	var spans []span
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		startPos := p.pos
		if _, ok, _ := p.tryParseFuncDecl(); ok {
			spans = append(spans, span{startPos, p.pos})
			continue
		}
		p.skipLine()
	}
	if len(spans) == 0 {
		return src
	}
	// Build source without function spans.
	var result strings.Builder
	prev := 0
	for _, s := range spans {
		result.WriteString(src[prev:s.start])
		prev = s.end
	}
	result.WriteString(src[prev:])
	return result.String()
}

// extractVarDecls extracts lines starting with "var " from the source.
func extractVarDecls(src string) string {
	var lines []string
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "var ") {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

// extractFuncs uses the existing script parser to extract function declarations.
func extractFuncs(src string) []FuncInfo {
	p := &parser{input: src, pos: 0}
	var funcs []FuncInfo
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		if fd, ok, _ := p.tryParseFuncDecl(); ok {
			funcs = append(funcs, FuncInfo{
				Name:   fd.Name,
				Params: fd.Params,
				Body:   fd.Body,
			})
			continue
		}
		p.skipLine()
	}
	return funcs
}

// exprToString converts an AST expression back to source code.
func exprToString(fset *token.FileSet, expr ast.Expr) string {
	var buf strings.Builder
	printer.Fprint(&buf, fset, expr)
	return buf.String()
}
