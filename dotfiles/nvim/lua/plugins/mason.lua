return {
	"mason-org/mason.nvim",
	dependencies = {
		"WhoIsSethDaniel/mason-tool-installer.nvim",
	},
	opts = {
		ui = {
			icons = {
				package_installed = "✓",
				package_pending = "➜",
				package_uninstalled = "✗",
			},
		},
		max_concurrent_installers = 10,
	},
	config = function(_, opts)
		require("mason").setup(opts)
		require("mason-tool-installer").setup({
			ensure_installed = {
				"bash-language-server",
				"css-lsp",
				"delve",
				"gopls",
				"harper-ls",
				"html-lsp",
				"json-lsp",
				"lua-language-server",
				"prettier",
				"prettierd",
				"pyright",
				"python-lsp-server",
				"tailwindcss-language-server",
				"typescript-language-server",
				"yaml-language-server",
			},
			run_on_start = true,
		})
	end,
}
