#!/bin/bash
set -euo pipefail

INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

if [[ "$FILE_PATH" =~ \.(yml|yaml)$ ]]; then
  yamllint "$FILE_PATH"
fi
