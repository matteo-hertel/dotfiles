-- Customize Treesitter

---@type LazySpec
return {
  "nvim-treesitter/nvim-treesitter",
  opts = {
    ensure_installed = {
      -- Core
      "lua",
      "luadoc",
      "vim",
      "vimdoc",

      -- TypeScript/JavaScript
      "typescript",
      "tsx",
      "javascript",
      "jsdoc",

      -- Go
      "go",
      "gomod",
      "gosum",
      "gowork",

      -- Web/Markup
      "html",
      "css",
      "json",
      "jsonc",
      "yaml",
      "toml",

      -- Documentation
      "markdown",
      "markdown_inline",

      -- Other
      "bash",
      "python",
      "regex",
      "gitignore",
      "git_config",
      "git_rebase",
      "gitcommit",
      "gitattributes",
    },
    highlight = {
      enable = true,
      additional_vim_regex_highlighting = false,
    },
    indent = {
      enable = true,
    },
  },
}
