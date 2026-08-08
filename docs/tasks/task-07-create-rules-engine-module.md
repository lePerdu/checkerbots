# Task: Create rules engine module/package

## Context

Checkers rules should live in a standalone module, separate from UI and transport code.

Relevant docs:
- `docs/epics-and-issues.md`

## Goal

Create the initial rules engine module/package structure where checkers game logic will live.

## In scope

- Create the package/module directory
- Add a minimal entry point
- Add a placeholder or initial exported API surface for future rules functions
- Add a minimal test setup if appropriate for the chosen stack

## Out of scope

- Full move generation
- Full game logic implementation
- Server integration

## Expected files or areas

- `apps/server/game-engine/`

## Implementation notes

- Keep it separate from server code.
- Shape the module so `newGame`, `getLegalMoves`, `applyMove`, and `isGameOver` can live there later.
- Keep initial exports simple.

## Acceptance criteria

- A dedicated game engine package/module exists.
- It has a clear entry point.
- It is ready to accept rules logic in later tasks.

## Validation

- If package tooling exists, ensure the module can be imported or type-checked.

## Deliverables

- Initial game engine package/module scaffold

## Notes for the final response

Summarize the module structure created and any assumptions about future API shape.
