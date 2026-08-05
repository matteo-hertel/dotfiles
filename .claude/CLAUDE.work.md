## Pull Request Requirements

Every PR must include the following in its description:

1. **Why** — Why is this code change needed? What problem does it solve or what value does it add?
2. **What** — What has been done? Summarize the changes made.
3. **References** — A Linear ticket link or relevant documentation link.

If any of these are unknown, use `AskUserQuestion` to ask the user rather than assuming or omitting them.

## Writing Comments

Default to no comment. Most comments you're tempted to write are noise; skip them.

Write one only when it earns its place:

- **Explain why, not what.** The code already shows what it does. A comment explains the non-obvious reason: a trade-off, a constraint, a workaround, a surprising business rule. Link the ticket/issue when that's the "why".

The test: would this still help someone reading the code cold in 3 months? If not, cut it.
