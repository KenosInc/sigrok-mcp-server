#!/bin/bash
set -euo pipefail

# Claude Code needs write access to the config and claude directories
chown -R vscode:vscode /home/vscode/.config
chown -R vscode:vscode /home/vscode/.claude

# Claude Code skip onboarding
if [ ! -f "$HOME/.claude.json" ]; then
  cat > "$HOME/.claude.json" << 'EOF'
{
  "hasCompletedOnboarding": true,
  "hasAckedPrivacyPolicy": true,
  "completedOnboardingAt": "2026-02-10T00:00:00.000Z",
  "opusProMigrationComplete": true
}
EOF
fi

# Install global npm tools
npm install -g markdownlint-cli2

# Verify all tools are available
echo "--- Tool verification ---"
sigrok-cli --version
yamllint --version
golangci-lint version
gh --version
node --version
markdownlint-cli2 --version

# Check Go module cache volume permissions
if [ ! -w /go/pkg/mod ]; then
  echo "WARNING: Go module cache is not writable. Run: docker volume rm sigrok-mcp-server-go-mod-cache and rebuild the container."
fi

echo "Dev container ready!"
