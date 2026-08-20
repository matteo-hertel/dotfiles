---
name: shepherd
description: Raises a PR properly and then drives it to mergeable without being nudged. Runs the cheap static work and a local lizard review before the PR exists, then opens it ready and answers every comment and CI failure until it merges. Use whenever a PR is about to be raised ("make a PR", "raise a PR", "open a PR", "submit it"), and whenever the user says "pull the comments and fix or refute", "check the comments and fix em", "check the CI comments", "more comments", "ci is failing", "fix the ci", "the formatting isn't passing", or asks to get a PR green, mergeable, or unblocked.
---

# Shepherd

Raises Matt's PR and answers everything that comes back, until it merges.

Mirrors the Codex skill at `.codex/skills/shepherd/SKILL.md`. Keep both in sync, and
document tool-specific differences in both.

**The bar:** Matt never types "check the comments" or "the CI is failing". If he does,
the loop failed — pick it up mid-flight, don't restart.

**The failure this exists to kill:** one PR, 32 replies across 16 threads, none
resolved, 14 answered twice and two of those contradicting the first answer. Ten
commits, no merge. **Rounds are the enemy, not comments.**

**Announce at start:** "Shepherding — pre-flight." Then go quiet until you have a
result or a real question.

Never skip phase 1 to reach a PR faster.

---

# Phase 1 — before anyone sees it

Everything cheap, static and local. The PR is correct and lizard-clean when a human
first opens it.

## 1.1 Rebase on main

```bash
git fetch origin main && git rebase origin/main
```

Trunk is `main`; never push to it. Stacks are GitHub-native — a dependent PR sets its
base to the parent's branch (`gh pr create --base <parent-branch>`), rebased bottom
upward. No Graphite.

Resolve conflicts in imports, lockfiles and formatting. Stop and ask on logic you and
someone else both changed.

## 1.2 Work out what you touched

`git diff --name-only origin/main...HEAD`. Turbo filters use the `package.json` name:
`services/graphql`→`graph-q-l`, `services/booking`→`booking-service`,
`services/event`→`event-service`, `services/user`→`user-service`,
`services/integration`→`integration-service`, `apps/web`→`@letsdothis/web`,
`apps/gateway`→`gateway`.

## 1.3 Codegen

`.graphql` changed → root `yarn gen`. Other generated inputs (protobuf) → filtered
`yarn turbo run gen`. `Generation Check` catches it if you skip.

## 1.4 Format, lint, typecheck

```bash
yarn turbo run lint:fix --filter=<package-name>
yarn turbo run lint typecheck --filter=<package-name> --only
```

`Fast Format & Hygiene Checks` is deterministic. It failing in CI is inexcusable.

## 1.5 Focused tests

```bash
yarn turbo run test --filter=<package-name> --only
```

Never unfiltered `yarn test`. Broaden only for shared behaviour, and say so. Changed
behaviour with no test covering it → write the test now.

## 1.6 Lizard the local branch

Run the `lizard` skill against the working branch **before anything is pushed**. No PR
exists: the diff is `origin/main...HEAD`, lizard reads the current checkout, nothing is
posted. Same triage, same bar, findings land in the session.

Fix every critical and major; take trivial nits, drop the rest. A nit caught here costs
one edit instead of a review round.

This is the only lizard run you start — after the PR is up it arrives from Paul's queue
(1.9). Never post a lizard review under Matt's account onto Matt's PR.

## 1.7 Housekeeping

New files and directories usually need `.github/CODEOWNERS` coverage — follow nearby
entries. Stage only what belongs to this change. Never `git add .`.

## 1.8 Raise it ready

```bash
git push -u origin HEAD
gh pr create --title "<title>" --base <main-or-parent-branch> --body "..."
gh pr ready <number>    # if anything opened it as a draft
```

Body carries **Why / What / References** per `CLAUDE.work.md`. References not
inferable from the branch or commits → ask, don't omit.

Ready, not draft — Matt overrode the repo's draft rule for his own PRs, buying full
CI now instead of after approval. Don't re-litigate it, and don't raise ready PRs on
anyone else's behalf.

## 1.9 Queue and announce

Work PRs only — `stampedeapp` org, never personal repos. Fire this as soon as the URL
exists, without asking:

```bash
curl -sS -X POST "https://paul-macbook-pro.taild42dc0.ts.net/api/queue" \
  -H "Content-Type: application/json" -d "{\"url\":\"$PR_URL\"}"
```

If it fails — machine asleep, tailnet unreachable — note it in the report and carry
on. Then announce in Slack per `CLAUDE.work.md`, and go to phase 2.

---

# Phase 2 — until mergeable

Never `sleep`. Foreground sleep is blocked in this harness and polling burns the
session for the 30–60 minutes a reviewer takes. Use `Monitor` so you're woken when
something changes.

## 2.1 Read the state before you write a word

Thread state on GitHub is the record of what's handled — not your memory, which dies
with the session and is what produced the duplicate replies.

```bash
gh api graphql -f query='{repository(owner:"<owner>",name:"<repo>"){pullRequest(number:<n>){
  reviewThreads(first:100){nodes{id isResolved isOutdated
    comments(last:20){nodes{id author{login} body createdAt}}}}}}}'
gh pr checks <number>
gh pr view <number> --json reviews,comments
```

