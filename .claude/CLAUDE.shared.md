## Working with Matt

You're working with me, Matt.
Things I'd like you to do:

### use the AskUserQuestsions Tool

whever you have a questsion or need my input, you must use the AskUserQuestion tool. if you need more room for explanation, write the optiosn first and then ask me which one I'd like with the AskUserQuestsions tool

## Track learnings (Compound Engineering)

After completing work — especially bug fixes, surprising discoveries, pattern validations, or process improvements — add an entry to the Claude Memories.md file. **The path to this file is defined in the context-specific config (CLAUDE.work.md or CLAUDE.personal.md).**

**When to log:**
- **After work:** Any time a solution, mistake, or decision could prevent future rework or inform future sessions.
- **During planning:** When you notice a recurring pattern emerging — e.g. similar architectural choices across projects, repeated estimation pitfalls, problem shapes that keep reappearing. Capture these in the moment, don't wait until the work is done.

Err on the side of logging — compounding only works if you capture the input.

**Categories** (pick one per entry):
- `bug-fix` — a bug was found and resolved; log the root cause pattern
- `pattern` — a validated approach worth reusing
- `decision` — an architectural or design choice and why alternatives were rejected
- `preference` — a user preference or convention discovered during work
- `tool` — a tool, library, or technique that proved effective (or ineffective)
- `process` — a workflow or process insight
- `planning` — a recurring pattern noticed during planning (e.g. "third project in a row where event-driven fits better than REST", "scope estimates for DB migrations are consistently 2x off")

**Entry format:**

```markdown
## YYYY-MM-DD | Project Name | category

**Trigger:** When/how did this come up?

**Learning:** What was discovered, solved, or decided?

**Compounds into:** How does this make future work easier? What should change in process, CLAUDE.md, or tooling?

**Tags:** #tag1 #tag2

---
```

**The compounding step is the most important part.** Don't just log what happened — ask: "Would the system catch this next time?" If the answer is no, the `Compounds into` field should propose a concrete change (update CLAUDE.md, add a lint rule, create a skill, etc.).
