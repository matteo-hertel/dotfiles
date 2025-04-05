return {
  "JellyApple102/flote.nvim",
  version = "*", -- Use for stability; omit to use `main` branch for the latest features
  event = "VeryLazy",
  config = function()
    require("flote").setup {
      notes_dir = "~/Documents/Main",
      files = {
        global = "To Do.md",
      },
    }
  end,
}
