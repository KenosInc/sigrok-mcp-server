# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation Language

All documentation, code comments, commit messages, and PR descriptions in this repository must be written in English.

## Project Overview

sigrok-mcp-server is an MCP (Model Context Protocol) server that wraps `sigrok-cli`, exposing sigrok's signal analysis capabilities to LLMs. It translates MCP tool calls into `sigrok-cli` invocations and returns structured results, enabling LLMs to control logic analyzers, decode protocols, and analyze signals.

Written in Go. Licensed under MIT (Kenos, Inc.).

## Language & Build

- **Language:** Go
- **Module:** `github.com/KenosInc/sigrok-mcp-server`
- **Build:** `go build ./...`
- **Run tests:** `go test ./...`
- **Run single test:** `go test ./... -run TestName`
- **Lint:** `go vet ./...`
- **Tidy dependencies:** `go mod tidy`

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use table-driven tests with `t.Run()` subtests
- Error handling: return errors, don't panic; wrap with `fmt.Errorf("context: %w", err)`
- Naming: follow Go conventions — exported names in PascalCase, unexported in camelCase
- Keep packages small and focused; avoid circular dependencies

## Architecture

- **MCP Server (Go):** Implements the MCP protocol, receives tool calls from LLMs, and delegates to `sigrok-cli`.
- **sigrok-cli wrapper:** Constructs and executes `sigrok-cli` commands, parses output, and returns structured responses. `sigrok-cli` is the sole interface to sigrok — no C bindings or libsigrok dependency.
- **Docker container:** The MCP server runs inside a Docker container with `sigrok-cli` pre-installed.
- **Transport:** MCP communication uses stdio (stdin/stdout JSON-RPC).

## Project Metadata

- **GitHub:** `github.com/KenosInc/sigrok-mcp-server`
- **License:** MIT (Kenos, Inc.)
- **Main branch:** `main`

## Development Environment

- **Dev container:** Use the devcontainer configuration for local development (`devcontainer.json` in `.devcontainer/`).
- **CI/CD:** GitHub Actions. Workflows live in `.github/workflows/`.

## Running

- **Local (devcontainer):** Open in VS Code / IDE with devcontainer support, or use `devcontainer up`.
- **Docker:** `docker build -t sigrok-mcp-server .` then `docker run sigrok-mcp-server`
