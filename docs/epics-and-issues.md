# CheckerBots Epics and Issues

This document converts the phased backlog into a more execution-ready set of epics and issues. It is intended to be easy to adapt into GitHub Issues, Linear, Jira, or a markdown project board.

## How to use this document

- Treat **epics** as major workstreams or milestones.
- Treat **issues** as concrete, assignable tasks.
- Each issue includes a short description and acceptance criteria.
- Labels are suggestions only.
- Dependencies are included when sequencing matters.

## Suggested labels

- `backend`
- `frontend`
- `protocol`
- `simulation`
- `hardware`
- `firmware`
- `camera`
- `testing`
- `docs`
- `ops`
- `mvp`
- `later`

---

# Epic 1 - Project foundation and repo setup

## Goal

Establish the repo structure, tooling, and baseline docs so implementation can begin cleanly.

## Issues

### 1.1 Create initial monorepo directory structure
**Labels:** `docs`, `ops`, `mvp`

Create the initial top-level directories for apps, shared packages, and documentation.

**Acceptance criteria**
- Repo contains `apps/`, `packages/`, and `docs/`.
- Directory purpose is briefly documented in `README.md` or an architecture doc.

---

### 1.2 Choose primary server and UI stacks
**Labels:** `docs`, `mvp`

Decide on the implementation stack for the coordination server and web UI.

**Acceptance criteria**
- A documented decision exists for server stack.
- A documented decision exists for UI stack.
- Tradeoffs are briefly captured.

---

### 1.3 Add baseline developer tooling
**Labels:** `ops`, `mvp`

Set up formatting, linting, and any standard project config needed to keep work consistent.

**Acceptance criteria**
- Formatter configuration exists.
- Lint configuration exists.
- Basic ignore files exist where appropriate.
- A developer can run the formatting/lint commands locally.

---

### 1.4 Write initial README
**Labels:** `docs`, `mvp`

Add a repository README that explains the project goal, major components, and intended MVP.

**Acceptance criteria**
- README describes the project vision.
- README lists the major planned components.
- README links to the backlog and epic docs.

---

### 1.5 Write architecture overview document
**Labels:** `docs`, `mvp`

Capture the high-level architecture and responsibilities of each subsystem.

**Acceptance criteria**
- Architecture doc exists under `docs/`.
- It covers server, UI, robots, simulation, and optional camera tracking.
- It identifies the server as the source of truth for game and board state.

---

# Epic 2 - Shared protocol and domain model

## Goal

Define the shared interfaces and types that simulation and physical robots will both use.

## Issues

### 2.1 Define shared robot protocol v0
**Labels:** `protocol`, `backend`, `simulation`, `mvp`
**Dependencies:** Epic 1

Create an initial message protocol covering server-to-robot and robot-to-server communication.

**Acceptance criteria**
- Protocol doc exists or a protocol package/module is created.
- Message shapes are defined for `hello`, `heartbeat`, `telemetry`, `pose_update`, `command_result`, `error`.
- Message shapes are defined for `identify`, `stop`, `move_to_square`, `move_to_pose`, `set_pose`, `ping`.
- Protocol includes required fields for `robot_id`, timestamps, and correlation/command IDs where needed.

---

### 2.2 Define core shared types
**Labels:** `protocol`, `backend`, `mvp`
**Dependencies:** 2.1

Define shared application types for games, robots, calibration, poses, and commands.

**Acceptance criteria**
- Types exist for `Robot`, `Pose`, `Game`, `BoardCalibration`, `Command`, and `PieceAssignment`.
- Pose includes source/confidence metadata.
- Game type supports single active game assumptions.

---

### 2.3 Define board coordinate model
**Labels:** `backend`, `protocol`, `mvp`
**Dependencies:** 2.2

Create the logical and physical board model, including square naming and world-coordinate mapping assumptions.

**Acceptance criteria**
- Board squares have a canonical naming convention.
- Physical coordinates are modeled in millimeters.
- Board model supports square-to-world and world-to-square mapping conceptually.

