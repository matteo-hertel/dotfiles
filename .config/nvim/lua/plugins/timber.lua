return {
  "Goose97/timber.nvim",
  version = "*", -- Use for stability; omit to use `main` branch for the latest features
  event = "VeryLazy",
  config = function()
    require("timber").setup {
      -- Configuration here, or leave empty to use defaults
      log_templates = {
        default = {
          javascript = [[console.log("%log_marker - %line_number: ", "%log_target", %log_target)]],
          typescript = [[console.log("%log_marker - %line_number: ", "%log_target", %log_target)]],
          jsx = [[console.log("%log_marker - %line_number: ","%log_target", %log_target)]],
          tsx = [[console.log("%log_marker - %line_number: ","%log_target", %log_target)]],
        },
        plain = {
          javascript = [[console.log("%log_marker - %line_number: ","%insert_cursor")]],
          typescript = [[console.log("%log_marker - %line_number: ","%insert_cursor")]],
          jsx = [[console.log("%log_marker - %line_number: ","%insert_cursor")]],
          tsx = [[console.log("%log_marker - %line_number: ","%insert_cursor")]],
        },
      },
      batch_log_templates = {
        default = {
          javascript = [[console.log("%log_marker - %line_number: ",{ %repeat<"%log_target": %log_target><, > })]],
          typescript = [[console.log("%log_marker - %line_number: ",{ %repeat<"%log_target": %log_target><, > })]],
          jsx = [[console.log("%log_marker - %line_number: ",{ %repeat<"%log_target": %log_target><, > })]],
          tsx = [[console.log("%log_marker - %line_number: ",{ %repeat<"%log_target": %log_target><, > })]],
        },
      },
    }
  end,
}
