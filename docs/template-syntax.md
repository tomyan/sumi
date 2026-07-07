# Template syntax

A `.sumi` file is one component: optional sections for imports, script,
and style, and a markup template. This chapter covers the file anatomy
and the template grammar — elements, attributes, text, and whitespace.
Signals and their auto-unwrap are covered in [signals](signals.md);
control-flow blocks in [control flow](control-flow.md).

## File anatomy

A file is split into up to four sections by their tags. Everything left
after the tagged sections are removed is the template.

| Section    | Delimiters                        | Contents            |
| ---------- | --------------------------------- | ------------------- |
| Imports    | `<sumi:imports>…</sumi:imports>`  | Go `import` lines.  |
| Script     | `<script>…</script>`              | Go declarations.    |
| Style      | `<style>…</style>`                | Scoped CSS.         |
| Template   | (everything else)                 | Markup.             |

All four are optional. The conventional order is imports, script, style,
template; editor diagnostics assume it when mapping errors back to
source. The imports section holds Go import lines and nothing else — the
generated file already imports the runtime, so list only the packages
your script and mounted child components need:

```sumi
<sumi:imports>
import "strings"
</sumi:imports>
```

Script declarations are plain Go: signal bindings (`count :=
sumi.New(0)`), `func` handlers, and `var` prop declarations. A script
that declares signals or `var` props compiles through the component path
and produces a `NewName(NameProps)` constructor; see
[components](components.md).

## Elements

Tag names fall into three groups:

- **HTML elements** — a fixed vocabulary (`div`, `span`, `p`, `ul`,
  `li`, `button`, `input`, `table`, …) rendered with a user-agent
  stylesheet. See [elements](elements.md).
- **Component references** — a capitalised tag (`<Card>`) mounts a
  same-package component; a dotted tag (`<pkg.Card>`) mounts an imported
  one. (Any non-HTML tag is treated as a component; use these two forms.)
- **`title`** — sets the terminal window title rather than rendering.

An element is a **text element** when its body is only text and
expressions, and a **container** when it holds child elements or
control-flow blocks. Most HTML tags can be either; a handful (`div`,
`section`, `ul`, `table`, …) stay containers even when their body is
plain text, so they keep borders, padding, and pseudo-element markers,
with the text becoming an implicit child.

### Self-closing and closing tags

An element with no body can self-close, and one with a body needs a
matching close tag:

```sumi
<input type="text" />
<Card title="Hi" />
<div class="row"><span>Hello</span></div>
```

Self-closing works for HTML elements and component references alike.

## Attributes

There are exactly three attribute forms:

| Form            | Meaning                                       |
| --------------- | --------------------------------------------- |
| `name="value"`  | Literal string.                               |
| `name={expr}`   | Expression — raw Go, no signal auto-unwrap.   |
| `{name}`        | Shorthand for `name={name}`.                  |

```sumi
<div class="card" id={rowID} />
<Card {title} />                   <!-- title={title} -->
```

There is no bare boolean attribute: write `disabled="true"`, not
`disabled`. Values that name a state — `checked`, `disabled`, `open`,
`selected`, `class` — accept an expression and are re-applied when the
signals in that expression change:

```sumi
<input type="checkbox" checked={enabled.Get()} />
```

Because attribute expressions are raw Go, a signal used there needs an
explicit `.Get()`. That is the reverse of template text, where the bare
name auto-unwraps — see [signals](signals.md).

## Text and expressions

Text mixes literal runs with `{expr}` interpolations. Inside text a bare
signal name auto-unwraps to a read; any other Go expression is used as
written:

```sumi
<p>Hello, {name} — you have {count} messages</p>
```

Braces do not nest inside a text expression: one `{` opens it and the
next `}` closes it, so a text expression cannot contain a literal `}`
(a Go composite literal, say). Attribute expressions do balance braces,
so `class={fn(map[string]int{})}` parses there.

## Whitespace

Whitespace between elements follows the JSX newline rule. A run of
whitespace that contains a newline is treated as source formatting and
dropped; a single-line gap collapses to one space. So the indentation
and line breaks you use to lay out the template do not appear in the
output, but a deliberate space between two inline elements on one line
does:

```sumi
<div>
    <span>a</span>
    <span>b</span>       <!-- no space between a and b: the gaps hold newlines -->
</div>
<div><span>a</span> <span>b</span></div>  <!-- one space: single-line gap -->
```

## Comments

There is no comment syntax. `<!-- … -->` is not recognised and is parsed
as a malformed tag, which is an error. To leave a note, put it in the Go
script as an ordinary `//` comment.
