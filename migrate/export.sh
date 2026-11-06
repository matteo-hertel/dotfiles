#!/bin/bash
# Packs what you pick from this Mac into one encrypted disk image, ready for import.sh on the new one.
#   migrate/export.sh [--all] [--to <tailscale device>] [-o <image path>]
# What it can pack is listed in migrate/manifest.local.sh, or migrate/manifest.example.sh if that is missing.
set -euo pipefail

DOTFILES_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ALL=0
TO=""
OUT="$HOME/Desktop/migrate-$(date +%Y%m%d-%H%M).sparseimage"
while [ $# -gt 0 ]; do
    case "$1" in
        --all) ALL=1 ;;
        --to) TO="$2"; shift ;;
        -o) OUT="$2"; shift ;;
        *) echo "usage: migrate/export.sh [--all] [--to <tailscale device>] [-o <image path>]" >&2; exit 2 ;;
    esac
    shift
done
[ -e "$OUT" ] && { echo "export: $OUT already exists" >&2; exit 1; }

SEP=$'\x1f'
NAMES=() KINDS=() ARGS=()
add() { NAMES+=("$1"); KINDS+=("$2"); shift 2; local IFS="$SEP"; ARGS+=("$*"); }
group() { local name="$1"; shift; add "$name" files "$@"; }
keychain() { add "$1" keychain "$2"; }
repos() { local name="$1"; shift; add "$name" repos "$@"; }

MANIFEST="$DOTFILES_DIR/migrate/manifest.local.sh"
[ -f "$MANIFEST" ] || MANIFEST="$DOTFILES_DIR/migrate/manifest.example.sh"
cd "$HOME"
shopt -s nullglob
# shellcheck source=/dev/null
source "$MANIFEST"
shopt -u nullglob
echo "Using $MANIFEST"

split() { local IFS="$SEP"; read -ra PARTS <<<"$1"; }

paths_of() {
    split "$1"
    local skip=0 part
    PATHS=() EXCLUDES=(--exclude node_modules --exclude .DS_Store)
    for part in "${PARTS[@]}"; do
        if [ "$skip" = 1 ]; then EXCLUDES+=(--exclude "$part"); skip=0
        elif [ "$part" = "-x" ]; then skip=1
        elif [ -e "$HOME/$part" ] || [ -L "$HOME/$part" ]; then PATHS+=("$part")
        fi
    done
}

size_of() {
    local i=$1
    case "${KINDS[$i]}" in
        files)
            paths_of "${ARGS[$i]}"
            [ ${#PATHS[@]} -gt 0 ] || { echo "nothing here"; return; }
            du -skc "${PATHS[@]}" 2>/dev/null | tail -1 | awk '{ s=$1/1024; if (s >= 1024) printf "%.1f GB", s/1024; else printf "%.0f MB", s }' ;;
        repos) split "${ARGS[$i]}"; echo "$(for d in "${PARTS[@]}"; do [ -d "$HOME/$d/.git" ] && echo; done | wc -l | tr -d ' ') repos" ;;
        keychain) echo "Keychain prompts at the end" ;;
    esac
}

PICKED=()
echo
echo "Pick what to move (Enter = yes):"
for i in "${!NAMES[@]}"; do
    if [ "$ALL" = 1 ]; then PICKED+=("$i"); continue; fi
    read -rp "  ${NAMES[$i]} ($(size_of "$i"))? [Y/n] " answer </dev/tty
    case "$answer" in [nN]*) ;; *) PICKED+=("$i") ;; esac
done

echo
read -rsp "Passphrase for the image: " PASS </dev/tty; echo
read -rsp "Again: " PASS2 </dev/tty; echo
[ "$PASS" = "$PASS2" ] && [ -n "$PASS" ] || { echo "export: passphrases differ or are empty" >&2; exit 1; }
unset PASS2

printf '%s' "$PASS" | hdiutil create -size 200g -type SPARSE -fs APFS -encryption AES-256 -stdinpass -volname migrate "$OUT" >/dev/null 2>&1
MNT=$(printf '%s' "$PASS" | hdiutil attach -stdinpass -nobrowse -owners on "$OUT" 2>/dev/null | awk -F'\t' '/\/Volumes\//{print $NF}')
unset PASS
trap 'hdiutil detach "$MNT" >/dev/null 2>&1 || true' EXIT
mkdir -p "$MNT/home"
echo "$HOME" > "$MNT/old-home"
: > "$MNT/repos.tsv"
: > "$MNT/keychain.jsonl"
chmod 600 "$MNT/keychain.jsonl"

