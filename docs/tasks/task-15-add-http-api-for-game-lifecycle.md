# Task: Add HTTP API for game lifecycle

## Context

The server needs minimal HTTP endpoints so the UI can create a game, inspect it, start it, and submit moves.

Relevant docs:
- `docs/epics-and-issues.md`
- existing server scaffold, game store, and game engine

## Goal

Implement a minimal HTTP API for the game lifecycle in the coordination server.

## In scope

- Add endpoint to create a game
- Add endpoint to fetch current game state
- Add endpoint to start the game
- Add endpoint to submit a move request
- Return meaningful errors for invalid requests

## Out of scope

- Full realtime support
- Authentication
- Multi-game routing beyond MVP needs

## Expected files or areas

- `apps/server/`

## Implementation notes

- Keep endpoints simple and consistent.
- Move submission should validate against the rules engine, even if physical execution is not wired in yet.
- If route naming differs from the planning docs, document the chosen convention.

## Acceptance criteria

- API endpoints exist for create, read, start, and move submission.
- Invalid moves produce meaningful error responses.
- Responses are consistent enough for a UI to consume.

## Validation

- Run targeted API tests if added.
- Otherwise start the server and verify routes locally if possible.

## Deliverables

- Minimal game lifecycle HTTP API

## Notes for the final response

Summarize the endpoints added and any assumptions about request/response shapes.
