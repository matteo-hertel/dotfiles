# Dotfiles

Years of fine tuning to achieve the perfect dev environment.

## Prerequisites

- [Homebrew](https://brew.sh/) (installed automatically by setup.sh if missing)
- [oh-my-zsh](https://ohmyz.sh/)
- [powerlevel10k](https://github.com/romkatv/powerlevel10k)
- [Nerd Fonts](https://nerdfonts.com/)
- [iTerm2](https://www.iterm2.com/) and/or [Ghostty](https://ghostty.org/) — `colorsync` exports themes for both.

## Installation

```bash
git clone git@github.com:matteo-hertel/dotfiles.git ~/mhdev/dotfiles
cd ~/mhdev/dotfiles
./setup.sh
```

This will:
- Install `stow` and `go` via Homebrew
- Symlink all dotfiles into `$HOME` using GNU Stow
- Link Claude Code and Codex config files
- Build and install `colorsync` to `~/.local/bin/`

After setup, create `~/.env/env.custom.sh` for machine-specific config (API keys, workspace aliases). This file is gitignored.

## How it works

The repo mirrors your home directory structure. [GNU Stow](https://www.gnu.org/software/stow/) creates symlinks from `$HOME` pointing into this repo. Edits to dotfiles show up as uncommitted changes in git.

## Maintenance

### Refresh stow links

Run Stow from the repo root and point the target at your home directory:

```bash
cd ~/mhdev/dotfiles
stow --no --verbose=1 --target="$HOME" .
stow --verbose=2 --target="$HOME" .
```

The package is `.` because this repo mirrors `$HOME` directly. `.stow-local-ignore` keeps repo-only paths such as `tools/`, `docs/`, `Readme.md`, `AGENTS.md`, `.claude/`, and `.codex/` out of Stow.

If Stow reports that an existing target is "not owned by stow", check it before removing anything:

```bash
readlink ~/.env/env.custom.sh
realpath ~/.env/env.custom.sh
```

Only remove and recreate a link if it already resolves to the matching file in `~/mhdev/dotfiles`. Do not replace real files in `$HOME` without first copying their contents into the repo.

`~/.env/env.custom.sh` is the machine-local secrets file for API keys and aliases. It is gitignored by `.gitignore`, but Stow still links it into place when the file exists locally.

Claude and Codex config are linked separately by `setup.sh` because `.claude/` and `.codex/` are excluded from Stow to avoid folding the whole live agent directories into this repo.

### Agent guidance

Claude guidance lives in `.claude/`. Codex guidance lives in `.codex/AGENTS.md`, with a repo-root `AGENTS.md` entrypoint for agents working inside this repo. Keep both sets of guidance in sync when updating working preferences or skills.

### Rebuild local binaries

`setup.sh` builds `colorsync` into `~/.local/bin/colorsync`. If the source has changed and the installed command looks stale, rebuild it from the tool directory:

```bash
cd ~/mhdev/dotfiles/tools/colorsync
go test ./...
go build -o "$HOME/.local/bin/colorsync" .
colorsync current
```

`colorsync current` shows the theme currently detected for each target.

## Tools

### colorsync

CLI tool to sync color schemes across neovim, tmux, iTerm2, Ghostty, and powerlevel10k. Lives in `tools/colorsync/`. See its [CLAUDE.md](tools/colorsync/CLAUDE.md) for details.

```bash
colorsync list
colorsync apply catppuccin-mocha
colorsync current
colorsync generate
colorsync ai-generate "warm autumn theme"
```

Generated themes are stored in `.config/colorsync/themes/` and tracked by git.
