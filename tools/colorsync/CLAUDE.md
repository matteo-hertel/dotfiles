# colorsync

A Go CLI tool that syncs color schemes across neovim, tmux, and iTerm2.

## Quick Reference

```bash
# Build
cd tools/colorsync && go build -o colorsync .
cd tools/colorsync-ai && swift build  # Build AI helper (macOS 26+, Apple Intelligence)

# Run tests
cd tools/colorsync && go test ./... -v

# Commands
./colorsync list                          # List available themes
./colorsync import catppuccin-mocha       # Import a built-in theme
./colorsync import ~/Downloads/theme.itermcolors  # Import from file
./colorsync generate                      # Create theme from bg/fg/accent
./colorsync ai-generate "warm autumn"     # AI-generate from description (macOS 26+)
./colorsync preview catppuccin-mocha      # Preview in terminal
./colorsync apply catppuccin-mocha        # Apply to all targets
./colorsync apply gruvbox-dark --target tmux,nvim  # Apply to specific targets
```

## Architecture

- **Palette model**: `palette/palette.go` - `Theme` struct with 18 colors (bg, fg, cursor, 16 ANSI). JSON serialization. Themes stored in `~/.config/colorsync/themes/`.
- **Color generation**: `palette/generate.go` - Derives 16 ANSI colors from bg/fg/accent using HSL manipulation.
- **Importers**: `importer/` - `builtin.go` has 6 hardcoded themes. `itermcolors.go` parses Apple plist XML.
- **Exporters**: `exporter/` - `neovim.go` writes standalone Lua colorscheme. `tmux.go` writes theme.conf. `iterm.go` writes .itermcolors and sends live escape sequences.
- **Preview**: `preview/preview.go` - Renders color swatches using 24-bit ANSI escapes.
- **AI helper**: `../colorsync-ai/` - Swift CLI using Apple Foundation Models (on-device LLM) for AI theme generation. Go shells out to it via `cmd/ai_generate.go`.
- **CLI**: `cmd/` - Subcommands registered via `init()` functions. No cobra/viper, plain stdlib.

## Output Paths

| Target | Output | Activation |
|--------|--------|------------|
| neovim | `~/.config/nvim/colors/<name>.lua` | `:colorscheme <name>` |
| tmux | `~/.tmux/theme.conf` | `source-file ~/.tmux/theme.conf` in `.tmux.conf` |
| iTerm | `~/.config/colorsync/output/<name>.itermcolors` + live escape sequences | Automatic for current session |

## Conventions

- Go stdlib only, no external dependencies
- Theme names use hyphens (`catppuccin-mocha`), Lua filenames use underscores (`catppuccin_mocha.lua`)
- All colors stored as 7-char hex strings (`#rrggbb`)
- Tests live alongside source in `_test.go` files
- Built-in themes are hardcoded in `importer/builtin.go`

## Adding a New Built-in Theme

Edit `importer/builtin.go` and add an entry to the `builtins` map:

```go
"theme-name": {
    Name: "theme-name", Background: "#...", Foreground: "#...", Cursor: "#...",
    Colors: [16]string{ /* ansi 0-15 */ },
},
```

## Adding a New Export Target

1. Create `exporter/<target>.go` with an `Export<Target>(theme *palette.Theme, path string) error` function
2. Add a test in `exporter/<target>_test.go`
3. Wire into `cmd/apply.go` by adding a new target check in `runApply`
