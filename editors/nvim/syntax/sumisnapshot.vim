" Vim syntax file for sumi snapshot files (.snapshot)

if exists("b:current_syntax")
  finish
endif

" Frame markers: === Frame: name ===
syntax match snapshotFrameMarker /^=== Frame: .* ===$/ contains=snapshotFrameName
syntax match snapshotFrameName /Frame: \zs[^=]*\ze ===/ contained

" ANSI SGR styling markers (styled text in snapshots)
syntax match snapshotSGR /\e\[[0-9;]*m/

" Box-drawing characters
syntax match snapshotBorder /[┌┐└┘─│┬┴├┤┼╭╮╰╯]/

highlight default snapshotFrameMarker guifg=#7dcfff gui=bold ctermfg=117 cterm=bold
highlight default snapshotFrameName   guifg=#ff9e64 gui=bold ctermfg=215 cterm=bold
highlight default snapshotSGR         guifg=#565f89 ctermfg=60
highlight default snapshotBorder      guifg=#545c7e ctermfg=60

let b:current_syntax = "sumisnapshot"