---

### 2.4 Define move execution state machine
**Labels:** `backend`, `docs`, `mvp`
**Dependencies:** 2.2

Document the states and transitions for processing a requested move from validation through completion or recovery.

**Acceptance criteria**
- State machine includes validation, dispatch, in-progress, completed, and recovery/error states.
- Failure transitions are defined at a high level.

---

# Epic 3 - Checkers rules engine

## Goal

Implement a standalone rules engine that validates and applies checkers moves independently from the UI and robot logic.

## Issues

### 3.1 Create rules engine module/package
**Labels:** `backend`, `testing`, `mvp`
**Dependencies:** Epic 1

Create the module/package that will own checkers rules.

**Acceptance criteria**
- A dedicated module/package exists for checkers logic.
- It is separate from UI and server transport code.

---

### 3.2 Implement `newGame()` board initialization
**Labels:** `backend`, `testing`, `mvp`
**Dependencies:** 3.1, 2.2

Create the initial board state for a standard game of checkers.

**Acceptance criteria**
- `newGame()` returns a valid initial board state.
- Pieces have side and king state.
- Tests verify expected starting arrangement.

---

### 3.3 Implement legal move generation
**Labels:** `backend`, `testing`, `mvp`
**Dependencies:** 3.2

Implement generation of legal moves for the current player.

**Acceptance criteria**
- Simple diagonal moves are supported.
- Capture moves are supported.
- Invalid moves are excluded.
- Tests cover representative legal/illegal cases.

---

### 3.4 Implement move application
**Labels:** `backend`, `testing`, `mvp`
**Dependencies:** 3.3

Apply a valid move to the board state and update turn/game state accordingly.

**Acceptance criteria**
- Moving a piece updates source/destination squares.
- Captures remove the captured piece logically.
- Turn state updates correctly.
- Tests cover standard move and capture paths.

---

### 3.5 Implement king promotion and game-over detection
**Labels:** `backend`, `testing`, `mvp`
**Dependencies:** 3.4

Add support for kinging and determining when the game is over.

**Acceptance criteria**
- Piece promotes when reaching the opposite edge under chosen rules.
- Game-over logic exists for no-move/no-piece situations as defined.
- Tests cover promotion and terminal cases.

---

# Epic 4 - Coordination server MVP skeleton

## Goal

Stand up the coordination server with the minimum game, robot, and realtime plumbing required for the simulation loop.

## Issues

### 4.1 Create coordination server app
**Labels:** `backend`, `mvp`
**Dependencies:** Epic 1

Create the server app and basic startup structure.

**Acceptance criteria**
- Server app runs locally.
- Basic route structure exists.
- Configuration approach is defined.

---

### 4.2 Implement in-memory game store
**Labels:** `backend`, `mvp`
**Dependencies:** 4.1, 3.2

Create an in-memory store for the single active game.

**Acceptance criteria**
- Server can create a game.
- Server can fetch current game state.
- Store supports single-game MVP assumption.

---

### 4.3 Implement in-memory robot registry
**Labels:** `backend`, `protocol`, `mvp`
**Dependencies:** 4.1, 2.1

Track connected robots and their latest known status.

**Acceptance criteria**
- Registry stores robot identity, status, last-seen time, and latest pose.
- Registry can update on heartbeat/telemetry.

---

### 4.4 Add HTTP API for game lifecycle
**Labels:** `backend`, `mvp`
**Dependencies:** 4.2, 3.4

Add minimal game endpoints for create, read, start, and move submission.

**Acceptance criteria**
- Endpoints exist for creating a game, reading current game, starting a game, and submitting a move.
- Move submission validates against the checkers engine.
- Invalid moves return a meaningful error response.

---

### 4.5 Add WebSocket support for UI and robot updates
**Labels:** `backend`, `protocol`, `mvp`
**Dependencies:** 4.1, 4.3

