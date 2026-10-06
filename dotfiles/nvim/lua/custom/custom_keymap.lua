-- plugins

-- NvimTree
vim.api.nvim_set_keymap("n", "<leader>e", ":NvimTreeToggle<CR>", { noremap = true, silent = true })

-- Oil
vim.api.nvim_set_keymap("n", "<leader>o", ":Oil<CR>", { noremap = true, silent = true })

-- Diffview
vim.api.nvim_set_keymap("n", "<leader>dvo", ":DiffviewOpen<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>dvc", ":DiffviewClose<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>dvm", ":DiffviewOpen origin/master<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>dvn", ":DiffviewOpen origin/main<CR>", { noremap = true, silent = true })

-- Neogit
vim.api.nvim_set_keymap("n", "<leader>ng", ":Neogit<CR>", { noremap = true, silent = true })


-- lsp
vim.api.nvim_set_keymap("n", "<leader>fs", ":Telescope lsp_document_symbols<CR>", { noremap = true, silent = true })

-- Custom
vim.api.nvim_set_keymap("n", "<Esc>", ":noh<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>mm", ":e ~/memo<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>cp", ":cprev<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>cn", ":cn<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap(
  "n",
  "<leader>ca",
  ":lua vim.fn.setreg('+', vim.fn.fnamemodify(vim.fn.expand('%:p'),':.'))<CR>",
  { noremap = true, silent = true }
)
vim.keymap.set("t", "<C-q>", "<C-\\><C-n>", { silent = true }) -- to exit from the terminal

vim.keymap.set("n", "<leader>x", "<cmd>.lua<cr>", { desc = "Execute the current line" })
vim.keymap.set("n", "<leader><leader>x", "<cmd>source %<cr>", { desc = "Execute the current file" })

vim.keymap.set("n", "<leader>de", function()
  vim.diagnostic.open_float()
end, { noremap = true, silent = true })

-- upscope plugin commands
vim.api.nvim_set_keymap("n", "<leader>fcb", ":LicovimLiveGrepWithClipboard<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>fcf", ":LicovimLiveGrepWithCurrentPath<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>upr", ":LicovimTestRunnerRun<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>upd", ":LicovimTestRunnerRunDebug<CR>", { noremap = true, silent = true })
vim.api.nvim_set_keymap("n", "<leader>upt", ":LicovimTestRunnerToggle<CR>", { noremap = true, silent = true })

-- open selected text as URL in default browser
vim.keymap.set("x", "O", function()
  local vstart = vim.fn.getpos("v")
  local vend = vim.fn.getpos(".")
  local line_start = vstart[2]
  local line_end = vend[2]
  if line_start > line_end then
    line_start, line_end = line_end, line_start
  end
  local lines = vim.api.nvim_buf_get_lines(0, line_start - 1, line_end, false)
  local text = vim.fn.trim(table.concat(lines, ""))
  if text ~= "" then
    vim.fn.jobstart({ "open", text }, { detach = true })
  end
end, { noremap = true, silent = true, desc = "Open selected text as URL in browser" })

-- disable close window shortcuts, accidently pressing those button is so annoying
vim.keymap.set("n", "<C-w>c", "<Nop>", { noremap = true })
vim.keymap.set("n", "<C-w>o", "<Nop>", { noremap = true })
vim.keymap.set("n", "<C-w>z", "<Nop>", { noremap = true })
vim.keymap.set("n", "<C-w>q", "<Nop>", { noremap = true })
