package sumitest

import (
	"testing"

	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/tui"
)

func newTestHarness(text string) *Harness {
	buf := render.NewBuffer(40, 1)
	buf.WriteText(0, 0, text)
	return New(&tui.App{TestBuffer: buf, OnRender: func() {}})
}

func newStyledTestHarness() *Harness {
	buf := render.NewBuffer(40, 1)
	buf.WriteStyledText(0, 0, "Hello", render.Style{FG: render.Color{Name: "green"}})
	return New(&tui.App{TestBuffer: buf, OnRender: func() {}})
}

func TestAssertTextPasses(t *testing.T) {
	h := newTestHarness("Hello")
	AssertText(t, h, "Hello")
}

func TestAssertTextFailsWithHelperT(t *testing.T) {
	// Given — use a fake T to capture failures
	fakeT := &testing.T{}
	h := newTestHarness("Hello")

	// When
	AssertText(fakeT, h, "World")

	// Then
	if !fakeT.Failed() {
		t.Error("expected AssertText to fail")
	}
}

func TestAssertStyledTextPasses(t *testing.T) {
	h := newStyledTestHarness()
	AssertStyledText(t, h, "<<green>>Hello<</>>")
}

func TestAssertStyledTextFailsWithHelperT(t *testing.T) {
	fakeT := &testing.T{}
	h := newStyledTestHarness()

	AssertStyledText(fakeT, h, "<<red>>Hello<</>>")

	if !fakeT.Failed() {
		t.Error("expected AssertStyledText to fail")
	}
}

func TestAssertContainsPasses(t *testing.T) {
	h := newTestHarness("Hello World")
	AssertContains(t, h, "World")
}

func TestAssertContainsFailsWithHelperT(t *testing.T) {
	fakeT := &testing.T{}
	h := newTestHarness("Hello")

	AssertContains(fakeT, h, "World")

	if !fakeT.Failed() {
		t.Error("expected AssertContains to fail")
	}
}

func TestAssertStyledContainsPasses(t *testing.T) {
	h := newStyledTestHarness()
	AssertStyledContains(t, h, "<<green>>Hello<</>>")
}

func TestAssertStyledContainsFailsWithHelperT(t *testing.T) {
	fakeT := &testing.T{}
	h := newStyledTestHarness()

	AssertStyledContains(fakeT, h, "<<red>>")

	if !fakeT.Failed() {
		t.Error("expected AssertStyledContains to fail")
	}
}
