---
name: receipt
description: Print a thermal-printer receipt for the current Claude Code session. Use when asked to "print a receipt", "receipt this session", or "/receipt" — summarizes tokens used and what was accomplished as a physical receipt on the Star TSP 100III.
allowed-tools: [Bash]
---

Print a physical thermal-printer receipt for **this** Claude Code session.

This mirrors the Codex receipt skill at `.codex/skills/receipt/SKILL.md`.
When changing one, update the other unless the difference is
tool-specific and documented in both files.

You have the full transcript of this session in your context already — use
it to write the narrative summary directly, then pass it to the CLI via
`--summary`. **Do not** read the session JSONL yourself; the CLI does that
for the deterministic stats.

## Steps

1. Write a **3 to 5 line** narrative summary of what was accomplished in
   this session. Each line at most 32 characters. Tone: warm,
   observational, concrete — like a journal note to your future self.
   No buzzwords. No "session completed successfully" boilerplate. No
   leading dashes or numbering. Plain text, one line per line.

2. Pass the summary to the CLI via `--summary`, plus the session id
   (`$CLAUDE_CODE_SESSION_ID`, verified live on 2026-05-27) and the
   working directory (`$PWD`). Both via argv, never via shell
   interpolation in a string — a malicious cwd cannot inject shell:

   ```bash
   thermal-print print receipt \
     --session-id "$CLAUDE_CODE_SESSION_ID" \
     --cwd "$PWD" \
     --summary "<your 3-5 line summary here>"
   ```

   This skill is installed globally and assumes `thermal-print` is on
   PATH (run `uv tool install .` once from the thermal-printer repo at
   ~/code/thermal-printer). From a checkout without that, substitute
   `uv run thermal-print …`.

3. If the CLI exits non-zero, surface the stderr verbatim and stop — do
   not try to print again or work around the error.

The deterministic stats portion (tokens, time, files, tools) is read by
the CLI from `~/.claude/projects/<encoded-cwd>/<session-id>.jsonl`. The
narrative is the *only* thing you write.
