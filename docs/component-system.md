# Component System

## Tiers

| Tier | Syntax | Import needed | Description |
|------|--------|---------------|-------------|
| Primitive | `<lowercase>` | No | Parser-level elements: `<box>`, `<text>` |
| Fundamental | `<lowercase>` | No | Framework components always available: `<textedit>`, `<scrollbar>` |
| Standard library | `<sumi:PascalCase>` | No | Higher-level composed components shipped with the framework: `<sumi:TextInput>` |
| User-defined | `<lib:PascalCase>` | Yes | Third-party or project-local component libraries |

## Imports

User-defined libraries are imported with a `<sumi:imports>` block at the top of the `.sumi` file, before `<script>`, template, or `<style>`:

```
<sumi:imports>
    "myui"
    alias "github.com/someone/otherui"
    * "github.com/someone/wildcard"
</sumi:imports>
```

- `"myui"` — import with default name, use as `<myui:Button />`
- `alias "path"` — import with alias, use as `<alias:Button />`
- `* "path"` — wildcard import, use as `<Button />` without prefix

### Conflict resolution

- Explicit prefix always wins over wildcard
- Two wildcard imports exposing the same component name is a compiler error — use a prefix to disambiguate

## Naming conventions

- Fundamental and primitive components are **lowercase**: `<textedit>`, `<scrollbar>`, `<box>`, `<text>`
- Standard library and user-defined components are **PascalCase** with a prefix: `<sumi:TextInput>`, `<myui:FancyButton>`
- Wildcard-imported user components are **PascalCase** without prefix: `<FancyButton>`
