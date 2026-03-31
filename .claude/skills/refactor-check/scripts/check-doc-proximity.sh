#!/bin/bash
# On-demand hook (via refactor-check skill): nudges when editing
# structural files near a CLAUDE.md.

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
[ -z "$FILE_PATH" ] && exit 0

# Only structural Go files warrant a reminder
BASENAME=$(basename "$FILE_PATH")
echo "$BASENAME" | grep -qE '^(server|config|cache|client|rewrite|sanitize|markdown|asciidoc|mdx|org|code|logging|landing)\.' || exit 0

# Walk up to find nearest CLAUDE.md
DIR=$(dirname "$FILE_PATH")
while [ "$DIR" != "." ] && [ "$DIR" != "/" ]; do
  if [ -f "$DIR/CLAUDE.md" ]; then
    printf '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"Refactor check: you modified a structural file (%s). Check if %s/CLAUDE.md needs updating — especially file references, key types, or extension patterns."}}' "$BASENAME" "$DIR"
    exit 0
  fi
  DIR=$(dirname "$DIR")
done
exit 0