Add realtime event delivery to support live status and board updates.

**Acceptance criteria**
- Clients can connect to a WebSocket endpoint.
- Game and robot updates can be published.
- Event format is documented or type-defined.

---

### 4.6 Implement game and robot state machines
**Labels:** `backend`, `mvp`
**Dependencies:** 4.2, 4.3, 2.4

Implement explicit lifecycle state transitions for game and robot behavior.

**Acceptance criteria**
- Game states include setup, ready, executing, recovery, paused, finished.
- Robot states include offline, pairing, idle, moving, blocked, error, manual.
- Transitions are enforced in code rather than loose flags only.

---

### 4.7 Add structured logging and event history
**Labels:** `backend`, `ops`, `mvp`
**Dependencies:** 4.1

Log commands, state transitions, and telemetry updates for debugging.

**Acceptance criteria**
- Command dispatches are logged.
- Robot status changes are logged.
- Move processing errors are logged.

---

# Epic 5 - Simulated robot service

## Goal

Create a robot simulator that behaves like a real robot from the server's perspective.

## Issues

### 5.1 Create simulated robot app/service
**Labels:** `simulation`, `backend`, `mvp`
**Dependencies:** 2.1, 4.1

Create a standalone process/service representing one or more simulated robots.

**Acceptance criteria**
- Simulated robot service starts locally.
- It can create at least one simulated robot instance.

---

### 5.2 Implement robot handshake and heartbeat
**Labels:** `simulation`, `protocol`, `mvp`
**Dependencies:** 5.1, 4.3

Make simulated robots connect to the server and maintain liveness.

**Acceptance criteria**
- Simulated robot sends `hello` on connect.
- Simulated robot sends periodic `heartbeat` messages.
- Server registry updates last-seen time.

---

### 5.3 Implement `move_to_square` and `stop` command handling
**Labels:** `simulation`, `protocol`, `mvp`
**Dependencies:** 5.2, 2.3

Allow the server to command a simulated robot to move and stop.

**Acceptance criteria**
- Simulated robot can accept a target square.
- Simulated robot transitions into moving state.
- `stop` interrupts motion and reports command result.

---

### 5.4 Emit simulated telemetry and pose updates
**Labels:** `simulation`, `protocol`, `mvp`
**Dependencies:** 5.3

Continuously emit movement status and pose updates during simulated motion.

**Acceptance criteria**
- Pose updates are emitted while moving.
- Final position settles near/on the requested square.
- Server receives and stores updates.

---

### 5.5 Add configurable simulation parameters
**Labels:** `simulation`, `testing`, `mvp`
**Dependencies:** 5.4

Support simulation of speed, latency, drift, and failures.

**Acceptance criteria**
- Speed is configurable.
- Command latency is configurable.
- Pose noise/drift is configurable.
- Failure mode toggles exist for testing.

---

### 5.6 Add simulated failure scenarios
**Labels:** `simulation`, `testing`, `mvp`
**Dependencies:** 5.5

Make it possible to reproduce stuck, timeout, and disconnect behaviors.

**Acceptance criteria**
- At least three failure scenarios can be triggered intentionally.
- Server receives clear failure signals or timeout symptoms.

---

# Epic 6 - Move execution pipeline

## Goal

Connect legal game moves to robot commands and only commit board state when motion completes.

## Issues

### 6.1 Implement piece-to-robot assignment model
**Labels:** `backend`, `protocol`, `mvp`
**Dependencies:** 2.2, 4.3

Track which robot currently represents which game piece.

**Acceptance criteria**
- A mapping exists from logical piece to robot ID.
- Mapping supports unassigned/captured states.

---

### 6.2 Implement command dispatcher
**Labels:** `backend`, `protocol`, `mvp`
**Dependencies:** 4.3, 5.3

Add a server-side mechanism to send commands to the correct robot and track their lifecycle.

**Acceptance criteria**
- Server can issue a command to a registered robot.
- Command IDs can be correlated with results.
- Command status is visible internally.

