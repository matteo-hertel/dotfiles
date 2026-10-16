---
name: codebase-walkthrough
description: Explains an unfamiliar codebase in layers, from a TL;DR down to traced request paths with file:line references, and offers to turn it into a clean HTML page. Read-only exploration. Use on "walk me through this codebase", "how does this repo work", "give me a tour", "onboard me", "/codebase-walkthrough".
---

# Codebase walkthrough

Get someone from zero to "I know where everything is" as fast as possible. Explain in layers: each layer goes one level deeper, so the reader can stop at any point and still have something useful.

## Rules

- **Read-only.** Read and search only. Don't run anything and don't edit code.
- **Every claim points to code:** a `path/to/file.ts:42` reference. If you're inferring rather than reading, say so.
- **Explain, don't audit.** Note smells you happen to see in one short list at the end. Finding problems is a different job.
- Prefer the real entry points (`package.json` scripts, server bootstrap, router, schema) over the README. Then say where the README is wrong or out of date.

## How to explore

1. Start with the root manifest and workspace config (`package.json`, `pnpm-workspace.yaml`, turbo or nx config, `tsconfig`, `docker-compose`, CI workflows).
2. List each app or package and what it does in one line.
3. Find how the parts talk to each other: HTTP, gRPC, queues, a shared DB, or imports.
4. Find the data layer: database, ORM, schema, migrations, seed data.
5. Pick the **three most important flows** (usually login/auth, the main read path, the main write path). Trace each one from entry point to storage.
6. Find trust boundaries: where identity is established, where it's checked, where secrets come from.
7. Find how to run it: install, dev, build, test (unit, integration, e2e), typecheck, lint, and what CI runs. Also the env vars, ports and services (Docker, DB) each needs.

## Output in chat (markdown, in this order)

1. **TL;DR**: 3–5 lines. What it is, the stack, the shape, and the one thing that surprised you.
2. **Map**: a table of each app/package, what it does, its key entry file, and what it talks to.
3. **Architecture**: a small text diagram of the parts and how they connect.
4. **Key flows**: for each of the three flows, numbered steps with `file:line`.
5. **Data**: the store, ORM, main tables and how they relate, and where migrations live.
6. **Auth and trust**: where tokens or sessions are created, verified and trusted, and where secrets come from.
7. **Run it**: a commands table (install, dev, test, e2e, typecheck, lint), CI summary, env vars, ports and services needed.
8. **Conventions**: folder layout, error handling, naming, test style.
9. **Things I noticed**: at most 5 bullets. Not an audit.

Keep the whole thing scannable. Headings and tables over paragraphs.

## Then offer the HTML version

End with: "Want this as an HTML page?" If yes, build one self-contained HTML file.

- **Where:** write to `notes/walkthrough.html` if a `notes/` folder exists. Otherwise ask for a path. If you can't open it yourself, give the path so the reader can open it.
- **Progressive detail.** The TL;DR sits at the top in a highlighted box. Each section below goes deeper. Put the deepest material (full flow traces, long tables) inside `<details>` blocks, so the page reads well folded.
- **Structure.** A sticky table of contents. A simple architecture diagram built with inline SVG or CSS boxes and arrows. Flows as numbered steps with monospace `file:line` references. Commands in code blocks with a copy button.
- **Clear design.** A readable system font, a comfortable line length (about 70–80 characters), generous spacing, and one accent colour.
  - Colours are CSS variables, with a dark mode via `prefers-color-scheme`.
  - It works on a phone, with no horizontal scrolling.
  - No external scripts or fonts.
- **Same content as the chat version.** Don't add claims you didn't verify.
