# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Beads Issue Tracking

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

### Plan-Beads Linking

Every implementation plan item MUST have a matching beads issue. This ensures:
- Progress is trackable across sessions via `bd list`/`bd ready`
- Dependencies between work items are explicit via `bd dep`
- History survives context compaction (beads persist, plans don't)

**Rules:**
- Create beads issues BEFORE starting implementation
- Use `--notes="Plan: <plan-item-title>"` to link beads back to the plan
- Set dependencies with `bd dep add` matching the plan's dependency graph
- Mark `in_progress` when starting, `bd close` when done
- Epic-level beads group related work items for overview

### Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

1. **File issues for remaining work** — Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) — Tests, linters, builds
3. **Update issue status** — Close finished work, update in-progress items
4. **PUSH TO REMOTE** — This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Verify** — All changes committed AND pushed

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing — that leaves work stranded locally
- NEVER say "ready to push when you are" — YOU must push
- If push fails, resolve and retry until it succeeds

---

## CRITICAL: Destructive Operations

**FORBIDDEN commands** (NEVER run without explicit user request):
- `git reset --hard` - destroys uncommitted changes
- `git clean -f` - deletes untracked files
- `git checkout .` - discards all changes
- `git stash drop` - permanently loses stashed work

**ASK FIRST before running**:
- `git revert` - modifies commit history
- `git push --force` - rewrites remote history
- `git branch -D` - deletes branch permanently

**RESPECT user actions**:
- If the user moved, renamed, or organized files, do NOT undo their changes without asking
- When git status shows files in unexpected locations, ASK before "fixing" them
- The user's file organization choices are intentional

When in doubt about ANY operation that could lose work, ask first.

## Working Guidelines

**Never assume versions or dates:**
- Run `date` to confirm the current year before searching for "latest" anything
- Use `gh release view --repo owner/repo --json tagName` to get actual latest versions
- Use `gh api repos/owner/repo/releases/latest --jq '.tag_name'` as alternative
- Do not guess version numbers from training data — they are likely outdated

**Verify before writing:**
- Test commands (curl, gh, etc.) before putting them in specs or Makefiles
- Check that APIs return expected format before writing parsers

**Clarify before assuming:**
- When requirements are ambiguous or there are multiple valid approaches, use the AskUserQuestion tool to clarify before proceeding

---

## Documentation Requirements

**When adding or modifying configurable features, you MUST update ALL relevant places:**

1. **CLI flags** — Add flag definition with env var fallback in `cmd/cooked/main.go`
2. **`.claude/skills/`** — Update the appropriate skill
3. **README.md** — Update user-facing documentation (configuration, security, usage)

Note: SPEC.md is the build specification, not user documentation.

**Checklist before completing a feature:**
- [ ] Is the CLI flag documented with `--help` text?
- [ ] Is the feature documented in the appropriate skill?
- [ ] Is README.md updated with user-facing documentation?

Undocumented features are incomplete features.

---

## Architecture

cooked is a **rendering proxy** — it fetches raw document URLs and serves them as styled, self-contained HTML. The binary is fully self-contained with all assets embedded via `go:embed`.

### URL Scheme

```
https://cooked.example.com/{upstream_url}
```

The upstream URL is everything after the first `/` in the path. cooked fetches it, detects the file type from the extension, renders it to HTML with styling, and returns it.

### Special Paths

- `GET /` — Landing page with URL input field
- `GET /healthz` — Health check (200 OK)
- `GET /_cooked/` — Reserved namespace for embedded assets (mermaid.min.js, etc.)

### Planned Packages

```
internal/
  server/       # HTTP server, routing, middleware
  render/       # Markdown/code rendering (goldmark, chroma)
  fetch/        # Upstream HTTP client, caching
  rewrite/      # Relative URL rewriting
  sanitize/     # HTML sanitization
  template/     # HTML template execution
```

Packages will be created as code is written. Don't create empty packages.

---

## Directory Structure

```
cmd/cooked/          # Binary entry point (main.go)
internal/            # All Go packages
embed/               # go:embed assets (CSS, JS) — populated by make deps
testdata/            # Test fixtures, golden files
  golden/            # Expected HTML output
  fixtures/          # Input test fixtures
.claude/skills/      # Claude Code skill files
```

---

## Development

Use `make help` to see all available commands:

```bash
make deps        # Download mermaid.js + github-markdown-css into embed/
make build       # Build the cooked binary
make test        # Run all tests
make test-race   # Run tests with race detector
make docker      # Build Docker image
make clean       # Remove binary + downloaded assets
make lint        # Run gitleaks
```

### Pre-commit

Always run before committing:
```bash
make lint
```

---

## Testing

See `testing` skill for patterns (fakes over mocks, golden files for HTML output, fuzzing, goleak, httptest).

### Key patterns

- **Golden files** are the primary pattern for testing HTML rendering output
- **httptest** for integration testing the full request cycle
- **Fuzz tests** for URL parsing, MDX preprocessing, HTML sanitization

```bash
go test ./...           # All tests
go test -race ./...     # With race detection
go test -update ./...   # Regenerate golden files
```

---

## Skills Reference

| Skill | Description |
|-------|-------------|
| `project-layout` | File and directory naming conventions |
| `http-patterns` | Go 1.22+ routing, middleware, graceful shutdown |
| `error-handling` | Go 1.20+ error patterns, slog integration |
| `testing` | Fakes over mocks, golden files, goleak, httptest |
| `logging-config` | Structured JSON logging via slog |
| `licensing` | MIT compatibility, dependency checks, write original code |

Skills are in `.claude/skills/`. Claude loads them automatically when relevant.

---

## Active Technologies

- Go 1.24+ (primary language)
- `github.com/yuin/goldmark` — CommonMark + GFM markdown parser
- `github.com/yuin/goldmark-highlighting` — syntax highlighting (chroma)
- `go.abhg.dev/goldmark/mermaid` — mermaid diagram support (client-side rendering)
- github-markdown-css — GitHub-style CSS (light + dark, embedded)
- mermaid.js — client-side diagram rendering (embedded)