---

### 6.3 Implement serialized move execution flow
**Labels:** `backend`, `simulation`, `mvp`
**Dependencies:** 6.1, 6.2, 3.4

For v1, process one moving robot at a time and commit the logical move after successful motion.

**Acceptance criteria**
- Move request validates before dispatch.
- Correct robot is selected for the piece being moved.
- Server waits for move completion before updating committed board state.
- A second move cannot start while one is executing.

---

### 6.4 Add command timeout and failure handling
**Labels:** `backend`, `testing`, `mvp`
**Dependencies:** 6.3, 5.6

Handle non-completion by moving the system into recovery instead of silently failing.

**Acceptance criteria**
- Timeout behavior is defined and implemented.
- Failed move execution transitions the game to recovery/attention-needed state.
- UI-visible error payload can be produced.

---

### 6.5 Add integration test for move request to completion
**Labels:** `testing`, `backend`, `simulation`, `mvp`
**Dependencies:** 6.3

Verify the full software loop using the server and simulated robot.

**Acceptance criteria**
- Test covers move submission, robot dispatch, simulated motion, and board update.
- Test verifies game state only changes after movement completes.

---

# Epic 7 - Web UI MVP

## Goal

Provide a browser-based interface for running a game and observing robots.

## Issues

### 7.1 Create web app scaffold
**Labels:** `frontend`, `mvp`
**Dependencies:** Epic 1

Create the initial web UI application.

**Acceptance criteria**
- Web app runs locally.
- Basic routing or page structure exists.

---

### 7.2 Build board view component
**Labels:** `frontend`, `mvp`
**Dependencies:** 7.1, 2.3

Render the checkers board and current logical state.

**Acceptance criteria**
- Board grid renders correctly.
- Pieces render in their current squares.
- Selected or active move state can be displayed.

---

### 7.3 Connect UI to server game API
**Labels:** `frontend`, `backend`, `mvp`
**Dependencies:** 7.1, 4.4

Allow the UI to create/start games and submit move requests.

**Acceptance criteria**
- User can create a game from the UI.
- User can start a game from the UI.
- User can submit a move request from the UI.
- Errors are displayed for invalid moves.

---

### 7.4 Subscribe to realtime game and robot updates
**Labels:** `frontend`, `protocol`, `mvp`
**Dependencies:** 7.1, 4.5

Keep the UI updated as robots move and game state changes.

**Acceptance criteria**
- UI receives WebSocket events.
- Board updates when the game changes.
- Robot status updates render without refresh.

---

### 7.5 Add robot status panel
**Labels:** `frontend`, `mvp`
**Dependencies:** 7.4

Show connected robots, statuses, and latest known positions.

**Acceptance criteria**
- Robot list shows ID, status, and last-seen or pose information.
- Moving/error states are visually distinguishable.

---

### 7.6 Add move history and command feedback
**Labels:** `frontend`, `mvp`
**Dependencies:** 7.4, 6.2

Expose move progress and recent actions to the operator.

**Acceptance criteria**
- User can see recent moves.
- User can see whether a move is pending, complete, or failed.

---

# Epic 8 - Calibration and manual recovery MVP

## Goal

Make it possible to operate the system even with imperfect localization and hardware behavior.

## Issues

### 8.1 Implement board calibration model and API
**Labels:** `backend`, `mvp`
**Dependencies:** 2.3, 4.4

Store board origin, orientation, and square size in the server.

**Acceptance criteria**
- Server stores board calibration.
- API exists to create/update calibration data.
- Board model uses calibration in square-to-world conversion.

---

### 8.2 Implement manual robot pose correction API
**Labels:** `backend`, `protocol`, `mvp`
**Dependencies:** 4.4, 4.3

Allow the operator to correct the server's notion of robot position.

**Acceptance criteria**
- API exists to set robot square or pose manually.
- Robot pose source can be marked as manual.

