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
mergeable.

**Before the PR exists**, do all the cheap static work — a reviewer should never
be the one to catch this:

1. **Rebase on latest `main`.** Trunk is `main`; never push to `main` directly.
   Stacks are GitHub-native — a dependent PR sets `--base` to its parent branch.
   Don't use Graphite.
2. **Run codegen if it applies** — any `.graphql` change needs root `yarn gen`.
3. **Format, lint, typecheck** the packages you touched, filtered.
4. **Run the focused tests** for those packages. Never run unfiltered `yarn test`
   locally. If you changed behaviour nothing covers, write the test now.

**Then raise it ready for review**, not as a draft, with Why/What/References —
and announce it in Slack (see above).

**Then stay on it** until it's mergeable: watch CI and reviewer comments as they
land and answer every one. Fix what's right. Post a refutation straight to the
thread where a comment is wrong — evidence first, two or three sentences, civil,
never the same comment twice, and never on something you only half understand.
Flag scope creep on the thread and offer the follow-up. Stop and ask when a
comment wants a redesign, contradicts an earlier decision, or touches schema,
auth, or money paths.

Matt should never have to type "check the comments and fix or refute" or "the CI
is failing". If he does, the loop failed — pick it up mid-flight, don't restart.

**Note on drafts:** the backend repo's own `AGENTS.md` tells agents to always open
drafts and never mark a PR ready. Matt has overridden that for his own PRs — that
override is the "unless the user explicitly says so" case the repo rule allows. It
costs more CI per PR (drafts defer build and preview E2E; system, web-unit and
smoke run only after approval), which is the trade he chose. Don't raise ready PRs
on anyone else's behalf.

**Codex difference:** Claude does this through the `shepherd` skill
(`~/.claude/skills/shepherd/`), auto-kicked by a `PostToolUse` hook and using its
`Monitor` tool to wake on CI/comment changes instead of polling. Codex has no
equivalent, so follow the steps above inline with `gh`.

## Writing Comments

Default to no comment. Most comments you're tempted to write are noise; skip them.

Write one only when it earns its place:

- **Explain why, not what.** The code already shows what it does. A comment explains the non-obvious reason: a trade-off, a constraint, a workaround, a surprising business rule. Link the ticket/issue when that's the "why".

The test: would this still help someone reading the code cold in 3 months? If not, cut it.
