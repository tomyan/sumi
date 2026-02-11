package codegen

import (
	"go/scanner"
	"go/token"
	"strings"
)

// prefixConditionExpr tokenizes a Go expression and prefixes known variable
// identifiers with "c." for component receiver access.
func prefixConditionExpr(expr string, varNames map[string]bool) string {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(expr))
	s.Init(file, []byte(expr), nil, 0)

	type tokenInfo struct {
		pos int
		end int
		lit string
	}
	var replacements []tokenInfo

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.IDENT && varNames[lit] {
			offset := int(pos) - file.Base()
			replacements = append(replacements, tokenInfo{
				pos: offset,
				end: offset + len(lit),
				lit: lit,
			})
		}
	}

	if len(replacements) == 0 {
		return expr
	}

	var sb strings.Builder
	prev := 0
	for _, r := range replacements {
		sb.WriteString(expr[prev:r.pos])
		sb.WriteString("c.")
		sb.WriteString(r.lit)
		prev = r.end
	}
	sb.WriteString(expr[prev:])
	return sb.String()
}
