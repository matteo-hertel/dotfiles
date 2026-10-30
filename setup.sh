#!/bin/bash
set -e

DOTFILES_DIR="$(cd "$(dirname "$0")" && pwd)"

link_agent_path() {
    local src="$1"
    local dst="$2"
    local backup

    if [ -L "$dst" ] && [ "$(readlink "$dst")" = "$src" ]; then
        return
    fi

    if [ -e "$dst" ] || [ -L "$dst" ]; then
        backup="${dst}.backup.$(date +%Y%m%d%H%M%S)"
        mv "$dst" "$backup"
        echo "  Backed up existing $dst to $backup"
    fi

    ln -s "$src" "$dst"
}

echo "Setting up dotfiles from $DOTFILES_DIR"

# 1. Packages: Homebrew on macOS; on Linux they come from apt and ~/.local
if [ "$(uname)" = Linux ]; then
    export PATH="$HOME/.local/bin:$HOME/.local/go/bin:$PATH"
    for tool in stow go zsh tmux git; do
        command -v "$tool" &>/dev/null || { echo "Missing $tool. Install it first (sudo apt-get install -y stow zsh tmux git; Go into ~/.local/go)."; exit 1; }
    done
    [ -d "$HOME/.oh-my-zsh" ] || git clone --depth 1 https://github.com/ohmyzsh/ohmyzsh.git "$HOME/.oh-my-zsh"
    p10k="$HOME/.oh-my-zsh/custom/themes/powerlevel10k"
    [ -d "$p10k" ] || git clone --depth 1 https://github.com/romkatv/powerlevel10k.git "$p10k"
else
    if ! command -v brew &>/dev/null; then
        echo "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        eval "$(/opt/homebrew/bin/brew shellenv)"
    fi

    echo "Installing packages from Brewfile..."
    brew bundle --file="$DOTFILES_DIR/Brewfile" || echo "WARNING: some Brewfile items failed to install, see above. Carrying on."
fi

# 1.5 Shell, runtimes and global CLIs
if [ ! -d "$HOME/.oh-my-zsh" ]; then
    echo "Installing oh-my-zsh..."
    RUNZSH=no KEEP_ZSHRC=yes sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
fi

if [ ! -s "$HOME/.nvm/nvm.sh" ]; then
    echo "Installing nvm..."
    PROFILE=/dev/null bash -c "$(curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh)"
fi
export NVM_DIR="$HOME/.nvm"
. "$NVM_DIR/nvm.sh"
nvm install 24 >/dev/null
nvm install 22 >/dev/null
nvm alias default 22 >/dev/null
nvm use default >/dev/null
npm install -g @openai/codex @google/gemini-cli

command -v claude &>/dev/null || curl -fsSL https://claude.ai/install.sh | bash
command -v bun &>/dev/null || curl -fsSL https://bun.sh/install | bash

go install github.com/kevindutra/crit/cmd/crit@v0.2.1
go install github.com/matteo-hertel/tmux-super-powers/cmd/tsp@latest

# 2. Stow dotfiles
echo "Linking dotfiles with stow..."
cd "$DOTFILES_DIR"
for f in .zshrc .gitconfig .gitignore_global .tmux.conf .zshtheme .tmux-super-powers.yaml; do
    if [ -f "$HOME/$f" ] && [ ! -L "$HOME/$f" ]; then
        mv "$HOME/$f" "$HOME/$f.backup.$(date +%Y%m%d%H%M%S)"
        echo "  Backed up existing ~/$f"
    fi
done
stow --target="$HOME" .

# 2.5 TPM (Tmux Plugin Manager)
if [ ! -f "$HOME/.tmux/plugins/tpm/tpm" ]; then
    echo "Installing TPM..."
    git clone https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"
fi
echo "Installing tmux plugins..."
"$HOME/.tmux/plugins/tpm/bin/install_plugins"

# 3. Claude config (separate from stow to avoid folding ~/.claude)
# CLAUDE.md is per-machine — create a default one if missing.
# Work rules (CLAUDE.work.md / .codex/AGENTS.work.md) are opt-in per machine:
# add @CLAUDE.work.md to ~/.claude/CLAUDE.md and see .codex/AGENTS.work.md for Codex.
echo "Linking Claude config..."
mkdir -p "$HOME/.claude/skills"
ln -sf "$DOTFILES_DIR/.claude/CLAUDE.shared.md" "$HOME/.claude/CLAUDE.shared.md"
ln -sf "$DOTFILES_DIR/.claude/CLAUDE.work.md" "$HOME/.claude/CLAUDE.work.md"
ln -sf "$DOTFILES_DIR/.claude/CLAUDE.personal.md" "$HOME/.claude/CLAUDE.personal.md"
ln -sf "$DOTFILES_DIR/.claude/statusline-command.sh" "$HOME/.claude/statusline-command.sh"
link_agent_path "$DOTFILES_DIR/.claude/skills/shepherd" "$HOME/.claude/skills/shepherd"
link_agent_path "$DOTFILES_DIR/.claude/skills/receipt" "$HOME/.claude/skills/receipt"
[ -e "$DOTFILES_DIR/.claude/skills/becky" ] && link_agent_path "$DOTFILES_DIR/.claude/skills/becky" "$HOME/.claude/skills/becky"

if [ ! -f "$HOME/.claude/CLAUDE.md" ]; then
    echo "Creating default ~/.claude/CLAUDE.md..."
    printf '@CLAUDE.shared.md\n@CLAUDE.personal.md\n' > "$HOME/.claude/CLAUDE.md"
fi

# 4. Codex config (separate from stow to avoid folding ~/.codex)
echo "Linking Codex config..."
mkdir -p "$HOME/.codex/skills" "$HOME/.agents/skills"
link_agent_path "$DOTFILES_DIR/.codex/AGENTS.md" "$HOME/.codex/AGENTS.md"
link_agent_path "$DOTFILES_DIR/.codex/skills/receipt" "$HOME/.codex/skills/receipt"
link_agent_path "$DOTFILES_DIR/.codex/skills/shepherd" "$HOME/.agents/skills/shepherd"
[ -e "$DOTFILES_DIR/.codex/skills/becky" ] && link_agent_path "$DOTFILES_DIR/.codex/skills/becky" "$HOME/.agents/skills/becky"

# 5. Build and install colorsync
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
