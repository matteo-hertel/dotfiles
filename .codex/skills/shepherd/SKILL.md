---
name: shepherd
description: Raises a PR properly and then drives it to mergeable without being nudged. Does the cheap static work first — rebase on main, format, lint, typecheck, focused tests — then opens a ready-for-review PR and watches it, fixing or refuting lizard and reviewer comments and fixing CI failures as they land. Use whenever a PR is about to be raised ("make a PR", "raise a PR", "open a PR", "submit it"), and whenever the user says "pull the comments and fix or refute", "check the comments and fix em", "check the CI comments", "more comments", "ci is failing", "fix the ci", "the formatting isn't passing", or asks to get a PR green, mergeable, or unblocked.
---

# Shepherd

The author-side counterpart to `lizard`. Lizard reviews other people's PRs; shepherd
raises Matt's and answers everything that comes back.

This mirrors the Claude skill at `.claude/skills/shepherd/SKILL.md`. Keep the
workflows aligned. Document tool-specific differences in both files.

**The bar:** Matt should never have to type "check the comments and fix or refute" or
"the CI is failing". If he does, the loop failed. Pick it up mid-flight, don't restart.

**Announce at start:** "Shepherding — pre-flight." Keep required status updates
brief. Otherwise wait until you have a result or a real question. No step-by-step
narration.

Two phases. **Never skip phase 1 to get to a PR faster** — every failure it catches
would otherwise cost a full CI cycle, a reviewer's attention, and a nudge from Matt.

---

# Phase 1 — Pre-flight, before the PR exists

Everything cheap, static and local happens here. The point is that the PR is already
correct when a human first sees it.

## 1.1 Rebase on latest main

```bash
git fetch origin main
git rebase origin/main
```

Trunk is `main`, PRs target `main`, never push to `main` directly.

**Stacks are GitHub-native** — a dependent PR just sets its base to the parent's
branch (`gh pr create --base <parent-branch>`). Rebase the parent first, then each
child onto its parent, bottom of the stack upward. Do not use Graphite; Matt doesn't.

Conflicts in imports, lockfiles or formatting: resolve them. Conflicts in logic you
and someone else both changed: stop and ask.

## 1.2 Work out what you touched

Turbo filters use the `package.json` `name`, not the directory name:

```bash
git diff --name-only origin/main...HEAD
```