---

### 8.3 Build calibration UI page
**Labels:** `frontend`, `mvp`
**Dependencies:** 8.1, 7.1

Add a page for entering board dimensions/origin and assigning robots to known squares.

**Acceptance criteria**
- User can set or edit board calibration values.
- User can assign/correct a robot's square from the UI.

---

### 8.4 Build recovery controls
**Labels:** `frontend`, `backend`, `mvp`
**Dependencies:** 8.2, 6.4, 7.4

Add controls for stop-all, retry, and manual correction when a move fails.

**Acceptance criteria**
- User can trigger stop-all.
- User can manually correct robot position.
- User can retry or clear a failed move workflow.

---

### 8.5 Document operator recovery workflow
**Labels:** `docs`, `hardware`, `mvp`
**Dependencies:** 8.4

Write down how a human operator is expected to recover when the robot drifts or gets stuck.

**Acceptance criteria**
- Recovery steps are documented.
- Document covers stop, reposition, set pose, and retry.

---

# Epic 9 - Roomba i3 feasibility and hardware integration path

## Goal

Determine the realistic control path for the first physical robot and reduce hardware uncertainty.

## Issues

### 9.1 Research Roomba i3 control and firmware options
**Labels:** `hardware`, `firmware`, `docs`, `mvp`

Investigate whether the `Roomba i3` can be controlled via stock interfaces, unofficial interfaces, firmware exploits, or external hardware modification.

**Acceptance criteria**
- Findings are documented.
- Potential control paths are listed.
- Risks and unknowns are listed.

---

### 9.2 Decide initial physical integration approach
**Labels:** `hardware`, `docs`, `mvp`
**Dependencies:** 9.1

Choose the first practical path for the real-robot prototype.

**Acceptance criteria**
- A documented decision exists.
- Fallback path is documented if the preferred path fails.

---

### 9.3 Measure physical board constraints using Roomba i3
**Labels:** `hardware`, `mvp`

Determine the minimum practical square size, spacing, and turning clearance.

**Acceptance criteria**
- Robot dimensions are recorded.
- Draft board square size is recorded.
- Notes exist on turning/clearance constraints.

---

### 9.4 Tape out or mock a prototype board area
**Labels:** `hardware`, `mvp`
**Dependencies:** 9.3

Create a rough physical test surface to validate movement feasibility.

**Acceptance criteria**
- A floor or board mock exists.
- Square centers are marked.
- Basic movement feasibility notes are recorded.

---

### 9.5 Define physical piece identity strategy
**Labels:** `hardware`, `docs`, `mvp`

Decide how side, piece identity, and king status will be represented physically.

**Acceptance criteria**
- A documented plan exists for robot/team identification.
- A documented plan exists for king representation.

---

# Epic 10 - Single physical robot adapter MVP

## Goal

Integrate one real robot with the shared protocol and prove server-to-hardware command flow.

## Issues

### 10.1 Create real robot adapter service
**Labels:** `hardware`, `backend`, `protocol`, `mvp`
**Dependencies:** 9.2, 2.1

Create the service/process that bridges the server protocol to the selected robot control mechanism.

**Acceptance criteria**
- Adapter service starts locally.
- Adapter can connect to the server.
- Adapter identifies itself as a robot endpoint.

---

### 10.2 Implement identify and stop behaviors
**Labels:** `hardware`, `protocol`, `mvp`
**Dependencies:** 10.1

Implement basic safe commands first.

**Acceptance criteria**
- `identify` produces an observable physical behavior.
- `stop` safely interrupts motion or command execution.

---

### 10.3 Implement square-to-square motion control
**Labels:** `hardware`, `backend`, `mvp`
**Dependencies:** 10.2, 8.1

Translate target square commands into physical movement behavior.

**Acceptance criteria**
- Adapter accepts a target square.
- Robot can attempt movement to known test squares.
- Results are reported back to the server.

---

