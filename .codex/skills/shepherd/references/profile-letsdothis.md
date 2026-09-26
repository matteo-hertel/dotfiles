# Profile: Let's Do This monorepo (`stampedeapp/*`)

Loaded by shepherd when the remote owner is `stampedeapp`. Work rules for this org
(PR body, Slack announce) live in `CLAUDE.work.md` on Claude and the work section of
`AGENTS.md` on Codex.

## Packages

Turbo filters use the `package.json` name: `services/graphql`→`graph-q-l`,
`services/booking`→`booking-service`, `services/event`→`event-service`,
`services/user`→`user-service`, `services/integration`→`integration-service`,
`apps/web`→`@letsdothis/web`, `apps/gateway`→`gateway`.

## Preflight commands

- Codegen: `.graphql` changed → root `yarn gen`. Other generated inputs (protobuf) →
  `yarn turbo run gen --filter=<package-name>`. `Generation Check` catches a skip.
- Format, lint, typecheck:
  ```bash
  yarn turbo run lint:fix --filter=<package-name>
  yarn turbo run lint typecheck --filter=<package-name> --only
  ```
- Tests: `yarn turbo run test --filter=<package-name> --only`. Never unfiltered
  `yarn test`.

## Post-raise: queue and announce

Fire this as soon as the PR URL exists, without asking. After the PR is up, lizard
arrives from Paul's queue; don't start another lizard run.

```bash
curl -sS -X POST "https://paul-macbook-pro.taild42dc0.ts.net/api/queue" \
  -H "Content-Type: application/json" -d "{\"url\":\"$PR_URL\"}"
```

If it fails (machine asleep, tailnet unreachable), note it in the report and carry on.
Then announce in Slack per the work rules.

## CI checks

| Check | Treat as |
|---|---|
| `Fast Format & Hygiene Checks`, `Lint (Scoped)`, `Code Quality` | Deterministic — fix. Shouldn't fail if phase 1 ran |
| `Typecheck (Scoped)`, `Build (Scoped, Non-web)`, `Generation Check` | Build — fix. `Generation Check` means you skipped `yarn gen` |
| `Unit Tests`, `System Tests`, `Smoke Tests`, `AI evals` | Test — root-cause it, 3 attempts max |
| `Resolve Approval & Scope`, `Report Required Tests` | Not yours — gated on a human. Check the latest run; older runs stay red forever |
| `Mongo Unique Index Data Preflight`, `Configuration Validation` | Read the log — usually a real data or config problem |

## Other

- Ready, not draft: Matt overrode the repo's draft rule for his own PRs. Don't
  re-litigate it.
- New files and directories need `.github/CODEOWNERS` coverage.
- No Graphite; stacks are GitHub-native.
