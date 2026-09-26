# Codex Work Guidance

Mirror of `.claude/CLAUDE.work.md`: the Let's Do This work rules. Not loaded on
Matt's personal Mac. Codex reads one global file, so on a work machine replace the
`~/.codex/AGENTS.md` symlink with both files joined:
`cat .codex/AGENTS.md .codex/AGENTS.work.md > ~/.codex/AGENTS.md`.
Keep it in sync with the Claude copy.

## Writing Code

`forge-principles` is the quality bar for every line — read the skill rather than
guessing at it (`~/.agents/skills/forge-principles`). The ones that bite most:

- **Economy of means.** Subtraction first. A new dependency, abstraction or config
  surface is denied until it earns its place in one line.
- **Strict by construction.** `any`, `@ts-ignore`, unchecked casts and lint disables
  are banned, not discouraged. Fix the type instead.
- **Root cause over symptom.** No fix before you can name the cause, and the
  regression test is part of the fix.
- **The edges are the work.** Empty, huge, malformed, double-submit, partial failure.
  Anything that can run twice, will.
- **Match the codebase.** New code reads like the code already around it.

### Comments

Default to none. Write a comment **only** in these three cases:

1. **A workaround** — with a link to the upstream issue. No link, no comment.
2. **A rule you can't see from the code** and would break by changing it: billing,
   auth, legal, an external API's undocumented behaviour. Name the source.
3. **A directive that demands a reason** — `eslint-disable`, `ts-expect-error`.

Not on the list means no comment. There is nothing to weigh up, and "but this one
explains why" is not an exception — it is the excuse that got us here. Every other
why (the bug you chased, what you tried first, what changed) goes in the PR
description, which is where it stays true.

Three hard limits on the ones that do qualify:

- **Two lines maximum.** Needs a paragraph? It was never one of the three.
- **Never above a test.** The `describe`/`it` name is the comment. If the case needs
  explaining, fix the name.
- **Never longer than the code it sits on.** Seven lines of JSDoc over a five-line
  regex means the regex needs a name, not a preface.

Deleting noise you find in a file you're already touching is always welcome.

## Pull Request Requirements

Every PR must include the following in its description:

1. **Why** — Why is this code change needed? What problem does it solve or what value does it add?
2. **What** — What has been done? Summarize the changes made.
3. **References** — A Linear ticket link or relevant documentation link.

If any of these are unknown, ask Matt with the structured user-input tool rather than assuming or omitting them.

## Never File a Ticket Without Asking

Do not create Linear issues on your own. Not as a follow-up, not for an out-of-scope
review comment, not for a bug you spotted in passing. Spotting the work and filing it
are two different decisions and only the second one is mine.

If something is worth tracking, say so in one line and ask with the structured user-input tool.
Run `lin issue new` only after I say yes. Reading, searching, and commenting on
existing tickets stays fine.

## Announcing PRs in Slack

Work only. Every PR raised in the `stampedeapp` GitHub org gets announced via the
"Check this PR out" Slack workflow, right after the PR URL exists. Do it automatically —
don't ask first. Never do this for personal repos.

Run it with the `agent-slack` CLI:

```bash
agent-slack workflow run Ft0C17QFND41 \
  --channel C01BYKUTE5Q \
  --field "🔗 PR Link(s)=<full PR url>" \
  --field "Notes=<one sentence saying what the PR does>

<one sarcastic comment about it>"
```

- **Notes is two lines.** First a plain sentence on what the PR does. Then one
  sarcastic comment about it. Aim the sarcasm at the code, the bug, or the
  situation — never at a person.
- **Where it lands:** `#rp-checkout-devs` (`C01BYKUTE5Q`). No group ping on purpose.
- **Fallback if the run fails:** check `agent-slack auth list` shows the
  `lets-dothis` browser credentials (form submission needs xoxc/xoxd), and
  re-read the field titles with `agent-slack workflow get Ft0C17QFND41`. If it
  still fails, hand Matt the shortcut link plus the two lines ready to paste:
  <https://slack.com/shortcuts/Ft0C17QFND41/4020d6fc9430b6d8b988caeafbb5135b>
- Include both the PR URL and the Slack permalink in the Resources block.

## Raising a PR Is The Start Of The Job

Never raise a PR and hand it back. Raising it starts a loop you own until it is
mergeable: rebase, lint, typecheck, focused tests and a **local `lizard` review of
the branch before it is ever pushed**, then CI and every comment after it, fixed or
refuted, until it is green.

**Rounds are the enemy, not comments.** One answer per thread, resolve it in the same
pass, batch replies into one submission, and after the first round ask the reviewer
for every remaining blocker in one go. Three rounds is the budget; then bring the
call to Matt.

**`shepherd` is that loop.** Run the skill rather than improvising a worse version
from memory. It holds the whole procedure and it is the only copy.

Matt should never have to type "check the comments and fix or refute" or "the CI
is failing". If he does, the loop failed. Pick it up mid-flight, don't restart.

Note on drafts: the backend repo's own `AGENTS.md` tells agents to always open
drafts and never mark a PR ready. Matt has overridden that for his own PRs, which is
the "unless the user explicitly says so" case the repo rule allows. It costs more CI
per PR (drafts defer build and preview E2E; system, web-unit and smoke run only after
approval), and that is the trade he chose. Don't raise ready PRs on anyone else's
behalf.

**Tool difference:** Claude runs `~/.claude/skills/shepherd` and waits with its
`Monitor` tool. Codex runs `~/.agents/skills/shepherd` (or `$shepherd`) and polls with
`gh`. Neither side starts it automatically; start it yourself when a PR needs it.
