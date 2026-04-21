return {
	"nvim-treesitter/nvim-treesitter",
	lazy = false,
	build = ":TSUpdate",
	config = function()
		require("nvim-treesitter.configs").setup({
			ensure_installed = {
				"css",
				"dockerfile",
				"gitcommit",
				"go",
				"html",
				"javascript",
				"json",
				"lua",
				"make",
				"markdown",
				"markdown_inline",
				"python",
				"toml",
				"tsx",
				"typescript",
				"vim",
				"yaml",
			},
			highlight = {
				enable = true,
			},
			indent = {
				enable = true,
			},
			endwise = {
				enable = true,
			},
			sync_install = false,
			auto_install = true,
			incremental_selection = {
				enable = true,
				keymaps = {
					-- init_selection = "gnn", -- set to `false` to disable one of the mappings
					-- node_incremental = "grn",
					-- scope_incremental = "grc",
					-- node_decremental = "grm",
				},
			},
		})
	end,
}
