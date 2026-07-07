// External tokenizers for constructs the declarative @tokens block cannot
// express: raw-text section bodies (read until the closing tag), and the
// brace islands whose boundaries mirror the Go parser's rules — control
// blocks vs interpolation, non-nesting text braces vs balanced attribute
// braces.

import { ExternalTokenizer } from "@lezer/lr";
import {
  CtrlBrace,
  InterpBrace,
  InterpText,
  AttrInterpText,
  ClauseText,
  SnippetArgs,
  RawGo,
  RawCSS,
} from "./parser.terms.js";

const LBRACE = 123; // {
const RBRACE = 125; // }
const LPAREN = 40; // (
const RPAREN = 41; // )
const LT = 60; // <
const SLASH = 47; // /

// Grabs up to `n` characters starting at the current position without
// consuming them, so a tokenizer can classify a brace island by lookahead.
function peekString(input, n) {
  let out = "";
  for (let i = 0; i < n; i++) {
    const c = input.peek(i);
    if (c < 0) break;
    out += String.fromCharCode(c);
  }
  return out;
}

// classifyBrace mirrors parser_mixed.go's controlFlowTokens. The trailing
// delimiter matters — `{if ` opens a block, `{ifx}` is an interpolation of a
// variable named ifx. Block openers (if/for/snippet/render) become a bare
// CtrlBrace whose keyword the grammar tokenizes next; {else} and the block
// terminators are declined ("closer") so the whole-literal ElseTok/EndIf/…
// tokens match them instead; anything else is an interpolation.
function classifyBrace(ahead) {
  const rest = ahead.slice(1); // drop the leading "{"
  if (rest.startsWith("if ") || rest.startsWith("if\t")) return "opener";
  if (rest.startsWith("for ") || rest.startsWith("for\t")) return "opener";
  if (rest.startsWith("snippet ") || rest.startsWith("snippet\t")) return "opener";
  if (rest.startsWith("render ") || rest.startsWith("render\t")) return "opener";
  if (rest.startsWith("else}")) return "closer";
  if (rest.startsWith("/if}")) return "closer";
  if (rest.startsWith("/for}")) return "closer";
  if (rest.startsWith("/snippet}")) return "closer";
  return "interp";
}

// braceOpen classifies a `{` as opening a control block or an interpolation
// and consumes just the brace, so downstream states are unambiguous. It
// declines block terminators and {else} so their literal tokens win.
export const braceOpen = new ExternalTokenizer((input) => {
  if (input.next != LBRACE) return;
  const kind = classifyBrace(peekString(input, 12));
  if (kind == "closer") return;
  input.advance();
  input.acceptToken(kind == "opener" ? CtrlBrace : InterpBrace);
});

// textInterp reads a text-body interpolation body. Braces do not nest in
// text (parseTextParts stops at the first `}`), so we read to the first `}`.
export const textInterp = new ExternalTokenizer((input) => {
  const start = input.pos;
  while (input.next >= 0 && input.next != RBRACE) input.advance();
  if (input.pos > start) input.acceptToken(InterpText);
});

// attrInterp reads an attribute-value interpolation body. Attribute braces
// balance (readBracedValue), so `class={fn(map[string]int{})}` reads whole.
export const attrInterp = new ExternalTokenizer((input) => {
  const start = input.pos;
  let depth = 0;
  while (input.next >= 0) {
    if (input.next == LBRACE) {
      depth++;
    } else if (input.next == RBRACE) {
      if (depth == 0) break;
      depth--;
    }
    input.advance();
  }
  if (input.pos > start) input.acceptToken(AttrInterpText);
});

// clauseText reads an {if}/{for} clause, which the Go parser takes with
// readUntil('}') — non-nesting, up to the first `}`.
export const clauseText = new ExternalTokenizer((input) => {
  const start = input.pos;
  while (input.next >= 0 && input.next != RBRACE) input.advance();
  if (input.pos > start) input.acceptToken(ClauseText);
});

// snippetArgs reads a balanced (…) parameter or argument list following a
// snippet/render name.
export const snippetArgs = new ExternalTokenizer((input) => {
  if (input.next != LPAREN) return;
  let depth = 0;
  do {
    if (input.next == LPAREN) depth++;
    else if (input.next == RPAREN) depth--;
    input.advance();
  } while (input.next >= 0 && depth > 0);
  input.acceptToken(SnippetArgs);
});

// rawSection reads a section body up to its closing `</…>` tag. Emits RawCSS
// inside <style> and RawGo inside <script>/<sumi:imports>, so parseMixed can
// pick the right nested language.
export const rawSection = new ExternalTokenizer((input, stack) => {
  const start = input.pos;
  while (input.next >= 0) {
    if (input.next == LT && input.peek(1) == SLASH) break;
    input.advance();
  }
  if (input.pos == start) return;
  if (stack.canShift(RawCSS)) input.acceptToken(RawCSS);
  else input.acceptToken(RawGo);
});
