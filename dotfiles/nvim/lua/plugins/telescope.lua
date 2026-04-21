-- plugins/telescope.lua:
return {
	"nvim-telescope/telescope.nvim",
	tag = "v0.2.2",
	-- or                              , branch = '0.1.x',
	dependencies = {
		"nvim-lua/plenary.nvim",
		"nvim-telescope/telescope-live-grep-args.nvim",
	},
}
