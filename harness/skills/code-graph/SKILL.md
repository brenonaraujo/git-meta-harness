---
name: code-graph
description: Use a code graph (or semantic search) instead of grep/ls/read to navigate a codebase. Activated by all personas when the codebase is large or unfamiliar.
---

# Skill — code-graph

## What this skill is

This skill teaches a persona **how to use a code graph**
(or, when one is not available, a semantic search index)
to navigate an unfamiliar codebase, instead of relying on
plain `grep` + `ls` + `read` loops.

The skill is **runtime-agnostic**: it does not depend on a
specific tool. It works with any of:

- **Sourcegraph** (Cody MCP) — index-first, enterprise.
- **CodeCompass** (Neo4j) — hybrid graph + retrieval.
- **CodeGraph** (local SQLite + tree-sitter) — hybrid, 100%
  local, no external service.
- **Cursor codebase indexing** — IDE-level semantic search.
- **Agentic search** (Claude Code, Hermes default) — no
  pre-index, just `gh` + `grep` + `read`. This is the
  fallback if nothing else is available.

## When to use

- **Use the code graph** when the codebase is:
  - Large (>50 files) or unfamiliar.
  - Cross-cutting (a feature touches multiple services).
  - Has "hidden dependencies" (a file is structurally
    important but semantically distant from the feature).
- **Use agentic search** when the codebase is:
  - Small (<50 files).
  - Familiar to the team.
  - The persona is doing a local change in a known module.

## The rule of thumb: code_search before grep

When a persona needs to find something in the codebase:

1. **First**, try the **code graph / semantic search**.
   The query is a natural-language question or a symbol
   name. Examples:
   - "where is the user authentication handled?"
   - "what calls `processPayment`?"
   - "what depends on the `users` table?"
2. **Then**, fall back to `gh` + `grep` for exact
   identifiers you already know.

This is the **semantic-first, exact-second** pattern.
It is more token-efficient and faster for unfamiliar
code, but it requires the code graph to be set up.

## How to set up a code graph (per tool)

### Sourcegraph (Cody MCP) — enterprise, index-first

```bash
# 1. Connect the repo to a Sourcegraph instance
# (UI: https://sourcegraph.com → Add repository)

# 2. Install the Sourcegraph MCP in the agentic CLI
# (UI: Cody → Settings → MCP → Add Sourcegraph MCP)

# 3. Use it from the agent
# (CLI: `cody mcp sourcegraph_query "where is the user auth?"`)
```

When the meta-harness team-manager is set up, the
`backend-engineer` and `frontend-engineer` personas can use
the Sourcegraph MCP for cross-repository queries.

### CodeCompass (Neo4j) — hybrid graph + retrieval

```bash
# 1. Install CodeCompass
npx codecompass install

# 2. Index the repo
codecompass index --repo .

# 3. Run the meta-harness with CodeCompass as the code
# graph backend
CODECOMPASS_MCP=1 codeagent
```

CodeCompass is **open-core** and works with any agent that
supports MCP.

### CodeGraph (local SQLite + tree-sitter) — hybrid, 100% local

```bash
# 1. Install CodeGraph
go install github.com/your-org/codegraph@latest

# 2. Index the repo (one-time, fast)
codegraph index --repo .

# 3. Run the meta-harness with CodeGraph as the code graph
# backend
CODEGRAPH_MCP=1 codeagent
```

CodeGraph is **94% fewer tool calls** and **77% faster** than
agentic search alone (per the 2026 benchmarks).

### Cursor codebase indexing — IDE-level semantic search

If the team uses Cursor, the codebase is automatically
indexed. The agent uses the index via `@code` / `@file` /
`@folder` references.

### Agentic search (default) — `gh` + `grep` + `read`

If no code graph is set up, the personas fall back to
**agentic search** (the Claude Code / Hermes default):

```bash
# 1. Find the file
gh api repos/:owner/:repo/git/trees/main --jq '.tree[].path' | grep backend

# 2. Read the file
gh api repos/:owner/:repo/contents/backend/internal/auth.go --jq '.content' | base64 -d

# 3. Search for references
gh search code 'processPayment' --owner :owner --repo :repo --language go
```

This is correct but slow. The code graph exists to make
this faster.

## Examples in the meta-harness

### Example 1: backend-engineer looking for the auth module

```bash
# With a code graph (preferred)
cody mcp sourcegraph_query "where is the auth middleware?"

# Without a code graph (fallback)
gh search code 'auth' --owner myorg --repo myproject --path backend/
```

### Example 2: solutions-architect mapping the dependencies

```bash
# With a code graph (preferred)
codecompass query "what depends on the orders table?"

# Without a code graph (fallback)
gh api repos/myorg/myproject/dependency-graph/snapshots
```

### Example 3: domain-expert understanding the business rules

```bash
# With a code graph (preferred)
cody mcp sourcegraph_query "what is the rate limit logic?"

# Without a code graph (fallback)
gh search code 'rate' --owner myorg --repo myproject --path backend/
```

## Metric: how to know the code graph is helping

Track these 3 metrics in your meta-harness pilot:

| Metric | What it measures | Target |
|---|---|---|
| **TTFRF (Time To First Relevant File)** | How quickly the persona reaches the right file | Lower is better; code graph should cut this in half |
| **Tokens-per-task** | Total tokens spent on a task | Lower is better; code graph should cut this by 50-90% |
| **Staleness incidents** | How often the code graph was out of date | Lower is better; the graph should be re-indexed on every commit |

If the code graph is helping, **TTFRF drops by 50%+** and
**Tokens-per-task drops by 50-90%**. If not, fall back to
agentic search.

## Why this is in the meta-harness

A code graph is **not** a requirement for the meta-harness
to work. The default is agentic search (`gh` + `grep` +
`read`), which is correct and works for any codebase.

A code graph is an **optional accelerator** for large or
unfamiliar codebases. It saves tokens, saves time, and
makes the personas more reliable. The meta-harness
documents the option and the compatible tools, so adopters
can choose.

## See also

- [`docs/HOWTO.md`](../../docs/HOWTO.md) — start a project
  with the meta-harness, including code graph setup.
- [`docs/LOOP.md`](../../docs/LOOP.md) — how the loop
  engineering pattern maps to the meta-harness.
- [`docs/CONCEPT.md`](../../docs/CONCEPT.md) — the
  template-vs-materialized distinction for personas.
