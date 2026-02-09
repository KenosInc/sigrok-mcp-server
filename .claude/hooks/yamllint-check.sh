#!/bin/bash
set -euo pipefail

INPUT=$(cat)
FILE_PATH=$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // empty')

if [[ "$FILE_PATH" =~ \.(yml|yaml)$ ]]; then
  yamllint -c "$CLAUDE_PROJECT_DIR/.yamllint.yml" "$FILE_PATH"
fi
