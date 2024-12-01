---@type LazySpec
return {
  "AstroNvim/astrocore",
  ---@type AstroCoreOpts
  opts = {
    features = {
      autopairs = false,
    },
    diagnostics = {
      virtual_text = true,
      underline = true,
    },
    options = {
      opt = {
        clipboard = "",
        mouse = "",
        linebreak = true,
        list = true,
        scrolloff = 999,
        colorcolumn = "80",
        cursorline = true,
        listchars = {
          extends = "⟩",
          nbsp = "␣",
          precedes = "⟨",
          tab = "│→",
          trail = "·",
        },
        guicursor = "",
        showbreak = "↪ ",
        wrap = true,
      },
    },
  },
}
