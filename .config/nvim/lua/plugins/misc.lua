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
    dependencies = { "nvim-telescope/telescope.nvim" },
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
    dependencies = { "MunifTanjim/nui.nvim" },
  },
  {
    "ThePrimeagen/harpoon",
    dependencies = { "nvim-lua/plenary.nvim" },
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
    opts = function(_, opts)
      local actions = require "telescope.actions"
      return vim.tbl_deep_extend("force", opts, {
        defaults = {
          vimgrep_arguments = {
            "rg",
            "--color=never",
            "--no-heading",
            "--with-filename",
            "--line-number",
            "--column",
            "--smart-case",
            "--hidden", -- This flag tells ripgrep to search hidden files and directories
          },
          file_ignore_patterns = { ".git" },
          mappings = {
            i = {
              ["<C-n>"] = actions.cycle_history_next,
              ["<C-p>"] = actions.cycle_history_prev,
            },
          },
        },
        pickers = {
          buffers = { sort_lastused = true },
        },
      })
    end,
  },
}
