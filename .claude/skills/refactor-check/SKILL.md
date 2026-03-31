---
name: refactor-check
description: Activate documentation freshness checks during refactoring. Use when renaming files, moving code between files, splitting/merging files, restructuring subsystems, deleting referenced files, or any work that changes file organization.
hooks:
  PostToolUse:
    - matcher: "Edit|Write"
      hooks:
        - type: command
          command: "$CLAUDE_SKILL_DIR/scripts/check-doc-proximity.sh"
  Stop:
    - hooks:
        - type: command
          command: "$CLAUDE_SKILL_DIR/scripts/doc-check-on-stop.sh"
---

# Refactor Check

Refactoring mode active. Two on-demand hooks are now running:

1. **PostToolUse** — after editing structural files (server, config, cache, render, etc.) near a CLAUDE.md, a reminder is injected to check if the CLAUDE.md needs updating.

2. **Stop** — before finishing a response, checks if Go files were modified near a CLAUDE.md that wasn't also modified. Blocks until verified.

## What it checks

- **Structural files**: server*, config*, cache*, client*, rewrite*, sanitize*, markdown*, asciidoc*, mdx*, org*, code*, logging*, landing*
- **Nearest CLAUDE.md**: walks up the directory tree from the edited file
- **Stop gate**: only blocks if Go files changed but the nearby CLAUDE.md didn't

## When the stop hook blocks

The hook lists specific CLAUDE.md files. For each one: read it, verify file references and key types are still accurate, and update if needed. If the CLAUDE.md is already correct, make a no-op edit (e.g., fix whitespace) so the hook passes.
