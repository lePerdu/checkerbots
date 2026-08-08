# Task: Define shared robot protocol v0

## Context

Simulation and physical robot support should use the same protocol from the start. This protocol will define how robots and the coordination server communicate.

Relevant docs:
- `docs/phased-implementation-backlog.md`
- `docs/epics-and-issues.md`

## Goal

Create a first-pass shared robot protocol specification for server-to-robot and robot-to-server communication.

## In scope

- Define message categories and message shapes
- Cover robot registration, liveness, telemetry, pose updates, commands, and command results
- Include required identifiers and timestamps
- Decide whether the protocol is documented first, type-defined first, or both

## Out of scope

- Implementing the server transport layer
- Implementing the simulator
- Designing an advanced versioning system beyond what v0 needs

## Expected files or areas

- `packages/robo-proto/`
- and/or `docs/`
- Suggested files:
  - `docs/protocol-v0.md`
  - `packages/robo-proto/...`

## Implementation notes

- Prefer clarity and MVP usefulness over completeness.
- The protocol should support both simulated robots and real robot adapters.
- Include enough structure to correlate commands and results.
- Keep one-game-at-a-time and one-moving-robot-at-a-time assumptions in mind, but do not hardcode them into the transport layer unnecessarily.

## Acceptance criteria

- Protocol covers `hello`, `heartbeat`, `telemetry`, `pose_update`, `command_result`, and `error`.
- Protocol covers `identify`, `stop`, `move_to_square`, `move_to_pose`, `set_pose`, and `ping`.
- Required fields include robot identity and timestamps where appropriate.
- Command/result correlation is addressed.

## Validation

- Review the protocol for internal consistency.
- If type definitions are included, ensure they are syntactically valid.

## Deliverables

- Protocol v0 doc and/or shared protocol types

## Notes for the final response

Summarize the protocol structure and note any intentionally deferred questions.
