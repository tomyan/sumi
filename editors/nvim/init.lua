-- Minimal nvim config for sumi preview editors.
-- Enables tree-sitter highlighting (if parsers are installed) with
-- fallback to our manual syntax files.

vim.o.termguicolors = true
vim.o.updatetime = 300
vim.o.autowriteall = true
vim.o.swapfile = false

-- Base colorscheme, then override with Tokyo Night-inspired TS highlights.
pcall(vim.cmd, 'colorscheme default')

local hl = vim.api.nvim_set_hl
hl(0, '@keyword',              { fg = '#bb9af7', bold = true })
hl(0, '@keyword.return',       { fg = '#bb9af7', bold = true })
hl(0, '@keyword.function',     { fg = '#bb9af7', bold = true })
hl(0, '@keyword.import',       { fg = '#bb9af7', bold = true })
hl(0, '@keyword.conditional',  { fg = '#bb9af7', bold = true })
hl(0, '@keyword.repeat',       { fg = '#bb9af7', bold = true })
hl(0, '@type',                 { fg = '#2ac3de', italic = true })
hl(0, '@type.builtin',         { fg = '#2ac3de', italic = true })
hl(0, '@type.definition',      { fg = '#2ac3de', bold = true })
hl(0, '@function',             { fg = '#7aa2f7', bold = true })
hl(0, '@function.call',        { fg = '#7aa2f7' })
hl(0, '@function.method.call', { fg = '#7aa2f7' })
hl(0, '@function.builtin',     { fg = '#e0af68' })
hl(0, '@variable',             { fg = '#c0caf5' })
hl(0, '@variable.parameter',   { fg = '#e0af68', italic = true })
hl(0, '@variable.member',      { fg = '#73daca' })
hl(0, '@property',             { fg = '#73daca' })
hl(0, '@constant',             { fg = '#ff9e64' })
hl(0, '@constant.builtin',     { fg = '#ff9e64' })
hl(0, '@boolean',              { fg = '#ff9e64' })
hl(0, '@number',               { fg = '#ff9e64' })
hl(0, '@string',               { fg = '#9ece6a' })
hl(0, '@string.escape',        { fg = '#73daca' })
hl(0, '@comment',              { fg = '#565f89', italic = true })
hl(0, '@operator',             { fg = '#89ddff' })
hl(0, '@punctuation.bracket',  { fg = '#545c7e' })
hl(0, '@punctuation.delimiter',{ fg = '#545c7e' })
hl(0, '@module',               { fg = '#7aa2f7' })
hl(0, '@label',                { fg = '#73daca' })

-- Add sumi syntax files to runtimepath.
local script_dir = vim.fn.fnamemodify(debug.getinfo(1, 'S').source:sub(2), ':h')
vim.opt.runtimepath:append(script_dir)

-- Enable syntax and filetype detection (for .sumi and .snapshot files).
vim.cmd('syntax on')
vim.cmd('filetype on')

-- Enable tree-sitter highlighting for supported filetypes.
-- BufReadPost fires after the file is loaded; start tree-sitter explicitly.
vim.api.nvim_create_autocmd({'BufReadPost', 'BufNewFile'}, {
  callback = function()
    local ok, _ = pcall(vim.treesitter.start)
    if ok then
      -- Disable regex syntax when tree-sitter is active (avoids conflicts).
      vim.bo.syntax = ''
    end
  end,
})
