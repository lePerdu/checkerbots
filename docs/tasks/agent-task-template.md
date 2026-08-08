# Task: <short task title>

## Context

Briefly describe where this task fits in the project and any important architectural constraints.

Example:
- Repo: `checkerbots`
- Follow the architecture and backlog docs under `docs/`
- Prefer minimal, focused changes
- Do not introduce new dependencies unless clearly justified

## Goal

State the concrete outcome this task should produce.

## In scope

- Item 1
- Item 2
- Item 3

## Out of scope

- Item 1
- Item 2

## Expected files or areas

List the files, directories, or modules the agent should expect to touch.

Examples:
- `apps/server/`
- `packages/game-engine/`
- `docs/`

## Implementation notes

Include any design constraints, naming guidance, or behavior expectations.

Examples:
- Keep the server as the source of truth for game state
- Use shared types where possible
- Keep v1 assumptions: one game at a time, one moving robot at a time

## Acceptance criteria

- Criterion 1
- Criterion 2
- Criterion 3

## Validation

List the most relevant commands or checks to run.

Examples:
- Run targeted tests for the changed package
- Run lint if project tooling already exists
- If validation cannot be run yet, explain why in the final summary

## Deliverables

List what should exist after the task is done.

Examples:
- New module/file created
- Tests added
- Doc updated

## Notes for the final response

Ask the agent to summarize:
- what changed
- which files were touched
- what validation was run
- any follow-up or known limitations
