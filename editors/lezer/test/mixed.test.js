// Mixed-nesting tests: the configured parser (src/index.js) overlays the
// real Go and CSS parsers on the language islands, so the tree should carry
// Go/CSS node types inside the sumi shell.

import assert from "assert";
import { sumiLanguage } from "../src/index.js";

const parse = (src) => sumiLanguage.parser.parse(src);

function nodeNames(tree) {
  const names = new Set();
  tree.iterate({ enter: (n) => names.add(n.name) });
  return names;
}

describe("mixed nesting", () => {
  it("nests Go into a <script> body", () => {
    const names = nodeNames(parse(`<script>func inc() { x.Set(1) }</script>`));
    assert(names.has("ScriptSection"), "sumi shell node present");
    assert(names.has("FunctionDecl"), "Go FunctionDecl nested");
    assert(names.has("VariableName"), "Go VariableName nested");
  });

  it("nests CSS into a <style> body", () => {
    const names = nodeNames(parse(`<style>.card { color: red; }</style>`));
    assert(names.has("StyleSection"), "sumi shell node present");
    assert(names.has("RuleSet"), "CSS RuleSet nested");
    assert(names.has("PropertyName"), "CSS PropertyName nested");
  });

  it("nests Go into an {if} condition", () => {
    const names = nodeNames(parse(`{if count.Get() > 0}<p>x</p>{/if}`));
    assert(names.has("IfBlock"), "sumi IfBlock present");
    assert(names.has("CompareOp"), "Go CompareOp nested in condition");
    assert(names.has("CallExpr"), "Go CallExpr nested in condition");
  });

  it("nests Go into an {expr} interpolation", () => {
    const names = nodeNames(parse(`<p>{user.Name}</p>`));
    assert(names.has("Interpolation"), "sumi Interpolation present");
    assert(names.has("SelectorExpr") || names.has("FieldName"), "Go selector nested");
  });

  it("balances braces in an attribute expression", () => {
    const names = nodeNames(parse(`<div class={fn(map[string]int{})} />`));
    assert(names.has("AttrInterpolation"), "sumi AttrInterpolation present");
    assert(names.has("MapType"), "Go MapType nested (balanced braces read whole)");
  });
});
