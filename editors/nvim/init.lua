-- Minimal nvim config for sumi preview editors.
-- Enables tree-sitter highlighting (if parsers are installed) with
-- fallback to our manual syntax files.

vim.o.termguicolors = true
vim.o.updatetime = 300
vim.o.autowriteall = true
vim.o.swapfile = false

-- Try to load a colorscheme that works well with tree-sitter.
pcall(vim.cmd, 'colorscheme default')

-- Add sumi syntax files to runtimepath.
local script_dir = vim.fn.fnamemodify(debug.getinfo(1, 'S').source:sub(2), ':h')
vim.opt.runtimepath:append(script_dir)

-- Enable syntax and filetype detection (for .sumi and .snapshot files).
vim.cmd('syntax on')
vim.cmd('filetype on')

-- Try to enable tree-sitter highlighting for the current buffer.
-- This gives rich Go/CSS/etc highlighting if parsers are installed.
-- Falls back silently to regex syntax if not available.
vim.api.nvim_create_autocmd('FileType', {
  callback = function()
    pcall(vim.treesitter.start)
  end,
})
