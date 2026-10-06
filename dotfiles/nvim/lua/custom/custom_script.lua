-- Config
local data_dir_path = vim.fn.stdpath("data") .. "/lico"
local config_file_path = data_dir_path .. "/config"

local function mkdir_data_dir_path()
	if vim.fn.isdirectory(data_dir_path) == 0 then
		vim.fn.mkdir(data_dir_path, "p")
	end
end
local function get_config()
	mkdir_data_dir_path()

	local fd = vim.uv.fs_open(config_file_path, "r", 438) -- 0666
	if not fd then
		return {}
	end
	local stat = vim.uv.fs_fstat(fd)
	local s_data = vim.uv.fs_read(fd, stat.size, 0)
	vim.uv.fs_close(fd)

	if s_data == nil or s_data == "" then
		return {}
	end

	local data = vim.json.decode(s_data)
	return data
end

local function set_config(key, value)
	mkdir_data_dir_path()

	local tmp_path = config_file_path .. "_tmp"
	local fd = vim.uv.fs_open(tmp_path, "w", 420) -- 0644
	if not fd then
		return false
	end

	local current_config = get_config()
	current_config[key] = value

	local s_config = vim.json.encode(current_config)

	vim.uv.fs_write(fd, s_config, 0)
	vim.uv.fs_close(fd)
	vim.uv.fs_rename(tmp_path, config_file_path)
	return true
end

-- Theme
--
local function apply_scheme(scheme)
	local config = get_config()
	set_config("scheme", scheme)
	vim.cmd("colorscheme " .. scheme)
end
local function init_scheme()
	local config = get_config()
	local scheme = config.scheme or "tokyonight-night"
	apply_scheme(scheme)
end
local function pick_scheme()
	local colors = vim.fn.getcompletion("", "color"), "\n"

	vim.ui.select(colors, {
		prompt = "Select Scheme:",
	}, function(choice)
		if choice and choice ~= "" then
			apply_scheme(choice)
		end
	end)
end
init_scheme()
vim.api.nvim_create_user_command("Scheme", pick_scheme, {})

-- Telescope Search hidden files

vim.keymap.set("n", "<leader>fa", function()
	local tb = require("telescope.builtin")

	tb.find_files({
		hidden = true,
		no_ignore = true,
		no_ignore_parent = true,
	})
end, { desc = "Find Files (Include Hidden/Ignore Files)" })

vim.keymap.set("n", "<leader>fn", function()
	local tb = require("telescope.builtin")

	tb.find_files({
		hidden = true,
	})
end, { desc = "Find Files (Include Hidden Files)" })

-- bring previous copy to current copy
vim.keymap.set("n", "yp", function()
	local last_yank = vim.fn.getreg("0")
	vim.fn.setreg("+", last_yank)
end, { desc = "Copy previous copied one" })