### 10.4 Add telemetry and pose reporting from real robot
**Labels:** `hardware`, `protocol`, `mvp`
**Dependencies:** 10.1

Report whatever movement/position data is available from the chosen hardware path.

**Acceptance criteria**
- Adapter reports robot status while moving.
- Adapter reports position/pose estimates if available.
- Unknown or limited telemetry is still normalized into the shared schema.

---

### 10.5 Run repeated movement trials and capture drift metrics
**Labels:** `hardware`, `testing`, `mvp`
**Dependencies:** 10.3, 10.4

Test repeatability on a real board or floor grid.

**Acceptance criteria**
- Multiple repeated move tests are run.
- Notes exist on drift, overshoot, and reliability.
- Findings are documented for later calibration or camera work.

---

# Epic 11 - Camera tracking prototype

## Goal

Introduce external position tracking if self-reported robot pose is insufficient.

## Issues

### 11.1 Select initial camera tracking approach
**Labels:** `camera`, `hardware`, `docs`, `later`

Choose between fiducials, colors, or another tracking approach for the prototype.

**Acceptance criteria**
- Tracking approach is documented.
- Rationale and likely constraints are noted.

---

### 11.2 Create camera tracking service skeleton
**Labels:** `camera`, `backend`, `later`
**Dependencies:** 11.1

Create a separate service/process for camera-based robot localization.

**Acceptance criteria**
- Service scaffold exists.
- Service boundary and output message format are defined.

---

### 11.3 Implement board calibration from camera landmarks
**Labels:** `camera`, `backend`, `later`
**Dependencies:** 11.2

Map image coordinates to board/world coordinates.

**Acceptance criteria**
- Board corner or landmark calibration path exists.
- Output can produce board-relative positions.

---

### 11.4 Implement robot identification and pose extraction
**Labels:** `camera`, `later`
**Dependencies:** 11.2

Detect robots in camera images and estimate position/heading.

**Acceptance criteria**
- At least one robot can be detected and identified.
- Position estimates can be sent to the server.

---

### 11.5 Ingest camera-derived poses into server localization
**Labels:** `backend`, `camera`, `later`
**Dependencies:** 11.4, 4.3

Allow the server to consume camera as another pose source.

**Acceptance criteria**
- Server accepts camera pose updates.
- Pose source/confidence is preserved.
- UI can display camera-derived positions.

---

# Epic 12 - Physical gameplay prototype

## Goal

Demonstrate a limited but real physical checkers experience.

## Issues

### 12.1 Implement captured-piece handling policy
**Labels:** `backend`, `hardware`, `later`
**Dependencies:** 6.3, 9.5

Define and implement what happens to captured pieces/robots in the physical system.

**Acceptance criteria**
- Captured piece behavior is defined.
- Server and UI reflect captured state.
- Physical handling approach is documented.

---

### 12.2 Implement king representation in UI and physical flow
**Labels:** `frontend`, `backend`, `hardware`, `later`
**Dependencies:** 3.5, 9.5

Show and manage promoted pieces.

**Acceptance criteria**
- UI shows king state.
- Server preserves king state through moves.
- Physical representation approach is defined and documented.

---

### 12.3 Add demo scenario with reduced piece count
**Labels:** `backend`, `frontend`, `hardware`, `later`
**Dependencies:** 10.3, 7.4

Create a smaller scenario that is easier to demo reliably than a full 24-piece setup.

**Acceptance criteria**
- A defined reduced-piece demo scenario exists.
- Setup instructions for the scenario are documented.

---

### 12.4 Validate end-to-end physical move flow
**Labels:** `hardware`, `testing`, `later`
**Dependencies:** 12.3, 10.4

Run the full workflow with a real robot and the production UI/server loop.

**Acceptance criteria**
- User can submit a move from the UI.
- Real robot executes the move.
- Server and UI reflect the completed board state.
- Failure path is recoverable.

---

# Epic 13 - Multi-robot scaling

## Goal

