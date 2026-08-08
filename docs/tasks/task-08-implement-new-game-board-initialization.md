# Task: Implement `newGame()` board initialization

## Context

The checkers rules engine needs a standard initial board state before move generation and server integration can proceed.

Relevant docs:
- `docs/epics-and-issues.md`
- any shared type definitions already added under `packages/`

## Goal

Implement `newGame()` for standard checkers and add tests for the initial arrangement.

## In scope

- Create the initial board state
- Represent pieces with side and king state
- Set the starting turn
- Add tests for expected initial placement

## Out of scope

- Legal move generation
- Move application
- King promotion logic beyond initial piece state

## Expected files or areas

- `apps/server/game-engine/`
- related tests in the same package/module

## Implementation notes

- Keep the initial state representation clean and easy to consume by the server/UI.
- Use standard checkers starting layout unless project docs already say otherwise.
- Prefer explicit tests over hand-wavy assumptions.

## Acceptance criteria

- `newGame()` returns a valid standard starting board.
- Pieces include side and non-king initial status.
- Starting turn is set.
- Tests verify the expected starting arrangement.

## Validation

- Run targeted tests for the game engine package.
- If no test harness exists yet, add the most minimal one appropriate to the chosen stack.

## Deliverables

- `newGame()` implementation
- Tests for initial board setup

## Notes for the final response

Summarize the chosen board representation and what the tests cover.
