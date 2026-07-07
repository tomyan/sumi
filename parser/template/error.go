package template

import "fmt"

// Error is a template parse error carrying the byte offset within the
// template section where the problem was detected. Editor diagnostics
// map this offset back to a line/column in the source file.
type Error struct {
	Offset int
	Msg    string
}

// Error returns the human-readable message.
func (e *Error) Error() string {
	return e.Msg
}

// errorf builds an *Error positioned at the parser's current offset.
func (p *parser) errorf(format string, args ...any) error {
	return &Error{Offset: p.pos, Msg: fmt.Sprintf(format, args...)}
}