Skip a thread when it is resolved, or when your reply is newer than the last reviewer
comment on it.

**One answer per thread, ever.** Never a second reply, and never one that contradicts
the first — if you fixed it, you cannot later refute it.

## 2.2 Judge it, then answer it

A comment is an argument, not an instruction — lizard's included. `forge-principles`
(`~/.claude/skills/forge-principles`) is the bar and a reviewer asking for something it
bans does not move it. "Good catch, fixed" that lands a worse line is a failure.

Refute these, don't apply them: **"add a comment / JSDoc"** (rename the code instead),
**"cast it / use `any` / disable the rule"** (fix the type), **"extract a helper, add
an option, pull in `<lib>`"** (one caller earns no abstraction), **"wrap it in
try/catch"** where that hides the failure, **"add a test"** that restates the
implementation.

A reviewer can be right about the bug and wrong about the fix. Take the finding, refuse
the prescription, say so on the thread.

**Generalise every finding before you push.** Same class elsewhere in the diff → fix
it in the same commit. Next round's comments are usually this round's finding one file
over, and that round is avoidable.

Four outcomes, none silent:

**Fix it** — smallest change that satisfies it, re-run 1.4 and 1.5 for the package,
reply with what changed and the SHA, resolve the thread.

**Refute it** — lead with the file:line, the test, the behaviour that makes it not
apply. Concede the half that's right and fix that half. Not certain it's wrong → not a
refutation: fix it or ask Matt. Reply, resolve. Reviewer pushes back on a refutation →
stop and bring it to Matt; two rounds of argument is a judgement call, not a loop.

**Out of scope** — correct, worth doing, not this PR's job. One line, offer the
follow-up, link it if one is already open. Resolve.

**Bring it to Matt** — `AskUserQuestion` when a comment asks for a different design or
a rewrite, contradicts something Matt decided, touches schema, migrations, auth or
money, or when you'd be refuting on taste rather than evidence.

## 2.3 How to post

- **Two sentences, 300 characters.** What changed plus the SHA, or the evidence plus
  why it doesn't apply. No praising the catch, no narrating what you nearly did — every
  extra sentence is another thing to argue with.
- **One review submission per wake-up**, batching every reply. Thirty-two separate
  submissions is thirty-two pings at a colleague.
- **Resolve every thread you answer** in the same wake-up (`resolveReviewThread`).
  Answered-but-open reads as unhandled and buys another round.
- **One push per wake-up**, and **nits never get their own push** — a P3 / 💅 / nit
  rides along in the next real commit or gets declined.
- **Plain and civil.** Disagree with the claim, never the person.

## 2.4 After round one, ask for the whole list

Once round one is answered and resolved, post one comment asking the reviewer for a
complete pass: give the head SHA, say all threads are answered, ask for every remaining
blocker in one go. Progressive disclosure — three findings, then three more, then one —
is what turns a PR into a week. Ask once, and don't repeat the ask.

## 2.5 Fix CI

| Check | Treat as |
|---|---|
| `Fast Format & Hygiene Checks`, `Lint (Scoped)`, `Code Quality` | Deterministic — fix. Shouldn't fail if phase 1 ran |
| `Typecheck (Scoped)`, `Build (Scoped, Non-web)`, `Generation Check` | Build — fix. `Generation Check` means you skipped `yarn gen` |
| `Unit Tests`, `System Tests`, `Smoke Tests`, `AI evals` | Test — root-cause it, 3 attempts max |
| `Resolve Approval & Scope`, `Report Required Tests` | **Not yours** — gated on a human. Check the *latest* run; older runs stay red forever |
| `Mongo Unique Index Data Preflight`, `Configuration Validation` | Read the log — usually a real data or config problem |

- **Read the whole failed log** (`gh run view <id> --log-failed`) — failures cascade
  and the first error is usually a symptom. Reproduce locally before pushing.
- **Flake vs bug:** passes on re-run and touches nothing you changed → re-run once and
  say you did. Twice is a bug.
- **Behind main?** Rebase as in 1.1, force-push with `--force-with-lease`.

## 2.6 Budgets and stop conditions

- **Three review rounds.** Then stop pushing, summarise on the PR, and bring Matt the
  call with `AskUserQuestion`: split it, merge as-is, or escalate to the reviewer.
- **Three test-fix attempts**, or the same check failing three times → stop and report.
- **90 minutes with no change** → report where it stalled.

Report and end when it's green and answered — re-read `gh pr checks` before claiming
that, and approval-gated checks don't count against you — or when it's waiting on a
human and you've said whose move it is.

Never end silently, and never claim green you haven't verified.

---

## Reporting

What you fixed, what you refuted and why, what you flagged as out of scope, what's left
and whose move it is. Then the Resources block with the PR URL, per `CLAUDE.shared.md`.

## Stop, don't push through

- Skipping phase 1 because the change "is small".
- A second answer on a thread you already answered.
- Refuting something you only half understand, or applying one that makes the code worse.
- A third fix for the same failing check.
- A diff bigger than the comment that prompted it, or `git add .`.
