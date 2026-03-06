---
name: issue-manager
description: >
  A skill for creating and updating GitHub Issues. Used during the planning phase of feature development to structure Issues with summary, background, implementation approach, impact scope, and Definition of Done.
  Trigger when the user says things like "create an issue", "write an issue", "make a plan", "summarize the requirements", or "make a ticket",
  or as the first step of feature development. Also handles updating existing Issues (requirement changes, DoD additions, scope changes, etc.).
  Includes partial updates by issue number, such as "update Issue #XX" or "add … to Issue #XX".
  Use this skill proactively whenever the user asks to write, structure, or plan Issue content.
  Do NOT use for simple operations like closing Issues, changing labels, or assigning, nor for code implementation, fixes, or reviews.
---

# GitHub Issue Manager

A skill for creating and updating GitHub Issues. Organizes development plans into a structured format within Issues.

## Workflow

### 1. Gather Information

Understand the following from the user's request. Investigate the codebase to fill in any gaps.

- What needs to be achieved
- Why it is needed (background / motivation)
- Preliminary technical approach
- Completion criteria

### 2. Codebase Investigation

Investigate related code to accurately describe the implementation approach and impact scope.

- Identify the components and files to be changed
- Review existing implementation patterns
- Enumerate existing features that may be affected

When investigating, refer to the project's package structure:
- `cmd/sigrok-mcp-server/` — Entrypoint (thin wiring, no tests needed)
- `internal/tools/` — MCP tool handlers; uses `Runner` interface for mock injection
- `internal/sigrok/` — CLI executor and parsers; uses `CommandFactory` test seam and `testdata/` golden files
- `internal/serial/` — Serial port communication; uses `PortOpener` seam
- `internal/devices/` — Device profile registry with embedded JSON
- `internal/config/` — Env-based configuration

### 3. Issue Body Structure

Compose the body following the template below. All sections should be included.
When updating an existing Issue, maintain the same section structure from the template.

```markdown
## Summary

<!-- What will be done. 1-2 sentences, concise. -->

## Background & Motivation

<!-- Why this is needed. User-experience issues, technical problems, etc. -->

## Implementation Approach

<!-- Technical approach. Files, components, and steps to change. -->

## Impact Scope

<!-- Scope of changes, effects on existing features, caveats. -->

## Test Strategy

<!-- Describe the testing approach:
  - Which test seams to use (Runner mock, CommandFactory, PortOpener, etc.)
  - Table-driven test cases to cover
  - Corner cases and boundary conditions
  - Error and exception scenarios (invalid input, timeouts, non-zero exit codes)
  - Silent failure scenarios (nil dependencies, empty output, missing fields)
-->

## DoD (Definition of Done)

- [ ] Completion criterion 1
- [ ] Completion criterion 2
- [ ] Tests cover the happy path
- [ ] Tests cover corner cases and error scenarios
- [ ] No silent failures (errors are surfaced, not swallowed)
- [ ] CI passes
```

#### Template Usage Guidelines

- **Summary**: Keep it brief. Technical terms are fine, but make it immediately clear what will be done
- **Background & Motivation**: Serves as the basis for implementation decisions. A clear "why" prevents scope drift
- **Implementation Approach**: Write concretely based on the codebase investigation. Include file paths and component names
- **Impact Scope**: Important for preventing breakage of existing features. Explicitly stating "no impact" is also valuable
- **Test Strategy**: Describe the testing approach concretely. List test seams to use
  (e.g. `Runner` mock, `CommandFactory`, `PortOpener`). Enumerate corner cases,
  boundary conditions, error scenarios (invalid input, timeouts, non-zero exit),
  and silent failure scenarios (nil dependencies, empty output, fallback behavior).
  Use table-driven tests with `t.Run()` subtests following the project convention.
- **DoD**: Checklist format. Always include tests and CI passing. Add review criteria as needed

### 4. Creating / Updating Issues

#### New Issue

```bash
gh issue create --title "<title>" --body "<body>"
```

- Keep the title concise and descriptive of what will be done
- Add labels with `--label` if applicable

#### Updating an Existing Issue

```bash
gh issue edit <number> --body "<body>"
```

- Before updating, review the current content with `gh issue view <number>`
- Restructure the existing content to match the template sections and add new requirements. Fill in any missing sections (Background & Motivation, Impact Scope, etc.)
- Leave a comment explaining the reason for the change to preserve context

```bash
gh issue comment <number> --body "Requirement change: ..."
```

### 5. Confirmation

After creating or updating, present the Issue URL to the user.
