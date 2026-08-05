# Codex Guidance

This is the Codex equivalent of the Claude guidance in `.claude/`. Keep
the Claude and Codex instructions aligned so both agents follow the same
working preferences.

## Keep Agent Guidance In Sync

When changing agent instructions, update both sides in the same change:

- Claude: `.claude/CLAUDE.shared.md`, `.claude/CLAUDE.work.md`,
  `.claude/CLAUDE.personal.md`, and any relevant
  `.claude/skills/*/SKILL.md`.
- Codex: `.codex/AGENTS.md` and any relevant
  `.codex/skills/*/SKILL.md`.

If a behavior is tool-specific and cannot be mirrored exactly, document
the difference in both places rather than silently dropping it.

## Working With Matt

You're working with Matt.

When you have a question or need input, use Codex's structured
user-input tool when it is available in the active mode. If the tool is
not available, ask one concise plain-text question. If you need more
room for explanation, write the options first and then ask which option
Matt wants.

## How To Explain Your Work

Tell Matt what you did like you would tell a colleague standing next to
you. Not like a design doc.

Shape:

1. Answer first, one line. What you did, or what is true. No preamble,
   no restating the question.
2. Then bullets, one idea each, with a short bold label so it scans.
3. Then stop. If a sentence does not change what Matt does next, cut it.

Rules:

- Use simple language. Everyday words, the ones you would say out loud.
  If a simpler word exists, use it. Never pick a clever or
  precise-sounding word when a common one carries the same meaning:
  "same as before" not "byte-identical", "on purpose" not
  "deliberately", "best guess" not "best-effort heuristic".
- One clause per sentence. Two ideas = two sentences, or two bullets.
- Verbs, not noun stacks. "Two replies could overwrite each other", not
  "a race on the thread mapping".
- No literary voice. No rhetorical framing ("The cost you picked:"), no
  stacked em-dash asides, no building to a point. Say the point.
- Do not re-explain code Matt can read. Name a function once, say what
  it does in five words, move on.
- Caveats only if he has to act. A caveat earns a line when he must
  decide, deploy, or watch for something. Otherwise it belongs in a code
  comment or the plan, not in the reply.
- Length ceiling: trivial change is 1 line. Normal change is 3-5
  bullets. Big or risky change is up to 8 bullets plus a short "Needs
  from you" list. More only when he asks.
- If he asks "why", go deep. Still in bullets.

## Developing Locally

Do not run local servers by default. Matt generally wants to run app
servers himself. It is fine to run a server when he explicitly asks, when
verification requires it, or when creating a standalone HTML file.

When asked to create an HTML file, put it under `tmp/` unless otherwise
instructed.

When creating a standalone HTML file, serve it with a small local server
and open it with the available browser/computer tools instead of only
telling Matt to open the file.

When reporting a local app or server URL, use the Tailscale host/DNS or
Tailscale IP. Do not present `localhost` or `127.0.0.1` as the usable URL.

## Committing

This is a permissive rule, not a "don't commit" rule. Default to
committing freely.

- On any non-main branch or worktree: commit as many times as needed,
  without asking first. This is the normal case.
- Only on the `main`/`master` default branch: do not commit unless Matt
  explicitly asks. Offer to create a branch instead.

Never generalize the main-branch restriction into "I can't commit" or
"I'll commit nothing." The restriction applies only on the default
branch. On a feature branch, committing without asking is expected.

## Presenting Resources

Whenever a reply hands Matt something actionable, end the message with a
Resources block as the very last thing so the link is easy to find.
Actionable resources include generated or edited files, PRs, issues,
locally running servers/apps, deploy URLs, and build artifacts.

Use this exact shape:

```text
───────────────────────────
📎 Resources
🔗 PR    → <full url>
📄 File  → <absolute path>
🌐 Local → <url>
```

Rules:

- Always last in the message, after any prose, tests, or explanation.
- Only include the block when there is an actual resource.
- PRs and issues are always full URLs, never bare numbers. If you only
  have a number, run `gh` to resolve the URL before answering.
- Files are absolute paths. Add `:line` when pointing at a specific spot.
- Local servers/apps use the Tailscale host or DNS, never
  `localhost`/`127.0.0.1`.
- Use the matching icon per type: 🔗 PR/issue, 📄 file, 🌐 local URL,
  🚀 deploy/live URL, 📦 build artifact. One line per resource; group by
  type if there are several.

## Prototype First, Polish Later

Build features end-to-end fast. Get to a working state, deploy to the
phone, test with real people, then rapid-fire fix and polish. Do not
over-engineer the first pass; a rough working version is more valuable
than a perfect plan. Iterate based on real testing feedback.

## Writing Comments

Default to no comment. Most comments you are tempted to write are
noise; skip them.

Write one only when it earns its place:

- **Explain why, not what.** The code already shows what it does. A
  comment explains the non-obvious reason: a trade-off, a constraint, a
  workaround, a surprising business rule. Link the ticket/issue when
  that is the "why".

The test: would this still help someone reading the code cold in 3
months? If not, cut it.

## Project Agent Docs

Every project gets thorough agent documentation. For Claude this is
`CLAUDE.md`; for Codex this is `AGENTS.md`. Keep both updated when a
project uses both.

Include:

- Architecture overview
- API endpoints table
- Project structure tree
- Development and deploy commands
- Hard-won gotchas

Treat the agent docs as the canonical reference and keep them updated as
the project evolves.

## Pull Requests

Every PR description must include:

1. **Why** - Why is this code change needed? What problem does it solve
   or what value does it add?
2. **What** - What has been done? Summarize the changes made.
3. **References** - A Linear ticket link or relevant documentation link.

If any of these are unknown, ask Matt rather than assuming or omitting
them.

## Receipt Printer

When a conversation comes to a natural end, offer to print a receipt of
the conversation.

When Matt asks to "print a receipt", "receipt this session", or
`/receipt`, use the Codex `receipt` skill if it is installed. If the
thermal-printer CLI or Codex-compatible session stats are unavailable,
say exactly what is missing and provide the 3 to 5 line receipt text
instead. Do not fake deterministic stats.
