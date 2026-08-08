# Changelog

All notable changes to this project are documented in this file.

## [v0.1.1] - 2026-08-08

### Added
- `memex` skill: user-facing README covering the concept, bootstrap flow, everyday commands, maintenance rhythm, and the agent hook loop.
- `memex` skill: scripted frontmatter conformance pass (checks `type`, `title`, `description`, a valid `status`, a well-formed `generated`, and source presence) as part of the post-write conformance check.
- `memex` skill: conformance check wired into the done condition of every writing branch (sync, capture, verify, rebuild).

## [v0.1.0] - 2026-08-08

Initial commit of the agent-gears CLI.
