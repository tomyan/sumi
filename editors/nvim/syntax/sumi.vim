" Vim syntax file for sumi (.sumi) single-file components
" Language: Sumi (Go + Terminal CSS + HTML-like templates)

if exists("b:current_syntax")
  finish
endif

" ============================================================================
" Script block — embed full Go syntax
" ============================================================================

syntax include @goSyntax syntax/go.vim
unlet! b:current_syntax

syntax region sumiScript
      \ matchgroup=sumiSectionTag
      \ start="<script>"
      \ end="</script>"
      \ contains=@goSyntax,sumiSignalNew,sumiSignalFrom,sumiSignalEffect,sumiSignalBatch,sumiSignalMethod,sumiTuiCall,sumiAppCall
      \ keepend

" Signal creation — most important sumi-specific calls
syntax match sumiSignalNew    /signal\.New\>/     containedin=sumiScript
syntax match sumiSignalFrom   /signal\.From\>/    containedin=sumiScript
syntax match sumiSignalEffect /signal\.Effect\>/  containedin=sumiScript
syntax match sumiSignalBatch  /signal\.Batch\>/   containedin=sumiScript

" Signal methods on instances
syntax match sumiSignalMethod /\.\(Get\|Set\|Update\)(/ containedin=sumiScript contains=sumiMethodParen
syntax match sumiMethodParen  /(/ contained

" Framework calls
syntax match sumiTuiCall /tui\.\(Env\|Run\|Quit\|TestApp\|RunWithOptions\)\>/ containedin=sumiScript
syntax match sumiAppCall /app\.\(Quit\|Dirty\|Wake\)\>/ containedin=sumiScript

" ============================================================================
" Style block — terminal CSS with manual patterns
" (syntax include @cssSyntax doesn't reliably color inside regions)
" ============================================================================

syntax region sumiStyle
      \ matchgroup=sumiSectionTag
      \ start="<style>"
      \ end="</style>"
      \ contains=sumiCSSSelector,sumiCSSBraces,sumiCSSProperty,sumiCSSValue,sumiCSSComment,sumiCSSAtRule
      \ keepend

" CSS selectors: .class, element, :pseudo
syntax match  sumiCSSSelector /[.#:][a-zA-Z][a-zA-Z0-9_-]*/ contained containedin=sumiStyle
syntax match  sumiCSSSelector /\<\(root\|box\|text\)\>/ contained containedin=sumiStyle
syntax match  sumiCSSBraces   /[{}]/ contained containedin=sumiStyle

" CSS properties and values
syntax match  sumiCSSProperty /\<[a-z][a-z-]*\s*:/ contained containedin=sumiStyle contains=sumiCSSColon
syntax match  sumiCSSColon    /:/ contained
syntax match  sumiCSSValue    /:\s*\zs[^;}\n]*/ contained containedin=sumiStyle
syntax region sumiCSSComment  start="/\*" end="\*/" contained containedin=sumiStyle

" CSS at-rules
syntax match  sumiCSSAtRule   /@\(media\|container\|keyframes\)\>/ contained containedin=sumiStyle

" ============================================================================
" Template — tags, attributes, expressions, control flow
" Use \C in patterns for case-sensitive matching (CSS sets 'syn case ignore').
" ============================================================================

" --- Element tags ---
syntax match sumiTagOpen    /\C<\(box\|text\|title\)\>/ contains=sumiTagKeyword
syntax match sumiTagClose   /\C<\/\(box\|text\|title\)>/ contains=sumiTagKeyword
syntax match sumiTagKeyword /\C\(box\|text\|title\)/ contained
syntax match sumiSelfClose  /\/>/

" --- Component tags (PascalCase) ---
syntax match sumiCompOpen  /\C<[A-Z][A-Za-z]*/ contains=sumiCompName
syntax match sumiCompClose /\C<\/[A-Z][A-Za-z]*>/ contains=sumiCompName
syntax match sumiCompName  /\C[A-Z][A-Za-z]*/ contained

" --- Slot tags ---
syntax match sumiSlotOpen    /\C<slot:[a-z][a-z0-9-]*/ contains=sumiSlotPrefix,sumiSlotName
syntax match sumiSlotClose   /\C<\/slot:[a-z][a-z0-9-]*>/ contains=sumiSlotPrefix,sumiSlotName
syntax match sumiSlotDefault /\C<slot:default\s*\/>/
syntax match sumiSlotPrefix  /\Cslot:/ contained
syntax match sumiSlotName    /\C:[a-z][a-z0-9-]*/ contained


" --- Attributes ---
" class attribute (special — often used)
syntax match sumiClassAttr /\<class=/ nextgroup=sumiAttrStringVal
" onkey attribute (event handler)
syntax match sumiOnkeyAttr /\<onkey=/ nextgroup=sumiAttrStringVal
" bind: attributes
syntax match sumiBindAttr  /\<bind:[a-z][a-z0-9-]*=/ nextgroup=sumiAttrExprVal
" Generic attributes
syntax match sumiAttrName  /\<[a-z][a-z0-9-]*=\@=/ nextgroup=sumiAttrEq
syntax match sumiAttrEq    /=/ contained nextgroup=sumiAttrStringVal,sumiAttrExprVal

" Attribute values
syntax region sumiAttrStringVal start=/"/ end=/"/ contained oneline
syntax region sumiAttrExprVal   matchgroup=sumiExprBrace start=/{/ end=/}/ contained oneline

" --- Template expressions {expr} ---
syntax region sumiExpr matchgroup=sumiExprBrace start=/{/ end=/}/ oneline
      \ contains=sumiExprKeyword

" --- Control flow blocks ---
" Opening: {if condition}, {for clause}, {slot name}, {snippet name(...)}, {render name(...)}
syntax match sumiCtrlIf      /{if\>/
syntax match sumiCtrlElse    /{else}/
syntax match sumiCtrlFor     /{for\>/
syntax match sumiCtrlSlot    /{slot\>/
syntax match sumiCtrlSnippet /{snippet\>/
syntax match sumiCtrlRender  /{render\>/

" Closing: {/if}, {/for}, {/slot}, {/snippet}
syntax match sumiCtrlEnd     /{\(\/if\|\/for\|\/slot\|\/snippet\)}/

" --- HTML comments ---
syntax region sumiComment start=/<!--/ end=/-->/

" ============================================================================
" Highlight definitions
" ============================================================================

" Section tags: <script>, </script>, <style>, </style>
highlight default link sumiSectionTag    Keyword

" Signal API — make these stand out as the core reactive primitives
highlight default sumiSignalNew    guifg=#ff9e64 gui=bold ctermfg=215 cterm=bold
highlight default sumiSignalFrom   guifg=#ff9e64 gui=bold ctermfg=215 cterm=bold
highlight default sumiSignalEffect guifg=#ff9e64 gui=bold ctermfg=215 cterm=bold
highlight default sumiSignalBatch  guifg=#ff9e64 gui=bold ctermfg=215 cterm=bold
highlight default sumiSignalMethod guifg=#e0af68 ctermfg=179
highlight default sumiMethodParen  guifg=#a9b1d6 ctermfg=146

" Framework calls
highlight default sumiTuiCall guifg=#7aa2f7 gui=italic ctermfg=111 cterm=italic
highlight default sumiAppCall guifg=#7aa2f7 gui=italic ctermfg=111 cterm=italic

" Template element tags
highlight default sumiTagKeyword guifg=#f7768e ctermfg=204
highlight default sumiTagOpen    guifg=#545c7e ctermfg=60
highlight default sumiTagClose   guifg=#545c7e ctermfg=60
highlight default sumiSelfClose  guifg=#545c7e ctermfg=60

" Component tags — prominent
highlight default sumiCompName guifg=#2ac3de gui=bold ctermfg=44 cterm=bold
highlight default sumiCompOpen guifg=#545c7e ctermfg=60
highlight default sumiCompClose guifg=#545c7e ctermfg=60

" Slot tags — distinct from regular tags
highlight default sumiSlotOpen    guifg=#545c7e ctermfg=60
highlight default sumiSlotClose   guifg=#545c7e ctermfg=60
highlight default sumiSlotDefault guifg=#bb9af7 gui=italic ctermfg=141 cterm=italic
highlight default sumiSlotPrefix  guifg=#bb9af7 ctermfg=141
highlight default sumiSlotName    guifg=#bb9af7 gui=bold ctermfg=141 cterm=bold

" Attributes
highlight default sumiClassAttr     guifg=#73daca ctermfg=79
highlight default sumiOnkeyAttr     guifg=#e0af68 ctermfg=179
highlight default sumiBindAttr      guifg=#ff9e64 gui=italic ctermfg=215 cterm=italic
highlight default sumiAttrName      guifg=#73daca ctermfg=79
highlight default sumiAttrEq        guifg=#545c7e ctermfg=60
highlight default sumiAttrStringVal guifg=#9ece6a ctermfg=149
highlight default sumiAttrExprVal   guifg=#e0af68 ctermfg=179

" Template expressions
highlight default sumiExpr      guifg=#e0af68 ctermfg=179
highlight default sumiExprBrace guifg=#545c7e ctermfg=60

" Control flow — highly visible
highlight default sumiCtrlIf      guifg=#bb9af7 gui=bold ctermfg=141 cterm=bold
highlight default sumiCtrlElse    guifg=#bb9af7 gui=bold ctermfg=141 cterm=bold
highlight default sumiCtrlFor     guifg=#bb9af7 gui=bold ctermfg=141 cterm=bold
highlight default sumiCtrlSlot    guifg=#7dcfff gui=bold ctermfg=117 cterm=bold
highlight default sumiCtrlSnippet guifg=#7dcfff gui=bold ctermfg=117 cterm=bold
highlight default sumiCtrlRender  guifg=#7dcfff gui=bold ctermfg=117 cterm=bold
highlight default sumiCtrlEnd    guifg=#bb9af7 ctermfg=141

" CSS in style block
highlight default sumiCSSSelector guifg=#73daca gui=bold ctermfg=79 cterm=bold
highlight default sumiCSSBraces   guifg=#545c7e ctermfg=60
highlight default sumiCSSProperty guifg=#7aa2f7 ctermfg=111
highlight default sumiCSSColon    guifg=#545c7e ctermfg=60
highlight default sumiCSSValue    guifg=#9ece6a ctermfg=149
highlight default sumiCSSAtRule   guifg=#bb9af7 gui=bold ctermfg=141 cterm=bold
highlight default link sumiCSSComment Comment

" HTML Comments
highlight default link sumiComment Comment

let b:current_syntax = "sumi"