Expand from one real robot to multiple managed units.

## Issues

### 13.1 Support multiple robot registrations in server and UI
**Labels:** `backend`, `frontend`, `later`
**Dependencies:** 4.3, 7.5

Ensure the system behaves correctly with several robots connected.

**Acceptance criteria**
- Multiple robots can be displayed and tracked.
- Piece assignments can distinguish robots correctly.

---

### 13.2 Standardize hardware setup for additional robots
**Labels:** `hardware`, `later`

Create a repeatable procedure for preparing additional robots.

**Acceptance criteria**
- Setup checklist exists.
- Identification and connectivity conventions are defined.

---

### 13.3 Add board occupancy and parking-zone modeling
**Labels:** `backend`, `later`
**Dependencies:** 12.1

Model occupied zones more explicitly as the fleet grows.

**Acceptance criteria**
- Board occupancy can distinguish active squares and off-board parking.
- Planner can detect occupied destinations.

---

### 13.4 Add richer fleet health diagnostics
**Labels:** `backend`, `frontend`, `later`
**Dependencies:** 13.1

Improve visibility into multi-robot health and readiness.

**Acceptance criteria**
- UI includes a multi-robot health summary.
- Server exposes useful fleet status data.

---

# Epic 14 - Testing, documentation, and operations

## Goal

Improve confidence, repeatability, and maintainability across the project.

## Issues

### 14.1 Add unit test suite for board math and shared models
**Labels:** `testing`, `backend`, `mvp`
**Dependencies:** 2.2, 2.3

Test foundational coordinate and domain logic.

**Acceptance criteria**
- Board math tests exist.
- Shared model validation tests exist where appropriate.

---

### 14.2 Add integration tests for server and simulated robots
**Labels:** `testing`, `backend`, `simulation`, `mvp`
**Dependencies:** 5.4, 6.3

Exercise the most important server-to-robot workflows automatically.

**Acceptance criteria**
- Test covers robot connect/update behavior.
- Test covers move execution happy path.
- Test covers at least one failure path.

---

### 14.3 Write operator setup guide
**Labels:** `docs`, `hardware`, `mvp`
**Dependencies:** 8.5, 9.4

Document how to set up the board, start services, and begin a game.

**Acceptance criteria**
- Operator guide exists.
- It covers board setup, startup, calibration, and recovery basics.

---

### 14.4 Write hardware experiment log template
**Labels:** `docs`, `hardware`, `mvp`

Create a lightweight format for recording repeated physical tests and findings.

**Acceptance criteria**
- A template file exists under `docs/`.
- It includes date, hardware setup, test procedure, and observations.

---

### 14.5 Add local runbook for all services
**Labels:** `docs`, `ops`, `mvp`
**Dependencies:** 4.1, 5.1, 7.1

Document how to run server, UI, and simulator together during development.

**Acceptance criteria**
- Runbook exists.
- It includes startup order and expected local URLs/processes.

---

# Suggested MVP milestone grouping

If you want to track only the first practical target, group these epics as the MVP:

- Epic 1 - Project foundation and repo setup
- Epic 2 - Shared protocol and domain model
- Epic 3 - Checkers rules engine
- Epic 4 - Coordination server MVP skeleton
- Epic 5 - Simulated robot service
- Epic 6 - Move execution pipeline
- Epic 7 - Web UI MVP
- Epic 8 - Calibration and manual recovery MVP
- Epic 9 - Roomba i3 feasibility and hardware integration path

A strong first milestone would be:

> A user can create a game in the UI, submit a legal move, watch a simulated robot execute it, and recover from a simulated failure.

# Suggested next 10 issues to do first

1. Create initial monorepo directory structure
2. Choose primary server and UI stacks
3. Add baseline developer tooling
4. Write initial README
5. Define shared robot protocol v0
6. Define core shared types
7. Create rules engine module/package
8. Implement `newGame()` board initialization
9. Create coordination server app
10. Create simulated robot app/service
