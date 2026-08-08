# Task: Define board coordinate model

## Context

The project needs a consistent board model to map logical squares to physical positions and support calibration later.

Relevant docs:
- `docs/phased-implementation-backlog.md`
- `docs/epics-and-issues.md`

## Goal

Define the board coordinate model for square naming and physical coordinate mapping.

## In scope

- Define the canonical square naming convention
- Define how board origin, square size, and orientation are represented
- Define the relationship between squares and world coordinates
- Implement types or documentation, depending on current project state

## Out of scope

- Full calibration UI
- Camera-based transforms
- Path planning

## Expected files or areas

- `packages/board-model/`
- and/or `docs/`

## Implementation notes

- Prefer physical units in millimeters.
- Keep the model simple enough for manual calibration in v1.
- Make sure the naming convention is consistent with the rules engine and UI.

## Acceptance criteria

- Square naming convention is clearly defined.
- Board calibration fields are defined.
- The conceptual mapping between squares and world coordinates is documented or type-defined.

## Validation

- Review for consistency with robo-proto/shared types.
- If code is added, ensure it type-checks.

## Deliverables

- Board coordinate model doc and/or package definitions

## Notes for the final response

Summarize the board model decisions and any assumptions left for later calibration work.
