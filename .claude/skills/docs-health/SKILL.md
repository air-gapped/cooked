---
name: docs-health
description: Audit documentation health — file sizes, nested CLAUDE.md coverage, stale references, skill inventory. Triggers on doc audit, check docs, missing CLAUDE.md, bloated files, documentation drift, stale file references, skill line counts.
---

# Documentation Health Check

## Root CLAUDE.md
!`wc -l < CLAUDE.md | xargs -I{} echo "Lines: {}"`
!`wc -c < CLAUDE.md | xargs -I{} echo "Bytes: {}"`

## Nested CLAUDE.md files
!`find internal/ -name CLAUDE.md -exec wc -l {} + 2>/dev/null || echo "None found"`

## Missing CLAUDE.md (directories with 5+ source files but no CLAUDE.md)
!`for dir in internal/*/; do count=$(find "$dir" -maxdepth 1 -name '*.go' ! -name '*_test.go' 2>/dev/null | wc -l); if [ "$count" -ge 5 ] && [ ! -f "$dir/CLAUDE.md" ]; then echo "MISSING: $dir ($count source files)"; fi; done`

## Skills (lines per SKILL.md)
!`wc -l .claude/skills/*/SKILL.md 2>/dev/null`

## Stale file references in nested CLAUDE.md
!`find internal/ -name CLAUDE.md 2>/dev/null | while read cmd; do dir=$(dirname "$cmd"); grep -oE '[a-z_]+\.go' "$cmd" 2>/dev/null | sort -u | while read f; do if [ ! -f "$dir/$f" ]; then echo "STALE: $cmd references $f (not found)"; fi; done; done`

## Guidelines
- Root CLAUDE.md: keep under 500 lines
- Nested CLAUDE.md: keep under 100 lines each
- Skills: load on-demand, size less critical but prefer <150 lines with references/
- If "MISSING" appears above, create a CLAUDE.md for that directory
- If "STALE" appears above, fix the reference or remove it
