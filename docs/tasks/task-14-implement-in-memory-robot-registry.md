# Task: Implement in-memory robot registry

## Context

The coordination server needs to track connected robots, their latest status, and latest known pose.

Relevant docs:
- `docs/epics-and-issues.md`
- existing protocol/shared types and server scaffold

## Goal

Add an in-memory robot registry to the coordination server for the MVP.

## In scope

- Store robot identity and latest known status
- Store last-seen timestamps
- Store latest known pose
- Support register/get/list/update operations

## Out of scope

- Persistence
- Sophisticated health scoring
- Fleet scheduling logic

## Expected files or areas

- `apps/server/`

## Implementation notes

- Keep the model aligned with shared types and protocol messages.
- Registry should be easy to update from heartbeat and telemetry handlers later.

## Acceptance criteria

- Server-side robot registry exists.
- It can register and update robots.
- It can expose the current set of known robots.

## Validation

- Run targeted tests if added.
- Otherwise run server validation/type-checking if available.

## Deliverables

- In-memory robot registry implementation

## Notes for the final response

Summarize what robot metadata is tracked and how the registry is intended to be used later.
