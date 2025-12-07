-- Customize Mason

---@type LazySpec
return {
  -- use mason-tool-installer for automatically installing Mason packages
  {
    "WhoIsSethDaniel/mason-tool-installer.nvim",
    -- overrides `require("mason-tool-installer").setup(...)`
    opts = {
      -- Make sure to use the names found in `:Mason`
      ensure_installed = {
        -- Lua support
        "lua-language-server", -- LSP
        "stylua", -- formatter

        -- TypeScript/JavaScript support
        "typescript-language-server", -- LSP
        "eslint-lsp", -- linting
        "prettier", -- formatter

        -- Go support
        "gopls", -- LSP
        "gofumpt", -- formatter (stricter than gofmt)
        "goimports", -- import management
        "gomodifytags", -- struct tag management
        "impl", -- interface implementation generator
        "delve", -- debugger

        -- Additional useful tools
        "json-lsp", -- JSON LSP
        "tree-sitter-cli", -- Treesitter CLI
      },
    },
  },
}
