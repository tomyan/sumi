// CodeMirror 6 language support for the sumi template language.
//
// The sumi Lezer parser handles the template shell (tags, attributes,
// control blocks, interpolation) and hands the language islands to the real
// Go and CSS parsers via parseMixed: <script>/<sumi:imports> bodies, {if}/
// {for} clauses, and every {expression} nest as Go; <style> bodies as CSS.

import {
  LRLanguage,
  LanguageSupport,
  indentNodeProp,
  foldNodeProp,
  foldInside,
  delimitedIndent,
} from "@codemirror/language";
import { styleTags, tags as t } from "@lezer/highlight";
import { parseMixed } from "@lezer/common";
import { goLanguage } from "@codemirror/lang-go";
import { cssLanguage } from "@codemirror/lang-css";
import { parser } from "./parser.js";

// Nodes whose content is a Go island, and the one CSS island.
const GO_ISLANDS = new Set(["RawGo", "ClauseText", "InterpText", "AttrInterpText"]);

const mixed = parseMixed((node) => {
  if (GO_ISLANDS.has(node.type.name)) return { parser: goLanguage.parser };
  if (node.type.name == "RawCSS") return { parser: cssLanguage.parser };
  return null;
});

export const sumiLanguage = LRLanguage.define({
  name: "sumi",
  parser: parser.configure({
    wrap: mixed,
    props: [
      styleTags({
        "KwIf KwFor KwSnippet KwRender": t.controlKeyword,
        "ElseTok EndIf EndFor EndSnippet": t.controlKeyword,
        ElementName: t.tagName,
        ComponentName: t.typeName,
        AttributeName: t.attributeName,
        AttributeString: t.attributeValue,
        SnippetName: t.function(t.variableName),
        Text: t.content,
        "StartTag CloseTagStart TagEnd SelfCloseEnd": t.angleBracket,
        "InterpBrace CtrlBrace RBrace": t.brace,
        Eq: t.definitionOperator,
      }),
      indentNodeProp.add({
        "OpenTag IfBlock ForBlock SnippetBlock": delimitedIndent({ closing: "}" }),
      }),
      foldNodeProp.add({
        "IfBlock ForBlock SnippetBlock Element": foldInside,
      }),
    ],
  }),
  languageData: {
    // The template has no comment syntax; notes live in the Go <script>.
    closeBrackets: { brackets: ["<", "{", '"'] },
    indentOnInput: /^\s*(<\/|\{\/|\{else\})/,
  },
});

export function sumi() {
  return new LanguageSupport(sumiLanguage);
}
