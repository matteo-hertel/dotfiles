## Working with Matt

You're working with me, Matt.
Things I'd like you to do:

### Keep Claude and Codex guidance in sync

This dotfiles repo mirrors Claude guidance in `.claude/` and Codex guidance in `.codex/AGENTS.md`. When you change agent instructions or skills on one side, update the other side in the same change. If a behavior is tool-specific and cannot be mirrored exactly, document the difference in both places.

### How to explain your work to me

Tell me what you did like you'd tell a colleague standing next to you. Not like a design doc.

**Shape:**

1. **Answer first**, one line. What you did, or what's true. No preamble, no restating my question.
2. **Then bullets**, one idea each, with a short bold label so I can scan.
3. **Then stop.** If a sentence doesn't change what I do next, cut it.

**Rules:**

- **Use simple language.** Everyday words, the ones you'd say out loud. If a simpler word exists, use it. Never pick a clever or precise-sounding word when a common one carries the same meaning: "same as before" not "byte-identical", "on purpose" not "deliberately", "best guess" not "best-effort heuristic".
- **One clause per sentence.** Two ideas = two sentences, or two bullets.
- **Verbs, not noun stacks.** "Two replies could overwrite each other", not "a race on the thread mapping".
- **No literary voice.** No rhetorical framing ("The cost you picked:"), no stacked em-dash asides, no building to a point. Say the point.
- **Don't re-explain code I can read.** Name a function once, say what it does in five words, move on.
- **Caveats only if I have to act.** A caveat earns a line when I must decide, deploy, or watch for something. Otherwise it belongs in a code comment or the plan, not in your reply.
- **Length ceiling:** trivial change → 1 line. Normal change → 3–5 bullets. Big or risky change → up to 8 bullets plus a short **Needs from you** list. More than that only if I ask.
- **If I ask "why", go deep.** Still in bullets.

### use the AskUserQuestsions Tool

whever you have a questsion or need my input, you must use the AskUserQuestion tool. if you need more room for explanation, write the optiosn first and then ask me which one I'd like with the AskUserQuestsions tool

### Developing locally

When you're asked to create an html file unless otherwise instructured put it in the tmp folder so we don't pollute the working folder

When creatin html file always open it instead of telling me to open it. Instead of using the open command, spin up a small server and serve the html from it

When spinning up something locally application or small servers for html file don't use the local ip but use the tailscale ip or dns and always report a clear list of the running urls ready for me

### Every conversation runs through maximum-effort

Run the `maximum-effort` skill at the first ask that changes anything, in every
conversation — don't wait for the task to look big. It triages first; S-size work stays
on the main thread, so the cost of always running it is one five-line brief. Then the
cheapest lane that clears the bar, and a lane moves up only on a named failure. A model
or effort I picked by hand always wins over it.

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
