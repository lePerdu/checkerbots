# Task: Implement in-memory game store

## Context

The coordination server needs a simple MVP store for a single active game before persistence is introduced.

Relevant docs:
- `docs/epics-and-issues.md`
- existing server scaffold and game engine code

## Goal

Add an in-memory game store to the coordination server that supports the single-game MVP assumption.

## In scope

- Create a store for the active game state
- Support creating and retrieving the current game
- Support replacing/updating the game state cleanly

## Out of scope

- Database persistence
- Multi-game support
- Full move execution flow

## Expected files or areas

- `apps/server/`

## Implementation notes

- Keep the API simple.
- Make the single-game assumption explicit in code or documentation.
- Prefer code organization that can later swap in persistence.

## Acceptance criteria

- Server-side store exists for the active game.
- Store supports create/get/update operations needed for MVP.
- Code is organized so later server routes can use it cleanly.

## Validation

- Run relevant tests if added.
- Otherwise run server validation/type-checking if available.

## Deliverables

- In-memory game store implementation

## Notes for the final response

Summarize the store interface and where it lives in the server app.
