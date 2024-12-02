return {
  "folke/snacks.nvim",
  keys = {
    {
      "<Leader>.",
      function() require("snacks").scratch() end,
      desc = "Toggle Scratch Buffer",
    },
    { "<Leader>s", function() require("snacks").scratch.select() end, desc = "Select Scratch Buffer" },
  },
}
