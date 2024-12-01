---@type LazySpec
return {
 			{
				"sam4llis/nvim-tundra",
				config = function()
					require("nvim-tundra").setup({
						plugins = {
							lsp = true,
							treesitter = true,
							nvimtree = true,
							cmp = true,
							context = true,
							dbui = true,
							gitsigns = true,
							telescope = true,
						},
					})
				end,
			},
}

