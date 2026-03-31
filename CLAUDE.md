# CLAUDE.md

## Beads Issue Tracking

This project uses **br** (beads_rust) for issue tracking.

```bash
br ready              # Find available work
br show <id>          # View issue details
br update <id> -s in_progress  # Claim work
br update <id> -s closed       # Complete work
br sync --flush-only  # Flush DB to JSONL
```

### Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

1. **File issues for remaining work** — Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) — Tests, linters, builds
3. **Update issue status** — Close finished work, update in-progress items
4. **PUSH TO REMOTE** — This is MANDATORY:
   ```bash
   git pull --rebase
   br sync --flush-only
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

**MANDATORY: Check latest versions before adding ANY dependency.**
Training data is stale — treat all version knowledge as wrong until verified.

```bash
gh api repos/OWNER/REPO/releases/latest --jq '.tag_name'  # check latest
go get github.com/foo/bar@v1.2.3                           # pin explicitly
```

Never run bare `go get` without checking the latest version first.

**Never assume dates:**
- Run `date` to confirm the current year before searching for "latest" anything

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
make lint        # Run golangci-lint + gitleaks
make lint-go     # Run golangci-lint only
```

### Conventional Commits

This repo uses **release-please** with Conventional Commits. The commit prefix determines
whether a release is triggered. Choose the wrong prefix and you create a spurious release PR.

**Release-triggering prefixes** (use only when shipping user-facing changes):
- `fix:` → patch release (bug fixes in code)
- `feat:` → minor release (new functionality)

**Non-release prefixes** (no release PR created):
- `docs:` — documentation-only changes (README, CLAUDE.md, comments)
- `chore:` — maintenance (deps, CI config, beads sync)
- `ci:` — CI/CD pipeline changes
- `test:` — test-only changes
- `refactor:` — code restructuring with no behavior change
- `style:` — formatting, whitespace

**Decision rule**: Choose the prefix based on **what files changed**, not the intent.
A "fix" to a README typo is `docs:`, not `fix:`. A test-only change is `test:`, not `fix:`.

---

## Active Technologies

- **Go 1.26** — primary language
- **goldmark** — markdown rendering (with mermaid + syntax highlighting)
- **libasciidoc** — AsciiDoc rendering
- **go-org** — Org-mode rendering
- **github-markdown-css + mermaid.js** — embedded assets (no CDN)
