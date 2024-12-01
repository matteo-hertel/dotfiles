---@type LazySpec
return {
  "AstroNvim/astrolsp",
  ---@type AstroLSPOpts
  opts = {
    -- Configuration table of features provided by AstroLSP
    features = {
      codelens = false, -- enable/disable codelens refresh on start
      inlay_hints = true, -- enable/disable inlay hints on start
      semantic_tokens = true, -- enable/disable semantic token highlighting
    },
    -- customize lsp formatting options
    formatting = {
      format_on_save = {
        enabled = true, -- enable or disable format on save globally
      },
      timeout_ms = 1000, -- default format timeout
      -- mappings to be set up on attaching of a language server
      mappings = {
        n = {
          ["K"] = false,
          ["gi"] = {
            vim.lsp.buf.hover,
            desc = "Hover simbol details",
          },
        },
      },
    },
  },
}
