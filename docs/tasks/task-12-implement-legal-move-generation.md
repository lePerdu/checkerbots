# Task: Implement legal move generation

## Context

After `newGame()` exists, the checkers engine needs legal move generation so the server can validate requested moves.

Relevant docs:
- `docs/epics-and-issues.md`
- existing game engine code and shared types

## Goal

Implement legal move generation for the current player in the checkers rules engine and add tests.

## In scope

- Generate legal non-capturing diagonal moves
- Generate legal capturing moves
- Exclude clearly invalid moves
- Add tests for representative legal and illegal move scenarios

## Out of scope

- Full move application if not already implemented
- Complex multi-jump behavior unless current engine design requires it immediately
- UI/server integration

## Expected files or areas

- `packages/game-engine/`
- related tests in the same package/module

## Implementation notes

- Keep the representation compatible with `newGame()` output.
- If there are rule ambiguities, document them clearly.
- Prefer correctness and readable tests over cleverness.

## Acceptance criteria

- Engine can generate legal moves for the current player.
- Standard diagonal moves are handled.
- Capture moves are handled.
- Tests cover representative legal/illegal cases.

## Validation

- Run targeted tests for the game engine package.

## Deliverables

- Legal move generation implementation
- Tests for move generation

## Notes for the final response

Summarize the move-generation behavior implemented and call out any deferred rule details.
