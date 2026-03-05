# colorsync - Unified Color Scheme Tool

## Overview

A Go CLI that imports, generates, previews, and applies color schemes across neovim, tmux, and iTerm2.

## Core Data Model

A unified palette with 18 named colors:

- `background`, `foreground`, `cursor`
- `color0`-`color7` (standard ANSI)
- `color8`-`color15` (bright ANSI)

Stored as hex strings (`#1e1e2e`). All import/export targets map to/from this palette.

Theme files stored as JSON in `~/.config/colorsync/themes/`.

```json
{
  "name": "catppuccin-mocha",
  "background": "#1e1e2e",
  "foreground": "#cdd6f4",
  "cursor": "#f5e0dc",
  "colors": ["#45475a", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#bac2de",
              "#585b70", "#f38ba8", "#a6e3a1", "#f9e2af", "#89b4fa", "#f5c2e7", "#94e2d5", "#a6adc8"]
}
```

## Commands

| Command | Description |
|---|---|
| `colorsync list` | List built-in and saved themes |
| `colorsync import <file.itermcolors>` | Import from iTerm colors file, save to local theme store |
| `colorsync import <name>` | Import a built-in theme (catppuccin-mocha, gruvbox-dark, tokyo-night, nord, etc.) |
| `colorsync generate` | Interactive: pick bg, fg, accent; tool derives the full 16-color palette |
| `colorsync preview [theme]` | Print color swatches in terminal |
| `colorsync apply [theme]` | Preview swatches, confirm, then write to all three targets |
| `colorsync apply [theme] --target tmux,nvim` | Apply to specific targets only |

## Import Sources

### iTerm2 `.itermcolors` XML files
Parse the Apple plist XML format. Maps `Ansi X Color`, `Background Color`, `Foreground Color`, `Cursor Color` keys to the unified palette.

### Built-in themes by name
Ship with ~6 popular palettes hardcoded: `catppuccin-mocha`, `catppuccin-latte`, `gruvbox-dark`, `gruvbox-light`, `tokyo-night`, `nord`.

## Output Targets

### Neovim -> `~/.config/nvim/colors/<theme>.lua`
Standalone Lua file that sets `vim.o.background`, `vim.cmd("hi clear")`, and defines highlight groups (Normal, CursorLine, StatusLine, Comment, String, Function, Keyword, etc.) using the palette. No plugin dependency.

### tmux -> `~/.tmux/theme.conf`
Generated conf file with status bar, pane border, message, and mode colors mapped from the palette. The user's `.tmux.conf.link` gets a one-time `source-file ~/.tmux/theme.conf` line added.

### iTerm2
- Writes `.itermcolors` XML file to `~/.config/colorsync/output/`
- Sends proprietary escape sequences to live-update the running terminal colors

## Generate Flow

```
$ colorsync generate
Background (#hex): #1a1b26
Foreground (#hex): #c0caf5
Accent (#hex): #7aa2f7
Name: my-custom-theme
-> Derives 16 ANSI colors by adjusting hue/saturation/lightness from the three inputs
-> Saves to ~/.config/colorsync/themes/my-custom-theme.json
-> Preview? [y/n]
-> Apply? [y/n]
```

## Preview

Print colored blocks and sample text in terminal showing all 16 colors, plus bg/fg/cursor. Prompt "apply? [y/n]" before writing.

## Project Structure

```
tools/colorsync/
  main.go
  go.mod
  CLAUDE.md
  palette/
    palette.go        # Theme struct, load/save JSON
    generate.go       # Derive 16 colors from bg/fg/accent
  importer/
    itermcolors.go    # Parse .itermcolors XML
    builtin.go        # Built-in theme definitions
  exporter/
    neovim.go         # Generate Lua colorscheme
    tmux.go           # Generate tmux theme.conf
    iterm.go          # Generate .itermcolors + escape sequences
  preview/
    preview.go        # Terminal color swatches
  cmd/
    root.go
    list.go
    import.go
    generate.go
    preview.go
    apply.go
```

Lives in `tools/colorsync/` within the dotfiles repo. No external dependencies beyond Go stdlib.

## Decisions

- Go with stdlib only (no cobra, no third-party color libs)
- Base16-style 16-color palette as the universal interchange format
- Standalone neovim Lua colorscheme (no plugin dependency)
- Separate tmux theme file sourced from main conf
- iTerm live escape sequences + .itermcolors file output
- CLAUDE.md included for AI-assisted development context
