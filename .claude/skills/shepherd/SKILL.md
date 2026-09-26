---
name: shepherd
description: Raises a PR properly, then drives it to mergeable without being nudged. Runs static checks and a local lizard review first, opens it ready, then answers every comment and CI failure. Use on "raise/open a PR", "fix or refute the comments", "CI is failing", or to get a PR green. --once for factories.
---

# Shepherd

Raises Matt's PR and answers everything that comes back, until it merges.

Mirrors `.codex/skills/shepherd/SKILL.md`; keep both in sync. Tool differences: Claude
asks with `AskUserQuestion` and waits with `Monitor`; Codex uses its structured
user-input tool and polls. No hook starts either copy.

**The bar:** Matt never types "check the comments" or "the CI is failing". If he does,
the loop failed. Pick it up mid-flight; don't restart.

**Rounds are the enemy, not comments.** The failure this kills: 32 replies across 16
threads, none resolved, 14 answered twice, ten commits, no merge.

**Announce at start:** "Shepherding — pre-flight." Then go quiet until you have a
result or a real question. Never skip phase 1 to reach a PR faster.

## Modes

- **Default (interactive):** phase 1, raise, then phase 2 until mergeable.
- **`--once` (factory):** for Arnold and other headless callers. Phase 1, raise, wait
  through one CI window, do one batched fix-and-reply round, then stop. No one is there
  to answer, so never ask: take the recommended option, note the assumption in the
  result, and block only on irreversible or costly calls (see `forge-principles`,
  rule 10). The last line of output is exactly:

  ```
  SHEPHERD_RESULT {"status":"green|waiting|blocked","pr":"<url>","checks":"pass|fail|pending","notes":"<one line>"}
  ```

  `waiting` means a human's move (review, approval gate); `blocked` means a decision
  shepherd may not take alone.

## Repo profile

Detect the base and load a profile before anything else:

```bash
BASE=$(gh repo view --json defaultBranchRef -q .defaultBranchRef.name 2>/dev/null \
  || git symbolic-ref --short refs/remotes/origin/HEAD | sed 's#^origin/##')
gh repo view --json owner,name -q '.owner.login + "/" + .name'
```

If the owner/repo matches a profile, read it and follow it where it is more specific
than this file:

| Remote | Profile |
|---|---|
| `stampedeapp/*` | `references/profile-letsdothis.md` |

With no profile, take the preflight commands (format, lint, typecheck, focused tests,
codegen) from the repo's `AGENTS.md`, `CLAUDE.md`, `CONTRIBUTING.md`, `package.json`
scripts or `Makefile`, in that order. Run the narrowest form each tool offers.

---

# Phase 1 — before anyone sees it

Everything cheap, static and local. The PR is correct and lizard-clean when a human
first opens it.

## 1.1 Rebase on the base

```bash
git fetch origin "$BASE" && git rebase "origin/$BASE"
```

Never push to `$BASE`. Stacks are GitHub-native: a dependent PR sets its base to the
parent's branch (`gh pr create --base <parent-branch>`), rebased bottom upward.

Resolve conflicts in imports, lockfiles and formatting. Stop on logic you and someone
else both changed.

## 1.2 Work out what you touched

`git diff --name-only "origin/$BASE"...HEAD`. Map paths to packages so every later
command runs only on what changed.

## 1.3 Codegen, format, lint, typecheck

Run codegen when its inputs changed, then format, lint and typecheck, scoped to the
touched packages. These are deterministic; failing in CI means phase 1 was skipped.

## 1.4 Focused tests

Run the tests for the touched packages. Broaden only for shared behaviour, and say so.
Changed behaviour with no test covering it: write the test now.

## 1.5 Lizard the local branch

Run `lizard --local "origin/$BASE"` before anything is pushed. It reviews the diff
from git, posts nothing, and prints its verdict in the session.

Fix every critical and major; take trivial nits, drop the rest. Never post a lizard review under Matt's account
onto Matt's PR.

## 1.6 Housekeeping

New files and directories usually need `CODEOWNERS` coverage when the repo has one;
follow nearby entries. Stage only what belongs to this change. Never `git add .`.

## 1.7 Raise it ready

```bash
git push -u origin HEAD
gh pr create --title "<title>" --base "$BASE" --body "..."   # or the parent branch
gh pr ready <number>    # if anything opened it as a draft
```

Body: **Why / What / References**. References not inferable from the branch or commits:
ask (interactive) or leave a `References: none found` line (`--once`).

