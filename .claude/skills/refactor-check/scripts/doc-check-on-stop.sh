#!/bin/bash
# On-demand hook (via refactor-check skill): blocks Claude from
# finishing if Go files were modified near a CLAUDE.md without
# the CLAUDE.md being updated too.

DIRTY_DIRS=$(git diff --name-only 2>/dev/null | grep '\.go$' | xargs -I{} dirname {} 2>/dev/null | sort -u)
[ -z "$DIRTY_DIRS" ] && exit 0

NEEDS_CHECK=""
for DIR in $DIRTY_DIRS; do
  CHECK="$DIR"
  while [ "$CHECK" != "." ] && [ "$CHECK" != "/" ]; do
    if [ -f "$CHECK/CLAUDE.md" ]; then
      # Check if CLAUDE.md was also modified
      if ! git diff --name-only 2>/dev/null | grep -q "^${CHECK}/CLAUDE.md$"; then
        NEEDS_CHECK="${NEEDS_CHECK} ${CHECK}/CLAUDE.md"
      fi
      break
    fi
    CHECK=$(dirname "$CHECK")
  done
done

NEEDS_CHECK=$(echo "$NEEDS_CHECK" | xargs)
if [ -n "$NEEDS_CHECK" ]; then
  printf '{"decision":"block","reason":"Go files were modified near these CLAUDE.md files: %s — verify they are still accurate or update them before finishing."}' "$NEEDS_CHECK"
  exit 0
fi
exit 0
