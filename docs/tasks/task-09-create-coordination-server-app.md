# Task: Create coordination server app

## Context

The coordination server will be the source of truth for game state, robot state, calibration, and move execution.

Relevant docs:
- `docs/phased-implementation-backlog.md`
- `docs/epics-and-issues.md`

## Goal

Create the initial coordination server app scaffold with a runnable entry point and basic structure.

## In scope

- Create the server application directory structure
- Add a runnable entry point
- Add a minimal health or root route
- Establish a sensible internal layout for routes, state, and realtime handling

## Out of scope

- Full API implementation
- Robot registry implementation
- Full game lifecycle logic

## Expected files or areas

- `apps/server/`

## Implementation notes

- Keep the app structure simple and aligned with the chosen server stack.
- Favor a layout that can naturally grow to support routes, WebSockets, stores, and command dispatch.
- Avoid premature abstraction.

## Acceptance criteria

- Server app exists and can start locally.
- There is a minimal route or health endpoint.
- Project structure clearly supports later server tasks.

## Validation

- Run the server locally or run the framework's validation/type-check command if appropriate.

## Deliverables

- Initial coordination server app scaffold

## Notes for the final response

Summarize the structure created and how to run the server.
