#!/bin/bash
# Installs the third-party agent skills into ~/.agents/skills for Claude Code and Codex.
set -euo pipefail

add() {
    local source="$1"; shift
    local skills=()
    for skill in "$@"; do skills+=(--skill "$skill"); done
    npx -y skills add "$source" "${skills[@]}" -g -a claude-code -a codex -y
}

add vercel-labs/skills find-skills
add vercel-labs/agent-skills web-design-guidelines vercel-react-best-practices
add pbakaus/impeccable impeccable
add matteo-hertel/loud-authority-brand loud-authority-brand

add shhac/agent-cloudflare agent-cloudflare
add shhac/agent-dd agent-dd
add shhac/agent-fathom agent-fathom
add shhac/agent-mongo agent-mongo
add shhac/agent-notion agent-notion
add shhac/agent-slack agent-slack
add shhac/agent-statsig agent-statsig
add shhac/agent-vercel agent-vercel
add shhac/lin lin

add tinyorbit-ai/skills forge forge-build forge-debug forge-design-explore forge-design-system \
    forge-discovery forge-docs forge-dx forge-harden forge-init forge-plan forge-polish \
    forge-principles forge-retro forge-review forge-ship forge-wiki lizard

if [ "${DOTFILES_WORK:-0}" = 1 ]; then
    add stampedeapp/ai-tools slide-builder ux-edit-save ux-modals ux-settings-pages ux-tables-bulk-actions ux-writing
    add git@github.com:stampedeapp/agent-nova.git nova
fi
