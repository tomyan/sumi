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
      \ contains=@goSyntax,sumiSignalCall
      \ keepend

" Signal API highlights (inside script)
syntax match sumiSignalCall /signal\.New\|signal\.From\|signal\.Effect\|signal\.Batch/ containedin=sumiScript
syntax match sumiSignalMethod /\.Get()\|\.Set(\|\.Update(/ containedin=sumiScript
syntax match sumiTuiCall /tui\.Env\|tui\.Quit\|tui\.Run/ containedin=sumiScript

" --- Style block: embed CSS syntax ---
syntax include @cssSyntax syntax/css.vim
unlet! b:current_syntax

syntax region sumiStyle
      \ matchgroup=sumiTag
      \ start="<style>"
      \ end="</style>"
      \ contains=@cssSyntax
      \ keepend

" --- Template region (everything outside script/style) ---

" HTML-like tags
syntax match sumiTagName /<\(box\|text\|title\)\>/ contained containedin=sumiTemplate
syntax match sumiTagClose /<\/\(box\|text\|title\)\>/ contained containedin=sumiTemplate
syntax match sumiSelfClose /\/>/ contained containedin=sumiTemplate
syntax match sumiAngleBracket /[<>]/ contained containedin=sumiTemplate

" Component tags (PascalCase)
syntax match sumiComponent /<[A-Z][A-Za-z]*/ contained containedin=sumiTemplate

" Slot tags
syntax match sumiSlotTag /<slot:[a-z][a-z0-9]*/ contained containedin=sumiTemplate
syntax match sumiSlotClose /<\/slot:[a-z][a-z0-9]*>/ contained containedin=sumiTemplate
syntax match sumiSlotDefault /<slot:default\s*\/>/ contained containedin=sumiTemplate

" Attributes
syntax match sumiAttrName /\<[a-z][a-z0-9-]*=/ contained containedin=sumiTemplate
syntax match sumiBindAttr /\<bind:[a-z][a-z0-9-]*=/ contained containedin=sumiTemplate
syntax region sumiAttrString start=/"/ end=/"/ contained containedin=sumiTemplate

" Template expressions {expr}
syntax region sumiExpr
      \ matchgroup=sumiExprDelim
      \ start="{"
      \ end="}"
      \ contained containedin=sumiTemplate
      \ contains=sumiExprContent

" Control flow keywords
syntax match sumiControlFlow /{\(if\|else\|for\|slot\|snippet\|render\)\>/ contained containedin=sumiTemplate
syntax match sumiControlEnd /{\(\/if\|\/for\|\/slot\|\/snippet\)}/ contained containedin=sumiTemplate

" CSS class shorthand
syntax match sumiClassName /class="[^"]*"/ contained containedin=sumiTemplate

" The template is everything not in script or style
syntax region sumiTemplate
      \ start="\%^"
      \ end="\%$"
      \ contains=sumiScript,sumiStyle,sumiTagName,sumiTagClose,sumiSelfClose,sumiComponent,sumiSlotTag,sumiSlotClose,sumiSlotDefault,sumiAttrName,sumiBindAttr,sumiAttrString,sumiExpr,sumiControlFlow,sumiControlEnd,sumiClassName,sumiAngleBracket

" --- Highlights ---

highlight link sumiTag Keyword
highlight link sumiSignalCall Function
highlight link sumiSignalMethod Special
highlight link sumiTuiCall Function
highlight link sumiTagName Statement
highlight link sumiTagClose Statement
highlight link sumiSelfClose Statement
highlight link sumiComponent Type
highlight link sumiSlotTag PreProc
highlight link sumiSlotClose PreProc
highlight link sumiSlotDefault PreProc
highlight link sumiAttrName Identifier
highlight link sumiBindAttr Special
highlight link sumiAttrString String
highlight link sumiExpr Special
highlight link sumiExprDelim Delimiter
highlight link sumiControlFlow Conditional
highlight link sumiControlEnd Conditional
highlight link sumiClassName Label
highlight link sumiAngleBracket Delimiter

let b:current_syntax = "sumi"
