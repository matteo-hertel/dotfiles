---@type LazySpec
return {
  "AstroNvim/astrocore",
  ---@type AstroCoreOpts
  opts = {
    mappings = {
      n = {
        ["<leader>a"] = {
          "<cmd>lua require('telescope.builtin').live_grep()<CR>",
          desc = "Find Words",
        },
        ["<leader>fd"] = {
          "<cmd>lua require('telescope').extensions.dir.find_files()<CR>",
          desc = "Find file in directory",
        },
        ["<leader>fa"] = {
          "<cmd>lua require('telescope').extensions.dir.live_grep()<CR>",
          desc = "Grep file in directory",
        },
        ["<leader>dn"] = {
          "<cmd>lua require('duck').hatch()<CR>",
          desc = "New duck",
        },
        ["<leader>dk"] = {
          "<cmd>lua require('duck').cook()<CR>",
          desc = "Cook duck",
        },
        [";"] = {
          "<cmd>lua require('telescope.builtin').buffers()<CR>",
          desc = "Find Buffers",
        },
        ["<S-k>"] = { "6k", desc = "Move 6 lines up" },
        ["<S-j>"] = { "6j", desc = "Move 6 lines down" },
        ["<S-h>"] = { "6h", desc = "Move 6 char left" },
        ["<S-l>"] = { "6l", desc = "Move 6 char right" },
        ["<C-h>"] = {
          "<cmd>lua require('smart-splits').move_cursor_left()<CR>",
          desc = "Move to tmux pane Left",
        },
        ["<C-j>"] = {
          "<cmd>lua require('smart-splits').move_cursor_down()<CR>",
          desc = "Move to tmux pane Down",
        },
        ["<C-k>"] = {
          "<cmd>lua require('smart-splits').move_cursor_up()<CR>",
          desc = "Move to tmux pane Up",
        },
        ["<C-l>"] = {
          "<cmd>lua require('smart-splits').move_cursor_right()<CR>",
          desc = "Move to tmux pane Right",
        },
        ["<leader>o"] = {
          "<cmd>lua require('before').jump_to_last_edit()<CR>",
          desc = "Jump to last edit",
        },
        ["<leader>i"] = {
          "<cmd>lua require('before').jump_to_next_edit()<CR>",
          desc = "Jump to last edit",
        },
        ["gc"] = {
          function() require("Comment.api").toggle.linewise.current() end,
          desc = "Comment line",
        },
        ["<leader>lt"] = {
          "<cmd>TroubleToggle<cr>",
          desc = "Toggle diagnostics",
        },
        ["<leader>lq"] = {
          "<cmd>TroubleToggle quickfix<cr>",
          desc = "Toggle quickfix",
        },
        ["<leader>ta"] = {
          "<cmd>AerialToggle<cr>",
          desc = "Toggle aerial",
        },
        ["<leader>S"] = {
          "<cmd>lua require('spectre').open()<CR>",
          desc = "Open Spectre",
        },
        ["<leader>sw"] = {
          "<cmd>lua require('spectre').open_visual({select_word=true})<CR>",
          desc = "Specre on current word",
        },
        ["<F1>"] = "<esc>",
        -- Harpoon
        ["<leader>ha"] = {
          "<cmd>lua require('harpoon.mark').add_file()<CR>",
          desc = "Add file to Harpoon",
        },
        ["<leader>hh"] = {
          "<cmd>lua require('harpoon.ui').toggle_quick_menu()<CR>",
          desc = "Show Harpoon menu",
        },
        ["<C-n>"] = {
          "<cmd>lua require('harpoon.ui').nav_next()<CR>",
          desc = "Go to next Harpoon mark",
        },
        ["<C-p>"] = {
          "<cmd>lua require('harpoon.ui').nav_prev()<CR>",
          desc = "Go to prev Harpoon mark",
        },
      },
      x = {
        ["<leader>p"] = {
          '"_dP',
          desc = "Paste without overwriting clipboard",
        },
      },
      v = {
        ["J"] = { ":m '>+1<CR>gv=gv", desc = "Move line down" },
        ["K"] = { ":m '<-2<CR>gv=gv", desc = "Move line up" },
        ["<leader>y"] = { '"+y', desc = "Copy to system clipboard" },
      },
    },
  },
}
