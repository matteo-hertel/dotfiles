---
name: codebase-audit
description: Senior-engineer audit of a whole codebase. Finds and ranks real problems across security, performance, reliability, observability, dev tooling/CI and tests, with root causes, impact, the smallest safe fix and the migration path. Read-only; outputs markdown in chat. Use on "audit this codebase", "what's wrong with this repo", "find the risks", "review the whole project", "/codebase-audit".
---

# Codebase audit

Review the whole project like a senior engineer who has to own it tomorrow. Find what's actually wrong, explain why it's wrong, and rank it so the reader knows what to fix first.

## Rules

- **Read-only.** Read and search only. Don't edit files and don't run anything.
- **Output markdown in chat.** Don't write files.
- **Verify every finding by reading the code.** No "might be", no pattern-matched guesses. If you suspect something but can't confirm it, put it under "Suspected, not verified".
- **Root cause, not symptom.** "Feed is slow" is a symptom. "The feed handler runs 3 queries per post in a loop" is a root cause.
- **The smallest fix that keeps current behaviour.** No rewrites. Say if a fix changes an API response, a schema or stored data. If it does, give a migration path.

## What to look for

- **Security:**
  - Injection: SQL via raw or concatenated queries, command injection, path traversal.
  - Broken authn or authz: missing checks, IDOR, trusting client-sent identity.
  - Token handling: decode used instead of verify, no algorithm pinned, no expiry, tokens minted client-side.
  - Hard-coded or fallback secrets, weak password hashing (fast hash, static salt), timing-unsafe comparisons.
  - CORS, XSS, SSRF, missing rate limits, sensitive data in logs, unsigned webhooks, risky dependencies.
- **Performance:**
  - N+1 queries, missing indexes on foreign keys or filters, unbounded queries with no pagination.
  - Sync I/O or crypto on hot paths, sequential awaits that could run in parallel.
  - Caches with no eviction, oversized payloads, needless re-renders.
- **Reliability:**
  - Swallowed errors, unhandled rejections.
  - No timeouts or deadlines on outbound calls, retries without idempotency.
  - Read-then-write races and missing transactions, resource leaks, no graceful shutdown.
- **Observability:** inconsistent error mapping (everything is a 500), internal details leaked to clients, no request IDs or tracing across services, unstructured logs.
- **Dev tooling and CI:**
  - CI missing, or not running typecheck, lint and tests (including e2e).
  - No frozen lockfile install, unpinned runtime version, `strict` off, broken or misleading scripts.
- **Tests:** `.only`/`.skip`, globs that match nothing, tests that can't fail, time- or order-dependent flakiness, untested critical paths (auth, payments, data writes).

## Severity

- **Critical:** exploitable now, or causes data loss or a cross-user leak.
- **High:** a serious risk under realistic conditions or normal growth.
- **Medium:** real but contained, or needs unusual conditions.
- **Low:** hygiene and maintainability.

Rank by severity first, then confidence, then lowest effort (S < 30 min, M < 2 h, L > 2 h).

## Output (markdown, in this order)

**Summary:** 3 lines. The overall state, the single biggest risk, and what to fix first.

**Ranked findings**

| # | Issue | Area | Severity | Effort | Where |
|---|---|---|---|---|---|

**Details**, one block per finding, in rank order:

```
### 1. <short title>  (<Severity>, <Effort>)
- Where: path/file.ts:42 (and any related spots)
- Root cause: <why it happens>
- Impact: <a concrete exploit or failure scenario, with who is affected and how badly>
- Fix: <the smallest change that keeps current behaviour>
- Migration: <how to roll it out safely if it touches data, tokens, APIs or clients; else "none">
- Test: <the regression test that would have caught it>
```

**Quick wins:** low-effort items worth doing straight away.

**Suspected, not verified:** things that look wrong but need running code or more context to confirm.

**How to check your fixes:** the exact commands to run tests, typecheck, lint and e2e, plus anything that needs a DB or Docker first.
