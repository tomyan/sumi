package css

import (
	"strconv"
	"strings"
)

// IsCalcValue reports whether a property value is a CSS math function.
func IsCalcValue(v string) bool {
	for _, fn := range []string{"calc(", "min(", "max(", "clamp("} {
		if strings.HasPrefix(v, fn) {
			return true
		}
	}
	return false
}

// EvalCalc evaluates a CSS math value (calc/min/max/clamp) to whole cells.
// Percentages resolve against percentBase; pass a negative base when no
// containing-block size is available — any percentage then fails the value.
func EvalCalc(value string, percentBase int) (int, bool) {
	p := &calcParser{input: value, base: percentBase}
	v, ok := p.parseFactor()
	if !ok {
		return 0, false
	}
	p.skipSpace()
	if p.pos != len(p.input) {
		return 0, false
	}
	return int(v + 0.5), true
}

type calcParser struct {
	input string
	pos   int
	base  int
}

func (p *calcParser) skipSpace() {
	for p.pos < len(p.input) && p.input[p.pos] == ' ' {
		p.pos++
	}
}

// parseExpr := term (('+'|'-') term)*
func (p *calcParser) parseExpr() (float64, bool) {
	v, ok := p.parseTerm()
	if !ok {
		return 0, false
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.input) || (p.input[p.pos] != '+' && p.input[p.pos] != '-') {
			return v, true
		}
		op := p.input[p.pos]
		p.pos++
		rhs, ok := p.parseTerm()
		if !ok {
			return 0, false
		}
		if op == '+' {
			v += rhs
		} else {
			v -= rhs
		}
	}
}

// parseTerm := factor (('*'|'/') factor)*
func (p *calcParser) parseTerm() (float64, bool) {
	v, ok := p.parseFactor()
	if !ok {
		return 0, false
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.input) || (p.input[p.pos] != '*' && p.input[p.pos] != '/') {
			return v, true
		}
		op := p.input[p.pos]
		p.pos++
		rhs, ok := p.parseFactor()
		if !ok {
			return 0, false
		}
		if op == '*' {
			v *= rhs
		} else {
			if rhs == 0 {
				return 0, false
			}
			v /= rhs
		}
	}
}

// parseFactor := number[unit] | fn '(' args ')' | '(' expr ')'
func (p *calcParser) parseFactor() (float64, bool) {
	p.skipSpace()
	if p.pos >= len(p.input) {
		return 0, false
	}
	if p.input[p.pos] == '(' {
		p.pos++
		v, ok := p.parseExpr()
		if !ok || !p.expect(')') {
			return 0, false
		}
		return v, true
	}
	for _, fn := range []string{"calc", "min", "max", "clamp"} {
		if strings.HasPrefix(p.input[p.pos:], fn+"(") {
			p.pos += len(fn) + 1
			return p.parseFunc(fn)
		}
	}
	return p.parseNumber()
}

func (p *calcParser) parseFunc(fn string) (float64, bool) {
	var args []float64
	for {
		v, ok := p.parseExpr()
		if !ok {
			return 0, false
		}
		args = append(args, v)
		p.skipSpace()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if !p.expect(')') {
		return 0, false
	}
	return applyCalcFunc(fn, args)
}

func applyCalcFunc(fn string, args []float64) (float64, bool) {
	switch fn {
	case "calc":
		if len(args) != 1 {
			return 0, false
		}
		return args[0], true
	case "min":
		return fold(args, func(a, b float64) float64 {
			if b < a {
				return b
			}
			return a
		})
	case "max":
		return fold(args, func(a, b float64) float64 {
			if b > a {
				return b
			}
			return a
		})
	case "clamp":
		if len(args) != 3 {
			return 0, false
		}
		v := args[1]
		if v < args[0] {
			v = args[0]
		}
		if v > args[2] {
			v = args[2]
		}
		return v, true
	}
	return 0, false
}

func fold(args []float64, f func(a, b float64) float64) (float64, bool) {
	if len(args) == 0 {
		return 0, false
	}
	v := args[0]
	for _, a := range args[1:] {
		v = f(v, a)
	}
	return v, true
}

func (p *calcParser) parseNumber() (float64, bool) {
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if (ch >= '0' && ch <= '9') || ch == '.' || (p.pos == start && (ch == '-' || ch == '+')) {
			p.pos++
			continue
		}
		break
	}
	if p.pos == start {
		return 0, false
	}
	n, err := strconv.ParseFloat(p.input[start:p.pos], 64)
	if err != nil {
		return 0, false
	}
	switch {
	case p.consume("%"):
		if p.base < 0 {
			return 0, false
		}
		return n / 100 * float64(p.base), true
	case p.consume("cell"), p.consume("ch"):
	}
	return n, true
}

func (p *calcParser) consume(s string) bool {
	if strings.HasPrefix(p.input[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *calcParser) expect(ch byte) bool {
	p.skipSpace()
	if p.pos < len(p.input) && p.input[p.pos] == ch {
		p.pos++
		return true
	}
	return false
}
