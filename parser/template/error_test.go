package template

import (
	"errors"
	"testing"
)

func TestErrorReturnsMessage(t *testing.T) {
	// Given
	e := &Error{Offset: 7, Msg: "boom"}

	// When
	got := e.Error()

	// Then
	if got != "boom" {
		t.Errorf("Error() = %q, want %q", got, "boom")
	}
}

func TestErrorfCapturesOffset(t *testing.T) {
	// Given
	p := &parser{input: "abcdef", pos: 4}

	// When
	err := p.errorf("bad %q", "x")

	// Then
	var perr *Error
	if !errors.As(err, &perr) {
		t.Fatalf("err is not *Error: %T", err)
	}
	if perr.Offset != 4 {
		t.Errorf("Offset = %d, want 4", perr.Offset)
	}
	if perr.Msg != `bad "x"` {
		t.Errorf("Msg = %q, want %q", perr.Msg, `bad "x"`)
	}
}

func TestParseReturnsTypedErrorWithOffset(t *testing.T) {
	// Given: an unexpected character partway through the input
	input := "  @oops"

	// When
	_, err := Parse(input)

	// Then
	var perr *Error
	if !errors.As(err, &perr) {
		t.Fatalf("err is not *Error: %T (%v)", err, err)
	}
	if perr.Offset != 2 {
		t.Errorf("Offset = %d, want 2 (the '@')", perr.Offset)
	}
}
