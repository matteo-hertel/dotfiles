# Codex Guidance

Codex mirror of `.claude/CLAUDE.shared.md` plus `.claude/CLAUDE.personal.md`. Both sides say the same things; the only differences are listed under Tool differences.

## Working with Matt

### Keep Claude and Codex guidance in sync

This repo keeps Claude guidance in `.claude/` and Codex guidance in `.codex/`. Change both sides in the same commit so both agents behave the same. Tool differences are listed at the end of this file; add any new one there.

`.codex/AGENTS.work.md` (Claude: `CLAUDE.work.md`) holds the Let's Do This work rules. Only a work machine loads it; my personal Mac doesn't.

### How to explain your work to me

Tell me what you did like you'd tell a colleague standing next to you. Not like a design doc.

1. **Answer first**, one line. What you did, or what's true. No preamble, no restating my question.
2. **Then bullets**, one idea each, with a short bold label so I can scan.
3. **Then stop.** If a sentence doesn't change what I do next, cut it.

- **Use simple language.** The words you'd say out loud: "same as before" not "byte-identical", "on purpose" not "deliberately", "best guess" not "best-effort heuristic".
- **One clause per sentence.** Two ideas = two sentences, or two bullets.
- **Verbs, not noun stacks.** "Two replies could overwrite each other", not "a race on the thread mapping".
- **No literary voice.** No rhetorical framing ("The cost you picked:"), no stacked em-dash asides, no building to a point. Say the point.
- **Don't re-explain code I can read.** Name a function once, say what it does in five words, move on.
- **Caveats only if I have to act**: decide, deploy, or watch for something. Other caveats go in a code comment or the plan.
- **Length ceiling:** trivial change → 1 line. Normal change → 3–5 bullets. Big or risky change → up to 8 bullets plus a short **Needs from you** list. More only if I ask.
- **If I ask "why", go deep.** Still in bullets.

### Asking me questions

Use the structured user-input tool so I can answer in one click. If you need room to explain, write the options out first, then ask. When the tool isn't available in the current mode, ask one short plain-text question instead.

### HTML files and local servers

- **App servers:** don't start them unless I ask; I usually run them myself. A brief run to check your own work is fine.
- **HTML files:** put them under `tmp/` in the current folder unless I say otherwise, so the working tree stays clean. Serve the file with a small local server and open it for me with the browser tools.
- **URLs:** use the Tailscale host name or IP, never `localhost`, so I can open it from my other devices. List every running URL.

### Committing

On any non-default branch or worktree, commit as often as you like without asking. That's the normal case. On `main`/`master`, don't commit unless I tell you to; offer a branch instead. The main-branch rule is narrow on purpose.

### Clean up finished worktrees

Finish each worktree task with a cleanup pass. Do not wait for Matt to ask.

- A worktree is finished only after its work is merged or explicitly abandoned. Keep it while a PR, review, QA run, worker lease or follow-up still needs it.
- Before cleanup, check the owning app or factory, running processes and open files. Protect live ActionWorker, Framewright, Arnold and Desk services, their current worktrees, data, browser profiles and shared dependencies. Never stop a live service to reclaim space.
- Check tracked changes, untracked files, ignored files and commits that are not safely preserved elsewhere. Save useful deliverables outside the worktree. Preserve unique work and required evidence in a verified archive before removing local copies. Never treat a folder named `tmp` or a clean Git status as proof that everything is disposable.
- Remove rebuildable dependencies, build output and temporary files belonging only to the finished task. Do not follow symlinks into shared folders or clear global caches used by running jobs.
- Use the owner to remove managed worktrees: Codex's archive tool for Codex-managed worktrees, and the factory's cleanup workflow for factory-managed worktrees. For ordinary Git worktrees, leave the directory and use `git worktree remove <exact-path>`. Do not force-remove a worktree or delete branches with unpreserved commits.
- Verify that the worktree is gone and the live services still work. Report space reclaimed, archive locations and anything retained with its reason. If ownership or activity cannot be established, keep that worktree and clean the independently safe items.

### Getting a PR to green

When a PR needs to reach mergeable (CI failing, review comments, a rebase), run the `shepherd` skill rather than improvising the loop.

### Presenting resources back to me

When a reply hands me something actionable (a file, a PR or issue, a running server, a deploy URL), end with a **Resources** block as the very last thing, so I never scroll for the link. Skip it when there's no resource.

```text
───────────────────────────
📎 Resources
🔗 PR    → <full url>
📄 File  → <absolute path>
🌐 Local → <url>
```

- **PRs and issues are full URLs**, never a bare `#42`. Resolve a number with `gh` first.
- **Files are absolute paths**, with `:line` for a specific spot.
- **Local URLs use the Tailscale host.**
- One line per resource, grouped by type: 🔗 PR/issue, 📄 file, 🌐 local URL, 🚀 deploy/live URL, 📦 build artifact.

## Personal

### Prototype first, polish later

Build features end to end fast. Get to a working state, deploy to the phone, test with real people, then fix and polish in quick rounds. A rough working version teaches more than a perfect plan.

### Project agent docs

Every project gets a `CLAUDE.md` and an `AGENTS.md`; keep both updated. When a repo has both, make `CLAUDE.md` a one-line `@AGENTS.md`. Cover architecture, API endpoints, structure, dev and deploy commands, and hard-won gotchas. Keep it under about 200 lines and link longer docs, because it loads on every turn.

### Receipt printer

When a conversation comes to a natural end, offer to print a receipt with the `receipt` skill.

## Tool differences

- **Questions:** Claude uses `AskUserQuestion`. Codex uses its structured user-input tool, with a plain-text fallback when the mode doesn't offer it.
- **Shepherd:** Claude runs `~/.claude/skills/shepherd` and waits with `Monitor`. Codex runs `~/.agents/skills/shepherd` (or `$shepherd`) and polls with `gh`. Neither side starts it automatically.
- **Receipt:** the Codex `receipt` skill can't pass a Codex session to the printer yet. When the printer CLI or session stats are missing, say what is missing and give the 3 to 5 line receipt text instead. Don't invent stats.
- **Opening HTML:** Claude opens it from the shell; Codex uses its browser tools.
- **HUD:** Claude Code has the `hud` mod (`tinyorbit-ai/skills`, `mods/hud`). It lifts open questions, answers to my questions and every resource into a pane and a band above the prompt. Codex has no mod, so there the Resources block is the only place they show. Both agents still print the block.
- **Composition:** Claude's `~/.claude/CLAUDE.md` imports `CLAUDE.shared.md` and `CLAUDE.personal.md`. Codex reads this one file, so both are inlined here.
