#!/bin/bash
set -euo pipefail

# Install global npm tools
npm install -g @anthropic-ai/claude-code markdownlint-cli2

# Ensure Claude Code config directory exists with correct ownership
mkdir -p "$HOME/.claude/debug"

# Verify all tools are available
echo "--- Tool verification ---"
sigrok-cli --version
yamllint --version
golangci-lint version
gh --version
node --version
claude --version
markdownlint-cli2 --version

# Check Go module cache volume permissions
if [ ! -w /go/pkg/mod ]; then
  echo "WARNING: Go module cache is not writable. Run: docker volume rm sigrok-mcp-server-go-mod-cache and rebuild the container."
fi

echo "Dev container ready!"
