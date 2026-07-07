import { EditorView, lineNumbers } from "@codemirror/view";
import { EditorState } from "@codemirror/state";
import { HighlightStyle, syntaxHighlighting } from "@codemirror/language";
import { tags as t } from "@lezer/highlight";
import { sumi } from "../src/index.js";

// Deliberately spaced hues so the four acceptance-gate tokens read as
// clearly different colours: Go keyword (purple), CSS property (teal),
// tag name (green), control keyword (red).
const highlight = HighlightStyle.define([
  { tag: t.controlKeyword, color: "#cf222e", fontWeight: "bold" },
  { tag: [t.keyword, t.definitionKeyword, t.moduleKeyword], color: "#8250df" },
  { tag: t.tagName, color: "#116329" },
  { tag: t.typeName, color: "#953800" },
  { tag: t.propertyName, color: "#1098ad" },
  { tag: t.attributeName, color: "#9a6700" },
  { tag: [t.string, t.attributeValue], color: "#0a3069" },
  { tag: t.number, color: "#0550ae" },
  { tag: [t.variableName, t.function(t.variableName)], color: "#24292f" },
  { tag: [t.propertyName, t.definition(t.propertyName)], color: "#1098ad" },
  { tag: [t.angleBracket, t.brace, t.operator, t.definitionOperator], color: "#57606a" },
  { tag: t.comment, color: "#6e7781", fontStyle: "italic" },
  { tag: t.content, color: "#24292f" },
]);

const sample = `<sumi:imports>
import "strings"
</sumi:imports>

<script>
items := sumi.New([]string{"Buy groceries", "Write tests"})
selected := sumi.New(0)

func handleKey(evt sumi.Event) {
    if evt.Rune == 'd' {
        idx := selected.Get()
        items.Set(append(items.Get()[:idx], items.Get()[idx+1:]...))
    }
}
</script>

<style>
.container { border: single; border-color: cyan; padding: 1 2; }
.title     { color: green; font-weight: bold; }
.selected  { color: cyan; font-weight: bold; }
</style>

<div class="container" onkey={handleKey}>
    <div class="title">Todo ({strings.Title("list")})</div>
    <Divider label="items" />
    {for i, item := range items.Get() key=item}
        {if i == selected.Get()}
            <div class="selected">> {item}</div>
        {else}
            <div>  {item}</div>
        {/if}
    {/for}
    {snippet footer(n int)}<span>{n} left</span>{/snippet}
    {render footer(len(items.Get()))}
</div>
`;

new EditorView({
  parent: document.getElementById("editor"),
  state: EditorState.create({
    doc: sample,
    extensions: [
      lineNumbers(),
      sumi(),
      syntaxHighlighting(highlight),
      EditorView.theme({
        "&": { fontSize: "13px", height: "100%" },
        ".cm-content": { fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace" },
      }),
    ],
  }),
});
