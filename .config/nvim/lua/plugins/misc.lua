---
---@type LazySpec
return {
  {
    "ray-x/lsp_signature.nvim",
    event = "BufRead",
    config = function() require("lsp_signature").setup() end,
  },
  { "package-info.nvim", enabled = false },
  { "windwp/nvim-autopairs", enabled = false },
  { "windwp/nvim-ts-autotag", enabled = false },
  { "stevearc/resession.nvim", enabled = false },

  { "ralismark/opsort.vim" },
  { "tamton-aquib/duck.nvim" },
  {
    "princejoogie/dir-telescope.nvim",
    requires = { "nvim-telescope/telescope.nvim" },
    config = function()
      require("dir-telescope").setup {
        hidden = true,
      }
    end,
  },
  {
    "echasnovski/mini.nvim",
    branch = "stable",
    config = function()
      require("mini.ai").setup()
      require("mini.surround").setup()
    end,
  },
  {
    "cseickel/diagnostic-window.nvim",
    requires = { "MunifTanjim/nui.nvim" },
  },
  {
    "ThePrimeagen/harpoon",
    requires = { "nvim-lua/plenary.nvim" },
    config = function()
      require("harpoon").setup {}
      require("telescope").load_extension "harpoon"
    end,
  },
  {
    "folke/trouble.nvim",
    config = function() require("trouble").setup {} end,
  },
  {
    "windwp/nvim-spectre",
  },
  {
    "nvim-neo-tree/neo-tree.nvim",
    opts = {
      window = {
        width = 40,
      },
      filesystem = {
        filtered_items = {
          hide_dotfiles = false,
          hide_gitignored = false,
          hide_by_name = { ".DS_Store", "thumbs.db" },
        },
      },
    },
  },
  {
    "nvim-telescope/telescope.nvim",
    opts = {
      defaults = {
        file_ignore_patterns = { ".git" },
        pickers = {
          buffers = { sort_lastused = true },
        },
        mappings = {
          i = {
            ["<C-n>"] = require("telescope.actions").cycle_history_next,
            ["<C-p>"] = require("telescope.actions").cycle_history_prev,
          },
        },
      },
    },
  },
}
