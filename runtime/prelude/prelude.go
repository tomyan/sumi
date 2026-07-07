// Package prelude re-exports the sumi runtime types and functions used by
// generated .sumi component code. This is the single implicit import in all
// generated files — user code imports everything else explicitly.
package prelude

import (
	"fmt"
	"strings"

	"github.com/tomyan/sumi/parser/style"
	"github.com/tomyan/sumi/runtime/anim"
	"github.com/tomyan/sumi/runtime/input"
	"github.com/tomyan/sumi/runtime/layout"
	"github.com/tomyan/sumi/runtime/render"
	"github.com/tomyan/sumi/runtime/signal"
	"github.com/tomyan/sumi/runtime/term"
	"github.com/tomyan/sumi/runtime/tui"
)

// --- layout ---

type Input = layout.Input
type DOMEvent = layout.DOMEvent
type Padding = layout.Padding
type ScrollState = layout.ScrollState
type TransitionSpec = anim.TransitionSpec
type AnimationSpec = anim.AnimationSpec
type TimingFunction = anim.TimingFunction
type KeyframeAnimation = anim.KeyframeAnimation
type KeyframeStop = anim.KeyframeStop
type Box = layout.Box

const (
	KindBox  = layout.KindBox
	KindText = layout.KindText
)

var (
	ParsePadding    = layout.ParsePadding
	Layout          = layout.Layout
	RenderTree      = layout.RenderTree
	ApplyChanges    = layout.ApplyChanges
	FindCursor      = layout.FindCursor
	DirectWriteText = layout.DirectWriteText
	DiffTrees       = layout.DiffTrees
)

// --- render ---

type Style = render.Style
type Color = render.Color
type Stylesheet = style.Stylesheet

// MustParseStylesheet parses embedded component CSS. The same text was
// already parsed successfully at codegen time, so failure is a build bug.
// ResolveStyles resolves component CSS against an input tree at runtime.
var ResolveStyles = layout.ResolveStyles

// ParseMargin parses a CSS margin shorthand.
var ParseMargin = layout.ParseMargin

func MustParseStylesheet(src string) *style.Stylesheet {
	ss, err := style.Parse(src)
	if err != nil {
		panic("sumi: embedded stylesheet failed to parse: " + err.Error())
	}
	return ss
}

type ColorPair = render.ColorPair

var (
	NewBuffer            = render.NewBuffer
	ClearScreen          = render.ClearScreen
	ShowCursor           = render.ShowCursor
	HideCursor           = render.HideCursor
	EnterAlternateScreen = render.EnterAlternateScreen
	ExitAlternateScreen  = render.ExitAlternateScreen
)

// --- signal ---

type Signal[T any] = signal.Signal[T]
type Computed[T any] = signal.Computed[T]

// New creates a new signal with an initial value.
func New[T any](initial T) *signal.Signal[T] {
	return signal.New(initial)
}

// From creates a computed signal derived from other signals.
func From[T any](fn func() T) *signal.Computed[T] {
	return signal.From(fn)
}

// Effect registers a side effect that re-runs when its signal dependencies change.
var Effect = signal.Effect

// Batch defers signal notifications until the batch function completes.
var Batch = signal.Batch

// --- tui ---

type App = tui.App
type Component = tui.Component
type RunOptions = tui.RunOptions

var (
	Run            = tui.Run
	RunWithOptions = tui.RunWithOptions
	TestApp        = tui.TestApp
	Quit           = tui.Quit

	// Native two-way binding display helpers (emitted by bind:value/bind:checked).
	BindInputValue  = tui.BindInputValue
	BindSelectValue = tui.BindSelectValue
	BindChecked     = tui.BindChecked
)

// Env returns a framework-provided signal for the given environment key.
func Env[T any](key string) *signal.Signal[T] {
	return tui.Env[T](key)
}

// --- input ---

type Event = input.Event
type SpecialKey = input.SpecialKey
type EventKind = input.EventKind

const (
	EventKey     = input.EventKey
	EventSpecial = input.EventSpecial
	EventMouse   = input.EventMouse
	EventSignal  = input.EventSignal
	EventFrame   = input.EventFrame
	EventPaste   = input.EventPaste
	EventFocus   = input.EventFocus
	EventBlur    = input.EventBlur
)

const (
	KeyUp        = input.KeyUp
	KeyDown      = input.KeyDown
	KeyLeft      = input.KeyLeft
	KeyRight     = input.KeyRight
	KeyHome      = input.KeyHome
	KeyEnd       = input.KeyEnd
	KeyPgUp      = input.KeyPgUp
	KeyPgDn      = input.KeyPgDn
	KeyTab       = input.KeyTab
	KeyShiftTab  = input.KeyShiftTab
	KeyEnter     = input.KeyEnter
	KeyEscape    = input.KeyEscape
	KeyBackspace = input.KeyBackspace
	KeyDelete    = input.KeyDelete
)

// --- term ---

var GetSize = term.GetSize

// --- fmt ---

var Sprint = fmt.Sprint
var Sprintf = fmt.Sprintf

// --- dynamic attributes ---

// SplitClasses splits a class attribute value into class names.
func SplitClasses(s string) []string {
	return strings.Fields(s)
}

// AttrString renders a dynamic attribute value (bool, int, string, ...)
// as its attribute string form.
func AttrString(v any) string {
	return fmt.Sprint(v)
}
