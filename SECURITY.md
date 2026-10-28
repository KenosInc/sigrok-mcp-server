# Security Policy

## Reporting a vulnerability

Please report security issues privately via
[GitHub Security Advisories](https://github.com/KenosInc/sigrok-mcp-server/security/advisories/new).
Do not open a public issue for anything you believe is exploitable.

We aim to acknowledge reports within 7 days.

## Scope and threat model

This server executes `sigrok-cli` and talks to serial instruments on behalf of
an LLM. Tool calls originate from the model, so a malicious or confused prompt
(including one injected through decoded bus traffic or an instrument's
responses) can translate into commands sent to real hardware. Treat the MCP
client, the model, and anything the model reads as untrusted input.

All tool arguments are validated before they reach `sigrok-cli` (see
`internal/tools/handlers.go`): string arguments against allow-list regexes,
numeric arguments as finite, non-negative numbers within fixed bounds.
`sigrok-cli` is launched with an argument vector, never through a shell.

Prefer the least-privileged container configuration documented in the README
(`--device=/dev/bus/usb` rather than `--privileged`).

## Supply chain

- GitHub Actions are pinned to commit SHAs and updated by Dependabot.
- sigrok, libsigrokdecode and sigrok-cli are built from commit-pinned
  HTTPS checkouts; the pins are `ARG`s at the top of the `Dockerfile`.
- `govulncheck` and `gosec` (via golangci-lint) run in CI.
- Published container images carry SLSA provenance and an SBOM.
