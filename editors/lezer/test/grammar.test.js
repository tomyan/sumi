// Tree-shape tests for the bare sumi grammar, driven by Lezer's standard
// test format (test/cases.txt). Covers every section form, all four
// attribute forms, component/element tags, each control block, snippets and
// render, self-closing, keyword-lookalike interpolation, and error recovery.

import { fileTests } from "@lezer/generator/test";
import { readFileSync } from "fs";
import { fileURLToPath } from "url";
import { dirname, join } from "path";
import { parser } from "../src/parser.js";

const dir = dirname(fileURLToPath(import.meta.url));
const spec = readFileSync(join(dir, "cases.txt"), "utf8");

describe("sumi grammar", () => {
  for (const { name, run } of fileTests(spec, "cases.txt")) {
    it(name, () => run(parser));
  }
});
