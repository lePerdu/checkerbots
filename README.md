# CheckerBots

CheckerBots is a project to make a physical game of checkers playable with robots acting as the pieces.

The current plan is to start with a software-first prototype, prove the game loop in simulation, and then move toward a physical demo using a `Roomba i3` as the initial robot platform. The repository is organized to support both the game software and the real-world experimentation needed to make that possible.

## Project vision

The goal is to build a system where:

- a user can create and manage a checkers game through a web UI
- a coordination server acts as the source of truth for game state and robot state
- robots can be commanded to move between board squares
- the same high-level protocol can support both simulated robots and physical robots
- physical calibration and recovery workflows can be layered in as the project matures

The near-term focus is on getting to a working MVP quickly without depending on full hardware automation from day one.

## Planned major components

The project is expected to center around these major pieces:

- **Coordination server** — owns game state, validates moves, tracks robot status, and dispatches commands
- **Web UI** — provides game setup, board view, status, and recovery workflows
- **Shared protocol and domain packages** — define message formats, board models, and checkers data structures shared across services
- **Checkers rules engine** — validates legal moves and applies game logic independently of transport or UI concerns
- **Simulated robot service** — enables early development and testing without requiring physical robots
- **Physical robot adapter/service** — connects the server to real hardware when the project moves beyond simulation
- **Optional camera tracking** — may later improve localization and recovery, but is not required for the first pass

## MVP priorities

Based on the current planning docs, the early MVP is intentionally narrow:

- support one game at a time
- treat simulation as a first-class requirement
- prove the full loop of creating a game, submitting valid moves, and watching robots move in software
- defer authentication and more advanced operational concerns for v1
- allow manual calibration and recovery workflows in the first physical iterations

The first major milestone is a **Simulation MVP**. A physical single-robot prototype follows after the core protocol, rules, and control flow are validated.

## Repository layout

- `apps/` — applications such as the coordination server, web UI, and robot services
- `packages/` — shared libraries and domain modules
- `docs/` — planning, architecture, backlog, and hardware notes

## Planning documents

For current scope and execution planning, start here:

- [Phased implementation backlog](docs/phased-implementation-backlog.md)
- [Epics and issues](docs/epics-and-issues.md)

These documents are the best source of truth for what the project is expected to deliver next.

## Current status

This repository is in an early planning and scaffolding stage. Some implementation details, stack choices, and hardware decisions are still intentionally undecided.

If you are contributing, prefer changes that keep the simulation path clear, preserve flexibility for the physical robot layer, and avoid locking the project into assumptions that are not yet documented.
