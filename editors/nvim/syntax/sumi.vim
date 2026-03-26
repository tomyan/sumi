" Vim syntax file for sumi (.sumi) single-file components
" Language: Sumi (Go + Terminal CSS + HTML-like templates)

if exists("b:current_syntax")
  finish
endif

" --- Script block: embed Go syntax ---
syntax include @goSyntax syntax/go.vim
unlet! b:current_syntax

syntax region sumiScript
      \ matchgroup=sumiTag
      \ start="<script>"
      \ end="</script>"
      \ contains=@goSyntax
      \ keepend

" --- Style block: embed CSS syntax ---
syntax include @cssSyntax syntax/css.vim
unlet! b:current_syntax

syntax region sumiStyle
      \ matchgroup=sumiTag
      \ start="<style>"
      \ end="</style>"
      \ contains=@cssSyntax
      \ keepend

" Reset case sensitivity (CSS syntax sets 'syn case ignore' which leaks).
syntax case match

" --- Template (top-level, outside script/style) ---

" HTML-like element tags
syntax match sumiTagName /<\/\?\(box\|text\|title\)\>/
syntax match sumiSelfClose /\/>/

" Component tags (PascalCase)
syntax match sumiComponent /<\/\?[A-Z][A-Za-z]*/

" Slot tags
syntax match sumiSlotTag /<\/\?slot:[a-z][a-z0-9]*/
syntax match sumiSlotDefault /<slot:default\s*\/>/

" Attributes
syntax match sumiAttrName /\<[a-z][a-z0-9-]*=\@=/
syntax match sumiBindAttr /\<bind:[a-z][a-z0-9-]*/
syntax region sumiAttrString start=/"/ end=/"/ oneline

" Template expressions {expr}
syntax region sumiExpr matchgroup=sumiExprDelim start="{" end="}" oneline

" Control flow
syntax match sumiControlFlow /{\(if\|else\|for\|slot\|snippet\|render\)\>/
syntax match sumiControlEnd /{\(\/if\|\/for\|\/slot\|\/snippet\)}/

" --- Highlights ---
highlight link sumiTag Keyword
highlight link sumiTagName Statement
highlight link sumiSelfClose Delimiter
highlight link sumiComponent Type
highlight link sumiSlotTag PreProc
highlight link sumiSlotDefault PreProc
highlight link sumiAttrName Identifier
highlight link sumiBindAttr Special
highlight link sumiAttrString String
highlight link sumiExpr Special
highlight link sumiExprDelim Delimiter
highlight link sumiControlFlow Conditional
highlight link sumiControlEnd Conditional

let b:current_syntax = "sumi"
