---@type LazySpec
return {
  "echasnovski/mini.icons",
  lazy = false,
  priority = 1000,
  opts = {},
  specs = {
    { "nvim-tree/nvim-web-devicons", enabled = false, optional = true },
  },
  config = function(_, opts)
    local mini_icons = require("mini.icons")
    mini_icons.setup(opts)
    mini_icons.mock_nvim_web_devicons()
  end,
}
