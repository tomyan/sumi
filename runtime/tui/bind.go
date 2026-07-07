package tui

import (
	"unicode/utf8"

	"github.com/tomyan/sumi/runtime/edit"
	"github.com/tomyan/sumi/runtime/layout"
)

// BindInputValue reconciles a controlled text input or textarea with its
// bound value. Typing keeps the edit state and the signal equal, so this is
// a no-op then; an external change adopts the new value and moves the cursor
// to the end. Codegen emits it as the display half of bind:value.
func BindInputValue(n *layout.Input, v string) {
	if n.Attrs == nil {
		n.Attrs = map[string]string{}
	}
	n.Attrs["value"] = v
	if n.Edit == nil {
		n.Edit = &edit.State{Value: v, Cursor: utf8.RuneCountInString(v)}
		return
	}
	if n.Edit.Value != v {
		n.Edit.Value = v
		n.Edit.Cursor = utf8.RuneCountInString(v)
	}
}

// BindSelectValue selects the option whose value matches v (falling back to
// the option's label when it has no value attribute). Codegen emits it as
// the display half of bind:value on a select.
func BindSelectValue(n *layout.Input, v string) {
	opts := selectOptions(n)
	for i, o := range opts {
		if optionValue(o) == v {
			setSelectedOption(opts, i)
			return
		}
	}
}

// BindChecked writes the checked state of a checkbox or radio. Codegen emits
// it as the display half of bind:checked.
func BindChecked(n *layout.Input, checked bool) {
	setChecked(n, checked)
}
