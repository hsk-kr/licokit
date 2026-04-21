return {
	"nvim-treesitter/nvim-treesitter",
	branch = "main",
	lazy = false,
	build = ":TSUpdate",
	config = function()
		-- Compat shim: some plugins (e.g. telescope 0.1.x) still call the
		-- old nvim-treesitter parsers.ft_to_lang(ft) which was removed in
		-- the main-branch rewrite. Route it to the native API.
		local parsers_ok, parsers = pcall(require, "nvim-treesitter.parsers")
		if parsers_ok and not parsers.ft_to_lang then
			parsers.ft_to_lang = function(ft)
				return vim.treesitter.language.get_lang(ft) or ft
			end
		end

		require("nvim-treesitter").install({
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
		})

		vim.api.nvim_create_autocmd("FileType", {
			callback = function(args)
				pcall(vim.treesitter.start, args.buf)
				vim.bo[args.buf].indentexpr = "v:lua.require'nvim-treesitter'.indentexpr()"
			end,
		})
	end,
}
