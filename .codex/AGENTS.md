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

## Writing Code

`forge-principles` is the quality bar for every line - read the skill
rather than guessing at it (`~/.agents/skills/forge-principles`). The
ones that bite most:

- **Economy of means.** Subtraction first. A new dependency, abstraction
  or config surface is denied until it earns its place in one line.
- **Strict by construction.** `any`, `@ts-ignore`, unchecked casts and
  lint disables are banned, not discouraged. Fix the type instead.
- **Root cause over symptom.** No fix before you can name the cause, and
  the regression test is part of the fix.
- **The edges are the work.** Empty, huge, malformed, double-submit,
  partial failure. Anything that can run twice, will.
- **Match the codebase.** New code reads like the code already around it.

### Comments

Default to none. Write a comment **only** in these three cases:

1. **A workaround** - with a link to the upstream issue. No link, no
   comment.
2. **A rule you cannot see from the code** and would break by changing
   it: billing, auth, legal, an external API's undocumented behaviour.
   Name the source.
3. **A directive that demands a reason** - `eslint-disable`,
   `ts-expect-error`.

Not on the list means no comment. There is nothing to weigh up, and "but
this one explains why" is not an exception - it is the excuse that got us
here. Every other why (the bug you chased, what you tried first, what
changed) goes in the PR description, which is where it stays true.

Three hard limits on the ones that do qualify:

- **Two lines maximum.** Needs a paragraph? It was never one of the
  three.
- **Never above a test.** The `describe`/`it` name is the comment. If the
  case needs explaining, fix the name.
- **Never longer than the code it sits on.** Seven lines of JSDoc over a
  five-line regex means the regex needs a name, not a preface.

Deleting noise you find in a file you are already touching is always
welcome.

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

### Announcing PRs In Slack

Work only. Every PR raised in the `stampedeapp` GitHub org gets announced
via the "FULL NOVA PRs" Slack workflow, right after the PR URL exists. Do
it automatically — do not ask first. Never do this for personal repos.

Run it with the `agent-slack` CLI:

```bash
agent-slack workflow run Ft0A2N2ZRJ6S \
  --channel C0A2NJPNTPF \
  --field "🔗 PR Link(s)=<full PR url>" \
  --field "Notes=<one sentence saying what the PR does>

<one sarcastic comment about it>"
```

- Notes is two lines. First a plain sentence on what the PR does. Then one
  sarcastic comment about it. Aim the sarcasm at the code, the bug, or the
  situation — never at a person.
- Where it lands: `#nova-devs` (`C0A2NJPNTPF`), pinging `@nova-engineers`.
- Fallback if the run fails: check `agent-slack auth list` shows the
  `lets-dothis` browser credentials (form submission needs xoxc/xoxd), and
  re-read the field titles with `agent-slack workflow get Ft0A2N2ZRJ6S`. If
  it still fails, hand Matt the shortcut link plus the two lines ready to
  paste:
  <https://slack.com/shortcuts/Ft0A2N2ZRJ6S/0fc0a60851adf0ddf1f68af6918b5a08>
- Include both the PR URL and the Slack permalink in the Resources block.

### Raising A PR Is The Start Of The Job

Never raise a PR and hand it back. Raising it starts a loop you own until it is
mergeable: the cheap static work before the PR exists, then CI and every comment
after it, fixed or refuted, until it is green.

**`shepherd` is that loop.** Run the skill (`~/.agents/skills/shepherd`, or an
explicit `$shepherd`) — do not improvise a worse version of it from memory. It
holds the whole procedure and it is the only copy. Codex has no automatic
kickoff hook, so start it yourself; Claude's copy lives at
`~/.claude/skills/shepherd` and is auto-kicked by a `PostToolUse` hook.

Matt should never have to type "check the comments and fix or refute" or "the CI
is failing". If he does, the loop failed — pick it up mid-flight, do not restart.

Note on drafts: the backend repo's own `AGENTS.md` tells agents to always open
drafts and never mark a PR ready. Matt has overridden that for his own PRs — that
override is the "unless the user explicitly says so" case the repo rule allows. It
costs more CI per PR (drafts defer build and preview E2E; system, web-unit and
smoke run only after approval), which is the trade he chose. Do not raise ready
PRs on anyone else's behalf.

**Tool difference:** Claude uses `~/.claude/skills/shepherd`, auto-kicked by a
`PostToolUse` hook, and its `Monitor` tool. Codex uses
`~/.agents/skills/shepherd`, starts from the skill trigger or an explicit
`$shepherd` invocation, and uses recurring wait/monitor support or a bounded
`gh` polling session. Codex has no equivalent automatic kickoff hook.

## Receipt Printer

When a conversation comes to a natural end, offer to print a receipt of
the conversation.

When Matt asks to "print a receipt", "receipt this session", or
`/receipt`, use the Codex `receipt` skill if it is installed. If the
thermal-printer CLI or Codex-compatible session stats are unavailable,
say exactly what is missing and provide the 3 to 5 line receipt text
instead. Do not fake deterministic stats.
