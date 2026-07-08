## Working with Matt

You're working with me, Matt.
Things I'd like you to do:

### Keep Claude and Codex guidance in sync

This dotfiles repo mirrors Claude guidance in `.claude/` and Codex guidance in `.codex/AGENTS.md`. When you change agent instructions or skills on one side, update the other side in the same change. If a behavior is tool-specific and cannot be mirrored exactly, document the difference in both places.

### use the AskUserQuestsions Tool

whever you have a questsion or need my input, you must use the AskUserQuestion tool. if you need more room for explanation, write the optiosn first and then ask me which one I'd like with the AskUserQuestsions tool

### Developing locally

When you're asked to create an html file unless otherwise instructured put it in the tmp folder so we don't pollute the working folder

When creatin html file always open it instead of telling me to open it. Instead of using the open command, spin up a small server and serve the html from it

When spinning up something locally application or small servers for html file don't use the local ip but use the tailscale ip or dns and always report a clear list of the running urls ready for me

### Committing

This is a permissive rule, **not** a "don't commit" rule. Default to committing freely.

- **On any non-main branch or worktree:** commit as many times as you want, without asking me first. This is the normal case — just commit.
- **Only on the `main`/`master` (default) branch:** don't commit unless I explicitly tell you to. Offer to create a branch instead.

Never generalize the main-branch restriction into "I can't commit" or "I'll commit nothing." The restriction applies *only* on the default branch. On a feature branch, committing without asking is the expected behavior.

### Presenting resources back to me

Whenever a reply hands me something actionable — a generated/edited file, a PR or issue, a locally running server/app, a deploy URL — end the message with a **Resources** block as the very last thing, so I never have to scroll to find the link.

Format (always this exact shape):

```
───────────────────────────
📎 Resources
🔗 PR    → <full url>
📄 File  → <absolute path>
🌐 Local → <url>
```

Rules:

- **Always last** in the message, after any prose, tests, or explanation.
- **Only when there's an actual resource.** Don't append an empty or "N/A" block to normal replies.
- **PRs and issues are always full URLs**, never a bare number (`#42` → `https://github.com/<owner>/<repo>/pull/42`). If you only have a number, run `gh` to resolve the URL before answering.
- **Files are absolute paths** so they're clickable in my terminal. Add `:line` when pointing at a specific spot.
- **Local servers/apps use the tailscale host or DNS**, never `localhost`/`127.0.0.1` (see Developing locally).
- Use the matching icon per type: 🔗 PR/issue, 📄 file, 🌐 local URL, 🚀 deploy/live URL, 📦 build artifact. One line per resource; group by type if there are several.
