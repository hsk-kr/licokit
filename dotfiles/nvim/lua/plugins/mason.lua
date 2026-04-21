return {
	"mason-org/mason.nvim",
	opts = {
		ui = {
			icons = {
				package_installed = "✓",
				package_pending = "➜",
				package_uninstalled = "✗",
			},
		},
		max_concurrent_installers = 10,
		ensure_installed = {
			"lua-language-server",
			"css-lsp",
			"html-lsp",
			"typescript-language-server",
			"prettier",
			"prettierd",
			"tailwindcss-language-server",
			"yaml-language-server",
			"delve",
			"gopls",
		},
	},
	config = true,
}
