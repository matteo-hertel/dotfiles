---@type LazySpec
return {
  "AstroNvim/astrocommunity",
  { import = "astrocommunity.pack.lua" },
  -- import/override with your plugins folder
  { import = "astrocommunity.recipes.disable-tabline" },
  { import = "astrocommunity.recipes.telescope-lsp-mappings" },
}