Ready, not draft, on Matt's own PRs. Don't raise ready PRs on anyone else's behalf.
Then run any post-raise steps the profile names, and go to phase 2.

---

# Phase 2 — until mergeable

Never `sleep`. Use `Monitor` so you're woken when something changes; a reviewer takes
30 to 60 minutes and polling burns the session. In `--once`, wait for the first CI
result (one window, at most 20 minutes), do one round of 2.1 to 2.5, then emit the
result line.

## 2.1 Read the state before you write a word

```bash
gh api graphql -f query='{repository(owner:"<owner>",name:"<repo>"){pullRequest(number:<n>){
  reviewThreads(first:100){nodes{id isResolved isOutdated
    comments(last:20){nodes{id author{login} body createdAt}}}}}}}'
gh pr checks <number>
gh pr view <number> --json reviews,comments
```

Skip a thread when it is resolved, or when your reply is newer than the last reviewer
comment on it. **One answer per thread, ever.** If you fixed it, you cannot later
refute it.

## 2.2 Judge it, then answer it

A comment is an argument, not an instruction, lizard's included. `forge-principles`
(`~/.claude/skills/forge-principles`) is the bar, and a reviewer asking for something it
bans does not move it.

Refute, don't apply: **"add a comment"** (rename instead), **"cast it / disable the
rule"** (fix the type), **"extract a helper / pull in `<lib>`"** (one caller earns no
abstraction), **try/catch** that hides a failure, a **test** that restates the code. A reviewer can be right about the bug and wrong about the fix: take the
finding, refuse the prescription, say so on the thread.

**Generalise every finding before you push.** Same class elsewhere in the diff: fix it
in the same commit.

Four outcomes, none silent:

- **Fix it:** smallest change that satisfies it, re-run 1.3 and 1.4, reply with what
  changed and the SHA, resolve the thread.
- **Refute it:** lead with the file:line, test or behaviour that makes it not apply.
  Concede the half that's right and fix that half. Not certain it's wrong: fix it or
  ask. Reviewer pushes back on a refutation: bring it to Matt.
- **Out of scope:** one line saying so, link an existing ticket if there is one,
  resolve. Never create a ticket; filing one is Matt's call, so ask first.
- **Bring it to Matt** (`AskUserQuestion`): a different design or a rewrite, a
  contradiction of something Matt decided, schema, migrations, auth or money, or a
  refutation on taste rather than evidence. In `--once`, these end the run as `blocked`.

## 2.3 How to post

- **Two sentences, 300 characters:** what changed plus the SHA, or the evidence.
- **One review submission per wake-up**, batching every reply, and **resolve every
  thread you answer** in it (`resolveReviewThread`).
- **One push per wake-up**, and nits never get their own push.

## 2.4 After round one, ask for the whole list

Once round one is resolved, post one comment with the head SHA asking the reviewer for
every remaining blocker in one pass. Ask once.

## 2.5 Fix CI

The profile names the repo's checks and how to treat each. Without one:

- Format, lint, typecheck, build and codegen checks are deterministic: fix them.
- Test checks: root-cause them, 3 attempts max.
- Checks gated on a human approval are not yours. Read the *latest* run only.
- **Read the whole failed log** (`gh run view <id> --log-failed`); the first error is
  usually a symptom. Reproduce locally before pushing.
- **Flake vs bug:** re-run once if it touches nothing you changed, and say so. Twice is
  a bug.
- **Behind the base?** Rebase as in 1.1, push with `--force-with-lease`.

## 2.6 Budgets and stop conditions

- **Three review rounds**, then stop pushing, summarise on the PR, and bring Matt the
  call: split it, merge as-is, or escalate to the reviewer.
- **Three test-fix attempts**, or the same check failing three times: stop and report.
- **90 minutes with no change:** report where it stalled.

Report and end when it's green and answered (re-read `gh pr checks` first;
approval-gated checks don't count against you), or when it's a human's move and you've
said whose. Never end silently, and never claim green you haven't verified.

---

## Reporting

What you fixed, what you refuted and why, what you flagged as out of scope, what's left
and whose move it is. Then the Resources block with the PR URL, per `CLAUDE.shared.md`.
In `--once`, the `SHEPHERD_RESULT` line comes last, after the report.

## Stop, don't push through

- Skipping phase 1 because the change "is small".
- A second answer on a thread you already answered.
- Refuting something you only half understand, or applying one that makes the code worse.
- A third fix for the same failing check.
- A diff bigger than the comment that prompted it, or `git add .`.