copy() {
    local path
    for path in "$@"; do
        echo "  $path"
        rsync -aR "${EXCLUDES[@]}" "./$path" "$MNT/home/" \
            || echo "  warning: some files in $path changed or vanished while copying"
    done
}

copy_repos() {
    split "$1"
    local repo remote f
    for repo in "${PARTS[@]}"; do
        [ -d "$HOME/$repo/.git" ] || continue
        remote=$(git -C "$HOME/$repo" remote get-url origin 2>/dev/null) || continue
        printf '%s\t%s\n' "$repo" "$remote" >> "$MNT/repos.tsv"
        echo "  $repo"
        git -C "$HOME/$repo" ls-files -o -i --exclude-standard -z 2>/dev/null \
            | tr '\0' '\n' \
            | grep -E '(^|/)(\.env[^/]*|[^/]*\.env|\.envrc|\.dev\.vars|\.claude\.env)$' \
            | grep -vE '(^|/)(node_modules|dbt_packages|\.ai-cache)/|\.(example|sample|template)$' \
            | while read -r f; do
                mkdir -p "$MNT/home/$repo/$(dirname "$f")"
                cp -p "$HOME/$repo/$f" "$MNT/home/$repo/$f"
                echo "    $f"
            done || true
    done
}

copy_keychain() {
    local svc acct secret
    security dump-keychain 2>/dev/null \
        | awk -F'<blob>=' '/"svce"/{s=$2} /"acct"/{a=$2} /^keychain:/{ if (s != "") print s "\t" a; s=""; a="" } END { if (s != "") print s "\t" a }' \
        | tr -d '"' | grep -E "$1" | sort -u > "$MNT/keychain-items.tsv" || true
    while IFS=$'\t' read -r svc acct; do
        [ "$acct" = "<NULL>" ] && continue
        if secret=$(security find-generic-password -s "$svc" -a "$acct" -w 2>/dev/null); then
            jq -cn --arg s "$svc" --arg a "$acct" --arg w "$secret" '{s:$s, a:$a, w:$w}' >> "$MNT/keychain.jsonl"
            echo "  ok      $svc / $acct"
        else
            echo "  skipped $svc / $acct"
        fi
    done < "$MNT/keychain-items.tsv"
    rm "$MNT/keychain-items.tsv"
}

echo "Dotfiles (always packed, so the new Mac can bootstrap)..."
EXCLUDES=(--exclude tmp --exclude .DS_Store)
copy "${DOTFILES_DIR#"$HOME"/}"
cp "$DOTFILES_DIR/migrate/import.sh" "$MNT/import.sh"
echo "${DOTFILES_DIR#"$HOME"/}" > "$MNT/dotfiles-path"

KEYCHAIN_LAST=()
for i in "${PICKED[@]+"${PICKED[@]}"}"; do
    case "${KINDS[$i]}" in
        files) echo "${NAMES[$i]}..."; paths_of "${ARGS[$i]}"; [ ${#PATHS[@]} -gt 0 ] && copy "${PATHS[@]}" ;;
        repos) echo "${NAMES[$i]}..."; copy_repos "${ARGS[$i]}" ;;
        keychain) KEYCHAIN_LAST+=("$i") ;;
    esac
done
for i in "${KEYCHAIN_LAST[@]+"${KEYCHAIN_LAST[@]}"}"; do
    echo "${NAMES[$i]} (approve each Keychain prompt with your password)..."
    copy_keychain "${ARGS[$i]}"
done

du -sh "$MNT" | awk '{print "Packed " $1}'
hdiutil detach "$MNT" >/dev/null 2>&1
trap - EXIT
echo "Done: $OUT"

if [ -z "$TO" ] && [ "$ALL" = 0 ]; then
    read -rp "Send it to which Tailscale device? (Enter to skip) " TO </dev/tty
fi
if [ -n "$TO" ]; then
    tailscale file cp "$OUT" "$TO:" && echo "Sent to $TO. Accept it there if Tailscale asks."
fi

echo
echo "On the new Mac:"
echo "  hdiutil attach ~/Downloads/$(basename "$OUT") && /Volumes/migrate/import.sh"
