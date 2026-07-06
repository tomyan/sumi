# Selectors and the cascade

The `<style>` block accepts a subset of CSS. This chapter documents the selector
surface, the cascade, at-rules, custom properties, and math functions that the
parser and resolver actually implement. Anything outside this set is parsed
leniently and dropped rather than erroring.

## Simple selectors

- Type: `button`, `div` — matches by tag name.
- Class: `.hint` — chain for AND (`.card.active` needs both classes).
- ID: `#sidebar`.
- Universal: `*` — matches any element.

Attribute selectors match against an element's attributes (case-sensitive; the
attribute must be present):

| Selector | Matches when |
| --- | --- |
| `[type]` | the attribute is present |
| `[type="radio"]` | value equals |
| `[href^="https"]` | value starts with |
| `[src$=".png"]` | value ends with |
| `[class*="col"]` | value contains |
| `[rel~="next"]` | value is one of the whitespace-separated words |
| `[lang\|="en"]` | value equals or starts with the value plus `-` |

Values may be quoted with single or double quotes. Deviation: the case-insensitive
flag (`[attr="v" i]`) is not supported.

## Combinators

All four combinators are supported:

```
<style>
nav a { }        /* descendant */
ul > li { }      /* child */
label + input { } /* adjacent sibling */
h2 ~ p { }       /* general sibling */
```

Whitespace around `>`, `+`, and `~` is optional (`.a+.b` works).

## Structural pseudo-classes

`:root`, `:empty`, `:first-child`, `:last-child`, `:only-child`,
`:nth-child()`, `:nth-last-child()`, `:first-of-type`, `:last-of-type`,
`:only-of-type`, `:nth-of-type()`, and `:nth-last-of-type()` are supported.
`:root` matches the synthetic tree root.

The `An+B` argument accepts `odd`, `even`, a plain integer, and forms like `2n`,
`2n+1`, `-n+3`, and `3n-1`. Deviation: no spaces are allowed inside the argument
(`2n + 1` does not parse); write `2n+1`.

```
<style>
li:first-child { color: cyan; }
tr:nth-child(odd) { background: black; }
</style>
```

## State pseudo-classes

- `:hover` — the element under the mouse. Hover only registers on elements that
  define a hover style.
- `:focus` — the element that currently holds focus.
- `:checked` — a checkbox or radio whose `checked` attribute is set.
- `:disabled` / `:enabled` — form controls (`input`, `button`, `textarea`,
  `select`, `option`) with or without a truthy `disabled` attribute.

```
<style>
button:focus { color: yellow; }
input:checked { color: green; }
button:disabled { opacity: dim; }
</style>
```

Deviation: `:active` parses but never matches — there is no active state at
runtime. State pseudo-classes are honoured only on the rightmost (subject)
compound of a selector; a state pseudo on an ancestor makes the rule inert.

## :not(), :is(), :where()

Each takes a comma-separated list of selectors. `:is()` and `:where()` match if
any argument matches; `:not()` matches if none do.

```
<style>
:is(h1, h2, h3) { font-weight: bold; }
p:not(.hint) { color: white; }
</style>
```

Deviation: each argument must be a single compound selector — arguments
containing combinators (`:is(.a > .b)`) never match. As in CSS, `:where()`
contributes zero specificity, while `:not()` and `:is()` take the specificity of
their most specific argument.

## Pseudo-elements

`::before` and `::after` (and the legacy single-colon `:before`/`:after`) insert
a generated child at the start or end of an element. The `content` value accepts
quoted string literals, `attr(name)` (substitutes an attribute value), `none`,
and space-separated concatenations of these.

```
<style>
li::before { content: "• "; }
a::after { content: " (" attr(href) ")"; }
</style>
```

Deviation: `content` supports only strings and `attr()` — no `counter()`, `url()`,
or other functions. A pseudo-element with no valid `content` is not generated.

## Specificity and the cascade

Specificity is the usual `(IDs, classes, types)` triple: IDs count IDs; classes,
attribute selectors, structural pseudo-classes, and state pseudo-classes count
classes; type selectors and pseudo-elements count types. Matching declarations
are applied in ascending specificity, with later source order breaking ties.

The user-agent stylesheet (element defaults) is layered underneath author rules,
so at equal specificity your rules win. There is no separate origin priority — it
is purely specificity plus source order.

Inline template attributes take precedence over CSS for layout properties. Layout
attributes written directly on an element (for example `width` or a template-set
class) always beat a CSS rule for the same property; visual properties (colour,
weight, decoration) come only from CSS.

## At-rules

`@media`, `@container`, and `@supports` are supported; their rules are flattened
into the cascade tagged with the condition. Conditions are joined with ` and `
only — there is no `or`, `not`, comma-OR, or media type (`screen`/`print`).

`@media` features:

- `min-width`, `max-width`, `min-height`, `max-height` — compared against the
  viewport in cells.
- `prefers-color-scheme` — `light` or `dark`.
- `prefers-reduced-motion` — `reduce` or `no-preference`.
- `display-mode` — matches `terminal`.

```
<style>
@media (min-width: 80) {
	.sidebar { width: 24; }
}
@media (prefers-color-scheme: dark) {
	.app { color: white; }
}
</style>
```

`@container` supports only size conditions (`min-width`/`max-width`/`min-height`/
`max-height`), evaluated against the nearest laid-out ancestor. Deviation: no
named containers.

`@supports` checks a `(property: value)` condition, but tests only whether the
**property name** is one sumi implements — the value is not validated. Custom
properties (`--*`) always pass.

Other at-rules (`@font-face`, `@import`, `@layer`, …) are parsed and dropped.
`@keyframes` is supported for animations.

## Custom properties

Any property beginning with `--` is a custom property. Custom properties inherit
down the tree, and `var(--name, fallback)` reads one with an optional fallback.
Unresolved references (with no fallback) drop the declaration.

```
<style>
.theme {
	--accent: cyan;
}
.theme button:focus {
	color: var(--accent, yellow);
}
</style>
```

## Math functions

`calc()`, `min()`, `max()`, and `clamp()` evaluate to whole cells. `calc()`
supports `+`, `-`, `*`, `/`, parentheses, and nesting; put spaces around `+` and
`-`. Units are cells (bare number, `cell`, or `ch`) and `%`; a percentage
resolves against the containing block. Deviation: no `px`/`em`/`rem`/`fr` in these
functions.

```
<style>
.panel {
	width: calc(100% - 4);
	height: clamp(10, 50%, 30);
}
</style>
```
