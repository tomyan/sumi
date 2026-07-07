" Vim syntax file for sumi single-file components (.sumi)
" Go in <script>, CSS in <style>, HTML-ish template with {expr}
" interpolation and {if}/{for} control tags.
" Source of truth for scopes: editors/grammar/sumi.tmLanguage.json

if exists('b:current_syntax')
  finish
endif

" Embedded Go for the script block and expressions.
unlet! b:current_syntax
syn include @sumiGo syntax/go.vim
unlet! b:current_syntax
syn include @sumiCSS syntax/css.vim
unlet! b:current_syntax

" <script> ... </script> — Go
syn region sumiScript matchgroup=sumiTagName start=+<script>+ end=+</script>+ contains=@sumiGo keepend

" <style> ... </style> — CSS
syn region sumiStyle matchgroup=sumiTagName start=+<style>+ end=+</style>+ contains=@sumiCSS keepend

" Control tags: {if expr} {else} {else if expr} {for ... key=expr} {/if} {/for}
syn region sumiControl matchgroup=sumiControlDelim start=+{\ze\%(if\>\|else\>\|for\>\|/if}\|/for}\)+ end=+}+ contains=sumiControlKeyword,@sumiGo keepend
syn keyword sumiControlKeyword if else for key contained

" Text interpolation and expression attribute values: {expr}
syn region sumiExpr matchgroup=sumiExprDelim start=+{+ end=+}+ contains=@sumiGo contained containedin=sumiTag,sumiText keepend

" Tags and attributes
syn region sumiTag start=+</\=\%(script\|style\)\@![a-zA-Z]+ end=+/\=>+ contains=sumiTagName,sumiComponentName,sumiAttr,sumiEventAttr,sumiBindAttr,sumiString,sumiExpr keepend
syn match sumiTagName +</\=\zs[a-z][a-zA-Z0-9-]*+ contained
syn match sumiComponentName +</\=\zs[A-Z][a-zA-Z0-9]*+ contained
syn match sumiAttr +\<[a-zA-Z][a-zA-Z0-9-]*\ze=+ contained
syn match sumiEventAttr +\<on[a-z]\+\ze=+ contained
syn match sumiBindAttr +\<bind:[a-zA-Z-]\++ contained
syn region sumiString start=+"+ end=+"+ contained

" Top-level interpolation in text content
syn region sumiText start=+>+ end=+<+me=s-1 contained containedin=sumiTag transparent contains=sumiExpr

hi def link sumiTagName        Statement
hi def link sumiComponentName  Type
hi def link sumiAttr           Identifier
hi def link sumiEventAttr      Special
hi def link sumiBindAttr       Special
hi def link sumiString         String
hi def link sumiControlKeyword Conditional
hi def link sumiControlDelim   Delimiter
hi def link sumiExprDelim      Delimiter

let b:current_syntax = 'sumi'
