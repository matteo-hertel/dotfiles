#!/bin/bash
set -e

DOTFILES_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Setting up dotfiles from $DOTFILES_DIR"

# 1. Homebrew
if ! command -v brew &>/dev/null; then
    echo "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    eval "$(/opt/homebrew/bin/brew shellenv)"
fi

echo "Installing packages..."
brew install stow go

# 2. Stow dotfiles
echo "Linking dotfiles with stow..."
cd "$DOTFILES_DIR"
stow --target="$HOME" .

# 3. Claude config (separate from stow to avoid folding ~/.claude)
# CLAUDE.md is per-machine — create a default one if missing
echo "Linking Claude config..."
mkdir -p "$HOME/.claude"
ln -sf "$DOTFILES_DIR/.claude/CLAUDE.shared.md" "$HOME/.claude/CLAUDE.shared.md"
ln -sf "$DOTFILES_DIR/.claude/CLAUDE.work.md" "$HOME/.claude/CLAUDE.work.md"
ln -sf "$DOTFILES_DIR/.claude/CLAUDE.personal.md" "$HOME/.claude/CLAUDE.personal.md"
ln -sf "$DOTFILES_DIR/.claude/statusline-command.sh" "$HOME/.claude/statusline-command.sh"

if [ ! -f "$HOME/.claude/CLAUDE.md" ]; then
    echo "Creating default ~/.claude/CLAUDE.md..."
    printf '@CLAUDE.shared.md\n@CLAUDE.work.md\n' > "$HOME/.claude/CLAUDE.md"
fi

# 4. Build and install colorsync
echo "Building colorsync..."
mkdir -p "$HOME/.local/bin"
cd "$DOTFILES_DIR/tools/colorsync"
go build -o "$HOME/.local/bin/colorsync" .
echo "  Installed colorsync to ~/.local/bin/colorsync"

# 6. Per-machine env
if [ ! -f "$HOME/.env/env.custom.sh" ]; then
    echo ""
    echo "NOTE: Create ~/.env/env.custom.sh with your machine-specific config."
    echo "Example:"
    echo "  export LC_ALL=en_GB.UTF-8"
    echo "  export SOME_API_KEY=your-key-here"
    echo "  alias myproject=\"WORKSPACE_SESSION_NAME=myproject;WORKSPACE_PATH=~/code/myproject;twosplit\""
fi

echo ""
echo "Done! Restart your shell or run: source ~/.zshrc"
