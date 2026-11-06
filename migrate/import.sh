#!/bin/bash
# Unpacks an image made by export.sh onto this Mac: dotfiles, setup.sh, then everything you packed.
#   hdiutil attach ~/Downloads/migrate-*.sparseimage && /Volumes/migrate/import.sh
#   migrate/import.sh <image.sparseimage>
set -euo pipefail

if [ $# -gt 0 ]; then
    read -rsp "Passphrase for the image: " PASS; echo
    MNT=$(printf '%s' "$PASS" | hdiutil attach -stdinpass -nobrowse -owners on "$1" 2>/dev/null | awk -F'\t' '/\/Volumes\//{print $NF}')
    unset PASS
    trap 'hdiutil detach "$MNT" >/dev/null 2>&1 || true' EXIT
else
    MNT="$(cd "$(dirname "$0")" && pwd)"
fi
[ -f "$MNT/old-home" ] || { echo "import: $MNT is not an image made by export.sh" >&2; exit 1; }

OLD_HOME=$(cat "$MNT/old-home")
if [ "$OLD_HOME" != "$HOME" ]; then
    echo "import: the old home was $OLD_HOME and this one is $HOME." >&2
    echo "Claude's per-project memory and many configs store absolute paths. Use the same username, or fix paths after." >&2
    read -rp "Carry on anyway? [y/N] " yes
    [ "$yes" = y ] || exit 1
fi

DOTFILES_REL=$(cat "$MNT/dotfiles-path")
if [ ! -d "$HOME/$DOTFILES_REL" ]; then
    echo "Copying dotfiles to ~/$DOTFILES_REL..."
    mkdir -p "$(dirname "$HOME/$DOTFILES_REL")"
    rsync -a "$MNT/home/$DOTFILES_REL/" "$HOME/$DOTFILES_REL/"
fi
read -rp "Run setup.sh now (Homebrew, apps, shell, links)? [Y/n] " answer
case "$answer" in [nN]*) ;; *) "$HOME/$DOTFILES_REL/setup.sh" ;; esac

echo "Copying files into $HOME (anything replaced is kept as *.pre-migrate)..."
rsync -a --backup --suffix=.pre-migrate --exclude "/$DOTFILES_REL" "$MNT/home/" "$HOME/"

if [ "$OLD_HOME" != "$HOME" ]; then
    old_key=$(printf '%s' "$OLD_HOME" | tr '/.' '--')
    new_key=$(printf '%s' "$HOME" | tr '/.' '--')
    for dir in "$HOME"/.claude/projects/"$old_key"*; do
        [ -d "$dir" ] && mv "$dir" "${dir/$old_key/$new_key}"
    done
fi

if [ -d "$HOME/.ssh" ]; then
    chmod 700 "$HOME/.ssh"
    chmod 600 "$HOME/.ssh/id_ed25519" "$HOME/.ssh/config" 2>/dev/null || true
fi
for dir in "$HOME"/.config/agent-* "$HOME/.config/app.paulie.agent-slack" "$HOME/.config/lin" "$HOME/.aws" "$HOME/.gnupg" "$HOME/.snowflake"; do
    [ -e "$dir" ] && chmod -R go-rwx "$dir"
done

pub="$HOME/.ssh/id_ed25519.pub"
if [ -f "$pub" ] && ! grep -qF "$(cut -d' ' -f2 "$pub")" "$HOME/.ssh/authorized_keys" 2>/dev/null; then
    cat "$pub" >> "$HOME/.ssh/authorized_keys"
    chmod 600 "$HOME/.ssh/authorized_keys"
    echo "Added your key to ~/.ssh/authorized_keys"
fi

if [ -s "$MNT/keychain.jsonl" ]; then
    echo "Writing CLI logins to the Keychain..."
    while IFS= read -r line; do
        svc=$(jq -r .s <<<"$line")
        acct=$(jq -r .a <<<"$line")
        [ "$svc" = "Claude Code-credentials" ] && acct="$USER"
        security add-generic-password -U -s "$svc" -a "$acct" -w "$(jq -r .w <<<"$line")" \
            && echo "  ok      $svc / $acct" || echo "  FAILED  $svc / $acct"
    done < "$MNT/keychain.jsonl"
fi

if [ -s "$MNT/repos.tsv" ]; then
    echo "Cloning repos that are missing..."
    while IFS=$'\t' read -r rel remote; do
        [ -d "$HOME/$rel/.git" ] && continue
        tmp="$HOME/$rel.migrate-env"
        [ -d "$HOME/$rel" ] && mv "$HOME/$rel" "$tmp"
        mkdir -p "$(dirname "$HOME/$rel")"
        if git clone --quiet "$remote" "$HOME/$rel"; then
            echo "  cloned  $rel"
        else
            echo "  FAILED  $rel ($remote)"
        fi
        [ -d "$tmp" ] && rsync -a "$tmp/" "$HOME/$rel/" && rm -rf "$tmp"
    done < "$MNT/repos.tsv"
fi

echo
echo "Login check:"
check() {
    command -v "$1" >/dev/null 2>&1 || return 0
    if "$@" >/dev/null 2>&1; then echo "  ok      $*"; else echo "  SIGN IN $*"; fi
}
check gh auth status
check claude auth status
check codex login status
check lin auth status
check agent-slack auth test
check agent-notion auth status
check agent-statsig project test
check agent-fathom auth check

if compgen -G "$MNT/home/Library/LaunchAgents/*.plist" >/dev/null; then
    echo
    echo "Launch agents were copied but not started. Start them with:"
    for plist in "$MNT"/home/Library/LaunchAgents/*.plist; do
        echo "  launchctl bootstrap gui/$(id -u) $HOME/Library/LaunchAgents/$(basename "$plist")"
    done
fi
echo
echo "Done. Open a new terminal, then delete the image once everything works. It holds your logins."
