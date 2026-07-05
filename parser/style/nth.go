package style

import (
	"fmt"
	"strconv"
	"strings"
)

// Nth is a parsed An+B expression from an :nth-* pseudo-class.
type Nth struct {
	A, B int
}

// Matches reports whether the 1-based index satisfies An+B: index = A*n + B
// for some non-negative integer n.
func (n Nth) Matches(index int) bool {
	if n.A == 0 {
		return index == n.B
	}
	diff := index - n.B
	return diff%n.A == 0 && diff/n.A >= 0
}

// ParseNth parses an An+B expression: odd, even, 5, n, 2n, 2n+1, -n+3, 3n-1.
// Whitespace is not permitted (selector tokenization would have split it).
func ParseNth(expr string) (Nth, error) {
	expr = strings.TrimSpace(strings.ToLower(expr))
	switch expr {
	case "odd":
		return Nth{A: 2, B: 1}, nil
	case "even":
		return Nth{A: 2}, nil
	case "":
		return Nth{}, fmt.Errorf("empty nth expression")
	}

	nPos := strings.IndexByte(expr, 'n')
	if nPos < 0 {
		b, err := strconv.Atoi(expr)
		if err != nil {
			return Nth{}, fmt.Errorf("invalid nth expression %q", expr)
		}
		return Nth{B: b}, nil
	}

	a, err := parseNthCoefficient(expr[:nPos])
	if err != nil {
		return Nth{}, fmt.Errorf("invalid nth expression %q", expr)
	}
	rest := expr[nPos+1:]
	if rest == "" {
		return Nth{A: a}, nil
	}
	b, err := strconv.Atoi(rest)
	if err != nil {
		return Nth{}, fmt.Errorf("invalid nth expression %q", expr)
	}
	return Nth{A: a, B: b}, nil
}

// parseNthCoefficient parses the A part before 'n': "", "+", "-", or digits.
func parseNthCoefficient(s string) (int, error) {
	switch s {
	case "", "+":
		return 1, nil
	case "-":
		return -1, nil
	}
	return strconv.Atoi(s)
}
