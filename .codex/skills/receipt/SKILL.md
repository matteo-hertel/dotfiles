---
name: receipt
description: Print a thermal-printer receipt for the current Codex session. Use when asked to "print a receipt", "receipt this session", or "/receipt" - summarizes what was accomplished and uses the local thermal-print CLI when it supports Codex session data.
---

# Receipt

## Overview

Print a short physical receipt for this Codex session when the local
thermal-printer tooling supports it. If deterministic Codex session stats
are not available, provide the receipt text and clearly explain what is
missing instead of inventing stats.

This mirrors `.claude/skills/receipt/SKILL.md`. When changing this skill,
update the Claude receipt skill at the same time unless the difference is
tool-specific and documented in both files.

## Steps

1. Write a 3 to 5 line narrative summary of what was accomplished in
   this session. Each line must be at most 32 characters. Tone: warm,
   observational, concrete. Avoid buzzwords, boilerplate, leading
   dashes, and numbering. Use plain text, one line per line.

2. Prefer a Codex-aware `thermal-print` command if it exists locally.
   Check help before printing:

   ```bash
   thermal-print print receipt --help
   ```

3. If the CLI exposes Codex session flags, pass the current session id,
   working directory, and summary via argv. Never interpolate an
   untrusted cwd into a shell string.

   ```bash
   thermal-print print receipt \
     --session-id "$CODEX_SESSION_ID" \
     --cwd "$PWD" \
     --summary "<your 3-5 line summary here>"
   ```

4. If `CODEX_SESSION_ID` is unset, `thermal-print` is not on `PATH`, or
   the installed CLI only supports Claude session JSONL files, do not
   print. Tell Matt the exact missing capability and provide the 3 to 5
   line receipt text in the response.

5. If the CLI exits non-zero, surface stderr verbatim and stop. Do not
   retry or work around the error.
