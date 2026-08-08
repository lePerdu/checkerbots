# Task: Create simulated robot app/service

## Context

The simulator is a first-class part of the MVP and should behave like a robot endpoint from the server's perspective.

Relevant docs:
- `docs/phased-implementation-backlog.md`
- `docs/epics-and-issues.md`
- protocol docs/types if already created

## Goal

Create the initial simulated robot application/service scaffold with a runnable entry point.

## In scope

- Create the simulated robot app directory structure
- Add a runnable entry point
- Establish a sensible layout for robot instances, protocol client code, and movement simulation logic

## Out of scope

- Full handshake implementation
- Full motion simulation
- Server integration beyond scaffold assumptions

## Expected files or areas

- `apps/sim-robot/`

## Implementation notes

- Keep the simulator separate from server code.
- Structure it so a single process can later host one or more simulated robots.
- Avoid implementing protocol logic unless required for a clean scaffold.

## Acceptance criteria

- Simulated robot app exists and can start locally.
- The structure clearly supports later handshake, command, and telemetry work.

## Validation

- Run the app locally or run the framework's validation/type-check command if appropriate.

## Deliverables

- Initial simulated robot app scaffold

## Notes for the final response

Summarize the scaffold structure and any assumptions about future simulator shape.
