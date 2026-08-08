# Task: Define core shared types

## Context

The system needs shared domain types so the server, UI, simulator, and future hardware adapters can agree on the shape of core data.

Relevant docs:
- `docs/epics-and-issues.md`
- protocol and stack decision docs if they exist

## Goal

Define the first version of shared types for games, robots, poses, calibration, commands, and piece assignments.

## In scope

- Define types for `Robot`, `Pose`, `Game`, `BoardCalibration`, `Command`, and `PieceAssignment`
- Include pose source/confidence metadata
- Keep types MVP-oriented and consistent with the protocol

## Out of scope

- Full persistence schemas
- UI component models unrelated to core domain data
- Exhaustive modeling of future games beyond checkers

## Expected files or areas

- `packages/protocol/`
- Optional supporting docs under `docs/`

## Implementation notes

- Keep the types practical for both server and UI consumption.
- Prefer minimal but extensible shapes.
- Be consistent with the shared protocol task.

## Acceptance criteria

- Shared types exist for the main domain entities.
- Pose includes source and confidence.
- Game type supports single-game MVP assumptions without making later extension impossible.
- Types are internally consistent with protocol definitions.

## Validation

- If implemented as code, ensure types compile or type-check.
- Review for naming consistency and obvious overlaps.

## Deliverables

- Shared domain type definitions

## Notes for the final response

Summarize the main entities defined and any important modeling choices.
