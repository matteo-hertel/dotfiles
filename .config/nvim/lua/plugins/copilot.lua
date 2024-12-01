---@type LazySpec
return {
  "zbirenbaum/copilot.lua",
  cmd = "Copilot",
  event = "InsertEnter",
  config = function()
    require("copilot").setup {
      suggestion = {
        enabled = true,
        auto_trigger = true,
        debounce = 75,
        keymap = {
          accept = "<C-CR>",
          dismiss = "<C-x>",
          next = "<C-n>",
        },
      },
      panel = { enabled = false, auto_trigger = false },
    }
  end,
}
