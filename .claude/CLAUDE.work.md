## Writing Code

`forge-principles` is the quality bar for every line — read the skill rather than
guessing at it (`~/.claude/skills/forge-principles`). The ones that bite most:

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

If any of these are unknown, use `AskUserQuestion` to ask the user rather than assuming or omitting them.

## Announcing PRs in Slack

Work only. Every PR raised in the `stampedeapp` GitHub org gets announced via the
"FULL NOVA PRs" Slack workflow, right after the PR URL exists. Do it automatically —
don't ask first. Never do this for personal repos.

Run it with the `agent-slack` CLI:

```bash
agent-slack workflow run Ft0A2N2ZRJ6S \
  --channel C0A2NJPNTPF \
  --field "🔗 PR Link(s)=<full PR url>" \
  --field "Notes=<one sentence saying what the PR does>

<one sarcastic comment about it>"
```

- **Notes is two lines.** First a plain sentence on what the PR does. Then one
  sarcastic comment about it. Aim the sarcasm at the code, the bug, or the
  situation — never at a person.
- **Where it lands:** `#nova-devs` (`C0A2NJPNTPF`), pinging `@nova-engineers`.
- **Fallback if the run fails:** check `agent-slack auth list` shows the
  `lets-dothis` browser credentials (form submission needs xoxc/xoxd), and
  re-read the field titles with `agent-slack workflow get Ft0A2N2ZRJ6S`. If it
  still fails, hand Matt the shortcut link plus the two lines ready to paste:
  <https://slack.com/shortcuts/Ft0A2N2ZRJ6S/0fc0a60851adf0ddf1f68af6918b5a08>
- Include both the PR URL and the Slack permalink in the Resources block.

## Raising a PR Is The Start Of The Job

Never raise a PR and hand it back. Raising it starts a loop you own until it's
mergeable: the cheap static work before the PR exists, then CI and every comment
after it, fixed or refuted, until it's green.

**`shepherd` is that loop.** Run the skill — don't improvise a worse version of it
from memory. It holds the whole procedure and it is the only copy.

Matt should never have to type "check the comments and fix or refute" or "the CI
is failing". If he does, the loop failed — pick it up mid-flight, don't restart.

**Tool difference:** Claude uses `~/.claude/skills/shepherd`, auto-kicked by a
`PostToolUse` hook, plus its `Monitor` tool. Codex uses `~/.agents/skills/shepherd`,
started by the skill trigger or an explicit `$shepherd`, and polls with `gh`. Codex
has no automatic kickoff.
