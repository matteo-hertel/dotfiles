# shellcheck shell=bash
# What export.sh can pack. Copy to manifest.local.sh (gitignored) and make it yours.
# Each line is one yes/no question. Paths are relative to $HOME and globs work.
#   group "<name>" [-x <exclude pattern>]... <path>...
#   repos "<name>" <repo path>...       clone list plus each repo's gitignored env files
#   keychain "<name>" '<regex>'          matched against "<service><TAB><account>"

group "Claude home and memories" -x shell-snapshots -x session-env -x debug -x paste-cache -x telemetry \
    .claude .claude.json
group "Codex config and memories (no chat history)" \
    .codex/auth.json .codex/config.toml .codex/AGENTS.md .codex/rules .codex/skills .codex/memories .codex/memories_1.sqlite*
group "Agent skills" .agents
group "SSH key" .ssh/id_ed25519 .ssh/id_ed25519.pub .ssh/config .ssh/known_hosts
group "CLI config" .config/gh .aws .npmrc .gnupg -x 'S.*'
group "Shell history" .zsh_history
repos "Repos and their env files" code/* work/code/*
keychain "CLI logins in the Keychain" '^(gh:github\.com|Claude Code-credentials)	'
