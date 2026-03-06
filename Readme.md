# Dotfiles

Years of fine tuning to achieve the perfect dev environment.

## Prerequisites

- [Homebrew](https://brew.sh/) (installed automatically by setup.sh if missing)
- [oh-my-zsh](https://ohmyz.sh/)
- [powerlevel10k](https://github.com/romkatv/powerlevel10k)
- [Nerd Fonts](https://nerdfonts.com/)
- [iTerm2](https://www.iterm2.com/)

## Installation

```bash
git clone git@github.com:matteo-hertel/dotfiles.git ~/mhdev/dotfiles
cd ~/mhdev/dotfiles
./setup.sh
```

This will:
- Install `stow` and `go` via Homebrew
- Symlink all dotfiles into `$HOME` using GNU Stow
- Link Claude Code config files
- Build and install `colorsync` to `~/.local/bin/`

After setup, create `~/.env/env.custom.sh` for machine-specific config (API keys, workspace aliases). This file is gitignored.

## How it works

The repo mirrors your home directory structure. [GNU Stow](https://www.gnu.org/software/stow/) creates symlinks from `$HOME` pointing into this repo. Edits to dotfiles show up as uncommitted changes in git.

## Tools

### colorsync

CLI tool to sync color schemes across neovim, tmux, iTerm2, and powerlevel10k. Lives in `tools/colorsync/`. See its [CLAUDE.md](tools/colorsync/CLAUDE.md) for details.

```bash
colorsync list
colorsync apply catppuccin-mocha
colorsync generate
colorsync ai-generate "warm autumn theme"
```

Generated themes are stored in `.config/colorsync/themes/` and tracked by git.