Common mappings (from the repo's `AGENTS.md`): `services/graphql` → `graph-q-l`,
`services/booking` → `booking-service`, `services/event` → `event-service`,
`services/user` → `user-service`, `services/integration` → `integration-service`,
`apps/web` → `@letsdothis/web`, `apps/gateway` → `gateway`.

## 1.3 Codegen, if it applies

- Any `.graphql` change → root `yarn gen`. Not optional, and CI's `Generation Check`
  will catch it if you skip it.
- Protobuf or other non-GraphQL generated inputs → filtered `yarn turbo run gen`.
- No generated inputs changed → no gen needed.

## 1.4 Format, lint, typecheck

```bash
yarn turbo run lint:fix --filter=<package-name>
yarn turbo run lint typecheck --filter=<package-name> --only
```

`Fast Format & Hygiene Checks` is the check that fails on Matt's PRs more than any
other, and it is entirely deterministic. There is no excuse for it failing in CI.

## 1.5 Focused tests

```bash
yarn turbo run test --filter=<package-name> --only
```

**Never run unfiltered `yarn test` locally** — the repo says so explicitly and it
takes forever. Run the tests for the packages you touched. Broaden only when you
changed shared behaviour, and say so when you do.

Tests co-locate with the code. If you changed behaviour and no test covers it, write
one now — a reviewer asking for it later costs a whole round trip.

## 1.6 Housekeeping

- New files or directories usually need `.github/CODEOWNERS` coverage. Follow nearby
  entries.
- Only stage what belongs to this change. Never `git add .`.

## 1.7 Raise it ready

```bash
git push -u origin HEAD
gh pr create --title "<title>" --base <main-or-parent-branch> --body "..."
gh pr ready <number>    # if anything opened it as a draft
```

Body must carry **Why / What / References** per `AGENTS.md`. If References can't be
inferred from the branch name or commits, ask — don't omit it.

**Raise it ready for review, not as a draft.** The repo's `AGENTS.md` tells agents to
always open drafts and never mark ready — Matt has overridden that for his own PRs.
Honour the override, and know what it costs: drafts defer the build and preview-E2E
jobs, and the system, web-unit and smoke suites only run after a human approves. A
ready PR runs all of that immediately. That's the trade Matt chose — faster signal,
more CI. Don't re-litigate it, but don't raise ready PRs on someone else's behalf.

## 1.8 Ping Paul's lizard

Work PRs only — same rule as the Slack announce. `stampedeapp` org PRs get queued;
personal repos never do.

```bash
curl -sS -X POST "https://paul-macbook-pro.taild42dc0.ts.net/api/queue" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"$PR_URL\"}"
```

Fire it right after the PR URL exists, alongside the Slack announce. Don't ask first.

If it fails — Paul's machine asleep, tailnet unreachable — don't block. Note it in the
final report and carry on to phase 2.

Then **announce it in Slack** per `AGENTS.md`, and go to phase 2.

---

# Phase 2 — Monitor until mergeable

Use the recurring monitor or wait mechanism available in the current Codex
environment. If there is no recurring monitor, run a bounded `gh` polling loop in a
long-running exec session, poll no faster than every 30 seconds, and continue that
same session instead of starting fresh commands. Keep Matt updated at least once a
minute while the loop is active. Stop after 90 minutes with no change.

This is the tool-specific difference from Claude. Claude uses its `Monitor` tool and
`PostToolUse` kickoff hook. Codex has no equivalent automatic kickoff hook, so this
skill starts from its trigger description or an explicit `$shepherd` invocation.

Watch both signals:

```bash
gh pr checks <number>                                    # CI
gh pr view <number> --json reviews,comments              # review bodies
gh api repos/{owner}/{repo}/pulls/<number>/comments      # inline threads
```

Keep a `handled` set of comment IDs for the whole loop. Never process an ID twice —
double-handling is what produced the "more comments" churn this skill exists to kill.

## 2.1 Answer every comment

Comments come from lizard, from Graphite's AI Reviews, and from humans. Each one gets
exactly one of three outcomes — none are left silent.

### Fix it

The comment is right. Make the smallest change that satisfies it. Re-run the focused
checks from 1.4 and 1.5 for the affected package before pushing — the same bar as
phase 1, every time. Then reply on the thread with what changed and the commit SHA.
Batch: one push per wake-up, not one per comment.

### Refute it

The comment is wrong, or right in general but not true of this code. **Post the
refutation yourself.** Reply on the thread, then mark it handled.

- **Lead with evidence** — the file and line, the test that already covers it, the
  behaviour that makes the concern not apply. Not "I don't think so."
- **Two or three sentences.** No essay.
- **Plain and civil.** These are Matt's colleagues. Disagree with the claim, never
  the person. No "actually", no scoring points.
- **Concede the half that's right** if they found a real thing but drew the wrong
  conclusion — then fix that half.
- **If you're not certain it's wrong, it is not a refutation.** Fix it, or ask Matt.

Never refute the same comment twice. If a reviewer pushes back on your refutation,
stop and bring it to Matt — two rounds of argument is a judgement call, not a loop.

### Flag it as out of scope

Correct, worth doing, not this PR's job. Say so on the thread in one line and offer
the follow-up. Matt asks for this by name — "if scope creep flag it", "tell the
lizard so". Lizard's `references/scope.md` sets the same bar from the other side.

If a follow-up PR is already open for this work, put it there and link it.

### Bring it to Matt

Stop and use Codex's structured user-input tool when available if a comment asks for
a different design or a rewrite, contradicts something Matt decided earlier,
touches schema, migrations, auth or money paths, or would be refuted on taste rather
than evidence. Ask one concise plain-text question when that tool is unavailable.

## 2.2 Fix CI

| Check | Treat as |
|---|---|
| `Fast Format & Hygiene Checks`, `Lint (Scoped)`, `Code Quality` | Deterministic — fix it. Should never fail if phase 1 ran |
| `Typecheck (Scoped)`, `Build (Scoped, Non-web)`, `Generation Check` | Build — fix. `Generation Check` means you skipped `yarn gen` |
| `Unit Tests`, `System Tests`, `Smoke Tests`, `AI evals` | Test — root-cause it, 3 attempts max |
| `Resolve Approval & Scope`, `Report Required Tests` | **Not yours.** Gated on human approval — needs a person, not a fix |
| `Mongo Unique Index Data Preflight`, `Configuration Validation` | Read the log first — usually a real data or config problem |

- **Read the whole failed log** (`gh run view <id> --log-failed`), not the first line.
  Failures cascade and the first error is often a symptom.
- **Reproduce locally before pushing.** A guessed fix costs another full CI cycle.
- **Three test-fix attempts, hard.** Then stop and report. Don't rationalise one more.
- **Flake vs bug:** if it passes on re-run and touches nothing you changed, re-run
  once and say you did. Twice is a bug.
- **Stayed behind main?** Rebase as in 1.1 and force-push with `--force-with-lease`.

## 2.3 Stop conditions

Report and end the loop when:

- **Green and answered** — checks pass (bar the approval-gated ones) and every comment
  is fixed, refuted or flagged. Re-read `gh pr checks` before claiming this.
- **Waiting on a human** — approval-gated checks, or a refutation awaiting a reply.
  Say whose move it is.
- **A budget ran out** — 3 test retries, or the same check failing 3 times running.
- **A judgement call landed** — anything from "bring it to Matt".
- **90 minutes with no change** — say where it stalled.

Never end silently, and never claim green you haven't verified.

---

## Reporting

At the end: what you fixed, what you refuted and the reasoning you posted, what you
flagged as out of scope, what's left and whose move it is. Then the Resources block
with the PR and Slack URLs, per `AGENTS.md`.

## Red flags — stop, don't push through

- About to skip phase 1 because the change "is small".
- About to refute something you only half understand.
- About to push a third fix for the same failing check.
- A comment asks for a redesign and you're tempted to just do it.
- The diff you're about to push is bigger than the comment that prompted it.
- About to `git add .`.
