# Task: Add WebSocket support for game and robot updates

## Context

The UI and robot/simulator flows need realtime updates for status changes and move progress.

Relevant docs:
- `docs/epics-and-issues.md`
- existing server scaffold and shared protocol/types

## Goal

Add a basic WebSocket layer to the coordination server for publishing game and robot updates.

## In scope

- Add a WebSocket endpoint or equivalent realtime channel
- Support client connection management
- Support publishing game updates
- Support publishing robot updates
- Define or implement a basic event format

## Out of scope

- Full robot command transport if that is being handled separately
- Advanced subscription filtering
- Production-grade scaling concerns

## Expected files or areas

- `apps/server/`
- possibly shared types under `packages/protocol/`

## Implementation notes

- Keep the event model simple.
- Make the format easy for the web UI to consume.
- Reuse shared types where practical.

## Acceptance criteria

- Server exposes a realtime connection path.
- Connected clients can receive game updates.
- Connected clients can receive robot updates.
- Event payload format is documented or type-defined.

## Validation

- Run local validation/type-checking.
- If practical, verify with a minimal client or test.

## Deliverables

- Basic server-side WebSocket/realtime support

## Notes for the final response

Summarize the event model and how clients are expected to connect.
