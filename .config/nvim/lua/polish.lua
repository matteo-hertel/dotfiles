-- This will run last in the setup process and is a good place to configure
-- things like custom filetypes. This is just pure lua so anything that doesn't
-- fit in the normal config locations above can go here

-- Set up custom filetypes
vim.filetype.add {
  extension = {
    mdx = "markdown",
  },
  pattern = {
    [".*%.link"] = "sh",
  },
}

-- Suppress less important notifications
-- vim.opt.shortmess:append "sI" -- Suppress startup messages and intro

-- Add keymap to dismiss all messages/notifications
vim.keymap.set("n", "<Esc><Esc>", "<cmd>nohlsearch<CR><cmd>echo ''<CR>", {
  desc = "Dismiss notifications and clear search highlighting",
  silent = true,
})

-- Filter out annoying notifications from Lazy
-- local notify = vim.notify
-- vim.notify = function(msg, level, opts)
--   -- Filter out specific annoying messages
--   if type(msg) == "string" then
--     if msg:match "No specs found for module" then return end
--     if msg:match "copilot.*disabled" then return end
--   end
--   notify(msg, level, opts)
-- end
