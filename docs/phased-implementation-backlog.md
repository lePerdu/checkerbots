# CheckerBots Phased Implementation Backlog

This document turns the high-level project vision into a concrete, phased backlog. It includes both software tasks and real-world/hardware tasks, with an emphasis on getting to a working prototype quickly while reducing hardware risk.

## Planning assumptions

- Project starts from scratch.
- Initial physical hardware is a single `Roomba i3`.
- Early development should not depend on having many physical robots.
- First playable version only needs to support one game at a time.
- No authentication is required for v1.
- Manual calibration and recovery workflows are acceptable in v1.
- Simulated robots are a first-class requirement.
- Camera tracking is optional in the first pass, but the architecture should leave room for it.

## Success criteria by milestone

- **Simulation MVP**: a user can create a game in the web UI, submit valid checkers moves, and watch simulated robots move between board squares.
- **Single robot physical MVP**: the server can communicate with one real robot, command movement to known square positions, and reflect robot status in the UI.
- **Calibration MVP**: the operator can define board coordinates, assign a robot to a square, and recover from drift or manual movement.
- **Playable physical demo**: at least a limited physical checkers scenario works reliably enough for a demo.

---

## Phase 0 - Project setup and feasibility

### Goals

- Establish the architecture and delivery plan.
- Reduce early hardware uncertainty.
- Create the repo structure and core docs.

### Software tasks

- [ ] Create monorepo structure for `apps`, `packages`, and `docs`.
- [ ] Choose primary stack for server and UI.
  - Recommended default: `React + TypeScript` for UI.
  - Recommended default: `TypeScript` or `Python` for server.
- [ ] Write an architecture overview document.
- [ ] Define the main system boundaries:
  - coordination server
  - web UI
  - robot adapter/service
  - simulated robot service
  - optional camera tracker
- [ ] Define a shared robot protocol draft.
- [ ] Define checkers game state data structures.
- [ ] Define a board coordinate system in physical units, preferably millimeters.
- [ ] Define a basic move execution state machine.
- [ ] Decide on a package/module naming scheme.
- [ ] Set up formatting, linting, and editor config.
- [ ] Add a basic README with project vision and quick start expectations.

### Real-world tasks

- [ ] Measure the `Roomba i3` physical footprint.
- [ ] Estimate minimum square size for a physical checkerboard.
- [ ] Sketch possible board layouts, including off-board parking/capture zones.
- [ ] Investigate `Roomba i3` control options:
  - stock API or unofficial API
  - firmware exploit possibilities
  - serial/debug interfaces
  - external controller retrofit possibilities
- [ ] Record findings in `docs/hardware/roomba-i3-feasibility.md`.
- [ ] Decide whether the first physical prototype uses:
  - the `Roomba i3` directly, or
  - a generic external control module attached to the robot
- [ ] Identify what tools and test space are available for physical experiments.
- [ ] Decide whether an overhead camera is practical in the intended space.

### Exit criteria

- Repo structure exists.
- Architecture and protocol drafts exist.
- `Roomba i3` feasibility is documented.
- Physical board scale assumptions are recorded.

---

## Phase 1 - Shared models, rules engine, and protocol

### Goals

- Create the core abstractions before building UI or robot control.
- Make simulation and real robots share the same interface.

### Software tasks

- [ ] Create a shared `protocol` package/module.
- [ ] Define message schemas for server-to-robot communication:
  - `identify`
  - `stop`
  - `move_to_square`
  - `move_to_pose`
  - `set_pose`
  - `ping`
- [ ] Define message schemas for robot-to-server communication:
  - `hello`
  - `heartbeat`
  - `telemetry`
  - `pose_update`
  - `command_result`
  - `error`
- [ ] Define shared types for:
  - `Robot`
  - `Pose`
  - `Game`
  - `BoardCalibration`
  - `Command`
  - `PieceAssignment`
- [ ] Create a standalone checkers rules engine package/module.
- [ ] Implement:
  - `newGame`
  - `getLegalMoves`
  - `applyMove`
  - `isGameOver`
- [ ] Add unit tests for checkers rules.
- [ ] Decide how to represent kings and captures in the logical model.
- [ ] Create a board model package/module for mapping squares to physical coordinates.
- [ ] Define confidence/source metadata for robot pose estimates.

### Real-world tasks

- [ ] Decide how robots are visually identified on the board.
  - color markers
  - fiducial markers
  - LEDs or labels
- [ ] Decide how a piece's side and king status will be represented physically.
- [ ] Decide whether captured robots move to a parking area or are manually removed in early demos.
- [ ] Decide whether robots are allowed to traverse only playable squares or the full board.

### Exit criteria

- Shared types and protocol are defined.
- Checkers engine is tested.
- Board coordinate model is defined.
- Physical representation assumptions are documented.

---

## Phase 2 - Coordination server skeleton

### Goals

- Build the server as the source of truth for gameplay and robot state.
- Enable basic API and realtime updates.

### Software tasks

- [ ] Create the coordination server app.
- [ ] Add HTTP endpoints for:
  - `POST /games`
  - `GET /games/:id`
  - `POST /games/:id/start`
  - `POST /games/:id/moves`
  - `POST /robots/:id/set-pose`
  - `POST /recovery/stop-all`
- [ ] Add WebSocket support for realtime updates.
- [ ] Implement in-memory stores for:
  - active game
  - robot registry
  - command queue
  - calibration state
- [ ] Implement game lifecycle state machine:
  - `setup`
  - `ready`
  - `executing_move`
  - `awaiting_recovery`
  - `paused`
  - `finished`
- [ ] Implement robot lifecycle state machine:
  - `offline`
  - `pairing`
  - `idle`
  - `moving`
  - `blocked`
  - `error`
  - `manual_control`
- [ ] Add structured logging for commands, telemetry, and failures.
- [ ] Implement a simple event stream for UI consumption.
- [ ] Add API tests for the game lifecycle.

### Real-world tasks

- [ ] Decide where the server will run during physical tests.
  - laptop on local network
  - mini PC near board
- [ ] Decide what local network setup will be used for robot communication.
- [ ] Decide whether robot pairing will use Wi-Fi, BLE, or a temporary USB setup for the first prototype.

### Exit criteria

- Server can create/start a game.
- Server can accept a move request and validate it logically.
- Server can publish game/robot updates over WebSocket.

---

## Phase 3 - Simulation MVP

### Goals

- Build a full playable software loop without depending on real hardware.
- Validate protocols and server behavior early.

### Software tasks

- [ ] Create a simulated robot service/app.
- [ ] Implement robot registration and heartbeat behavior.
- [ ] Implement support for robot commands:
  - `identify`
  - `stop`
  - `move_to_square`
  - `move_to_pose`
  - `set_pose`
- [ ] Simulate continuous movement over time.
- [ ] Simulate status transitions during command execution.
- [ ] Emit telemetry and pose updates on a timer.
- [ ] Add configurable robot parameters:
  - max speed
  - turn rate
  - positional noise
  - artificial latency
  - command failure rate
- [ ] Implement server-side command dispatch to simulated robots.
- [ ] Implement move execution flow:
  - validate move
  - identify robot for moving piece
  - issue command
  - monitor progress
  - commit move when complete
  - surface recovery state on failure
- [ ] Support serialized motion only: one robot moving at a time.
- [ ] Add simulation integration tests.

### Real-world tasks

- [ ] Define what real-world failure modes should be simulated first.
  - timeout
  - stuck robot
  - drift
  - disconnect
- [ ] Decide how closely the simulator should mimic the `Roomba i3` versus a generic robot.

### Exit criteria

- Simulated robots can connect to the server.
- A valid move results in a visible robot movement and a committed board update.
- Failures can push the game into recovery mode.

---

## Phase 4 - Web UI MVP

### Goals

- Provide the user-facing experience for setup, play, status, and recovery.

### Software tasks

- [ ] Create the web app.
- [ ] Build a game dashboard page.
- [ ] Build a board view showing:
  - square grid
  - current pieces
  - robot assignments
  - current robot positions
- [ ] Build controls to:
  - create game
  - start game
  - submit move
  - pause game
  - stop all robots
- [ ] Subscribe to server events over WebSocket.
- [ ] Build a robot list/status panel.
- [ ] Build a move history panel.
- [ ] Show command progress and errors in the UI.
- [ ] Add a setup flow for assigning robots to initial pieces.
- [ ] Add a basic calibration page.
- [ ] Add a manual recovery page with:
  - set robot square
  - retry last move
  - mark robot offline
- [ ] Add client-side tests for core UI logic.

### Real-world tasks

- [ ] Decide whether game operation happens from one operator station or multiple browsers.
- [ ] Decide how much manual control should be exposed in early demos.
- [ ] Draft an operator workflow for setup, play, and recovery.

### Exit criteria

- A user can run a simulated game from the browser.
- Robot statuses and move progress are visible.
- Basic recovery actions are available.

---

## Phase 5 - Calibration and physical board modeling

### Goals

- Make the system aware of real-world board geometry.
- Support manual correction and setup workflows.

### Software tasks

- [ ] Extend the board model to store:
  - board origin
  - square size
  - board rotation
  - optional corner points for perspective transform
- [ ] Implement square-to-world and world-to-square mapping.
- [ ] Implement manual board calibration endpoints.
- [ ] Implement UI tools for:
  - setting board corners or origin
  - setting square size
  - placing robots onto known squares
- [ ] Add tolerance rules for deciding when a robot is "on" a square.
- [ ] Add server-side support for pose source/confidence:
  - manual
  - odometry
  - camera
  - fused
- [ ] Add tests for board coordinate transforms.

### Real-world tasks

- [ ] Build or tape out a prototype physical checkerboard.
- [ ] Measure actual square spacing and clearances.
- [ ] Mark square centers and test whether a robot can reliably stop within tolerances.
- [ ] Design a simple capture/parking area off the main board.
- [ ] Decide whether robots need visual markings for the operator even before camera tracking.
- [ ] Document physical calibration steps.

### Exit criteria

- The server can map board squares to physical coordinates.
- The operator can manually calibrate the board and correct robot positions.
- Physical board dimensions are validated.

---

## Phase 6 - Single real robot integration

### Goals

- Prove end-to-end control of one physical robot.
- Keep the integration behind the shared robot protocol.

### Software tasks

- [ ] Create a real robot adapter/service for the `Roomba i3` or fallback controller.
- [ ] Implement robot registration and heartbeat.
- [ ] Implement `identify` behavior suitable for the physical robot.
- [ ] Implement `stop` behavior with safe interruption.
- [ ] Implement basic movement command execution:
  - move to target square
  - move to target pose if available
- [ ] Ingest telemetry from available hardware sources.
- [ ] Translate hardware-specific state into shared protocol messages.
- [ ] Add retry and timeout handling for flaky communications.
- [ ] Log raw hardware communication for debugging.
- [ ] Add a test harness for isolated robot command testing.

### Real-world tasks

- [ ] Confirm the chosen control method for `Roomba i3`.
- [ ] If necessary, attach an external controller/microcontroller.
- [ ] Determine how the robot receives commands:
  - Wi-Fi
  - BLE
  - USB tether during early testing
- [ ] Determine how the robot reports pose/telemetry.
- [ ] Test start/stop safety behavior.
- [ ] Run repeated square-to-square movement tests on a taped floor grid.
- [ ] Measure drift, overshoot, and turn bias.
- [ ] Record battery/runtime considerations.
- [ ] Document setup procedure for the real robot.

### Exit criteria

- The server can talk to one real robot using the shared protocol.
- The UI can show live status from the real robot.
- The robot can execute simple board-relative movement commands.

---

## Phase 7 - Recovery and operator tooling

### Goals

- Make the system usable despite localization or hardware imperfections.
- Improve debugging and recovery speed.

### Software tasks

- [ ] Add a global emergency stop flow.
- [ ] Add command timeout handling with operator-visible errors.
- [ ] Add stuck/blocked detection rules.
- [ ] Add UI actions for:
  - stop all robots
  - clear error
  - manually set robot pose
  - mark command complete
  - retry command
  - undo logical move if needed
- [ ] Add server-side event history for debugging.
- [ ] Add alert banners for inconsistent state.
- [ ] Add per-robot diagnostics view.
- [ ] Add support for parking a robot outside the board.

### Real-world tasks

- [ ] Define operator safety procedures for physical tests.
- [ ] Decide how manual repositioning is communicated to the system.
- [ ] Test recovery flows when the robot is:
  - physically bumped
  - blocked by an obstacle
  - moved by hand
  - disconnected mid-command
- [ ] Document common failure modes and recovery procedures.

### Exit criteria

- The operator can recover from common failures without restarting the whole system.
- Recovery flows are documented and usable.

---

## Phase 8 - Camera tracking prototype (optional but likely important)

### Goals

- Improve position awareness beyond robot self-reporting.
- Support future localization fusion.

### Software tasks

- [ ] Create a camera tracking service.
- [ ] Select initial tracking approach:
  - AprilTags / fiducials recommended
  - color markers only if simpler and sufficient
- [ ] Implement board calibration from camera-visible landmarks.
- [ ] Implement robot identification from camera input.
- [ ] Estimate robot center position and heading.
- [ ] Publish camera-derived poses to the server.
- [ ] Add source/confidence handling in localization logic.
- [ ] Add a UI view for camera/tracking health.
- [ ] Add tests for pose ingestion and transform math where practical.

### Real-world tasks

- [ ] Choose a camera and mount position.
- [ ] Test lighting conditions in the intended play area.
- [ ] Attach visible markers to the robot.
- [ ] Attach board corner markers if needed.
- [ ] Measure camera field of view and distortion constraints.
- [ ] Evaluate whether camera tracking is reliable enough for square-level accuracy.

### Exit criteria

- The camera service can identify a robot on the board.
- The server can ingest camera-based poses.
- The system can display camera-derived position in the UI.

---

## Phase 9 - Physical gameplay prototype

### Goals

- Move from isolated robot control to actual game-like physical behavior.
- Support a limited but real demo scenario.

### Software tasks

- [ ] Support initial physical piece assignment workflow.
- [ ] Implement captured-piece handling policy.
- [ ] Implement promotion to king in both logic and display.
- [ ] Add physical move completion checks based on tolerance/confidence.
- [ ] Add optional pre-move and post-move verification logic.
- [ ] Improve planner to account for occupied squares and parking zones.
- [ ] Keep serialized motion as the default unless testing proves otherwise.
- [ ] Add demo-specific scenarios with reduced piece counts.

### Real-world tasks

- [ ] Run physical tests with one real robot and simulated peers.
- [ ] Test whether reduced-piece setups are more reliable for demos.
- [ ] Validate physical handling of captures and king promotion.
- [ ] Confirm whether one moving robot at a time is acceptable for the demo experience.
- [ ] Tune speed for visibility, safety, and reliability.

### Exit criteria

- A limited physical checkers scenario is demonstrable.
- Physical move execution is stable enough for repeated demos.

---

## Phase 10 - Multi-robot scaling

### Goals

- Expand from one physical robot to multiple controllable units.
- Begin tackling fleet-level challenges.

### Software tasks

- [ ] Support multiple simultaneous robot registrations.
- [ ] Improve robot assignment management.
- [ ] Add battery/availability awareness if telemetry supports it.
- [ ] Improve planner to schedule non-conflicting robot actions.
- [ ] Add richer occupancy modeling for the board and parking zones.
- [ ] Add fleet diagnostics and health summary views.
- [ ] Introduce scenario-based tests for multi-robot coordination.

### Real-world tasks

- [ ] Acquire or prepare additional robots.
- [ ] Standardize any external controller setup across robots.
- [ ] Assign stable visual IDs to all robots.
- [ ] Test Wi-Fi/BLE reliability with multiple robots present.
- [ ] Determine charging/storage workflow for multiple robots.
- [ ] Measure startup/setup time for a full board.

### Exit criteria

- Multiple robots can connect and be managed consistently.
- The system remains usable with more than one real robot.

---

## Phase 11 - Robustness and polish

### Goals

- Make the system more reliable, demo-friendly, and easier to operate.

### Software tasks

- [ ] Persist game and calibration state across server restarts.
- [ ] Add richer audit/event logging.
- [ ] Add UI polish for operator workflows.
- [ ] Add better automated tests around command sequencing and recovery.
- [ ] Add health checks for all services.
- [ ] Add configuration profiles for different robot models.
- [ ] Add packaging/deployment docs for local setup.

### Real-world tasks

- [ ] Create a repeatable demo checklist.
- [ ] Create printed board/marker assets if needed.
- [ ] Document transport/setup requirements for taking the system elsewhere.
- [ ] Validate the system under longer play sessions.

### Exit criteria

- System is easier to set up and operate repeatedly.
- Demo reliability is improved.

---

## Phase 12 - Future extensions

These are intentionally deferred until the core system is stable.

### Candidate software tasks

- [ ] Add support for chess or other board games.
- [ ] Add an Alexa or voice-command integration.
- [ ] Add authentication and multi-user support.
- [ ] Add support for multiple simultaneous games.
- [ ] Add cloud deployment or remote spectating.
- [ ] Add more advanced path planning and multi-robot choreography.
- [ ] Add support for additional robot brands/models through adapters.

### Candidate real-world tasks

- [ ] Evaluate alternate robot platforms if vacuums are too constrained.
- [ ] Build a more polished board enclosure/camera mount.
- [ ] Create a public demo setup.

---

## Cross-cutting backlog themes

These should be revisited throughout development.

### Testing

- [ ] Unit tests for rules engine and board math.
- [ ] Integration tests for server + simulated robots.
- [ ] UI tests for core workflows.
- [ ] Hardware test scripts for repeated move trials.

### Documentation

- [ ] Architecture decision records for major choices.
- [ ] Hardware notes for each robot model.
- [ ] Operator setup and recovery guide.
- [ ] Protocol documentation.

### Safety and reliability

- [ ] Emergency stop procedure.
- [ ] Battery management guidance.
- [ ] Safe-speed defaults.
- [ ] Physical environment checklist.

### Observability

- [ ] Structured logs.
- [ ] Move/command history.
- [ ] Robot telemetry inspection.
- [ ] Error and timeout metrics.

---

## Recommended immediate next tasks

If starting now, do these first:

1. [ ] Create the repo/app/package skeleton.
2. [ ] Choose server stack and UI stack.
3. [ ] Write the shared robot protocol v0.
4. [ ] Implement the checkers rules engine with tests.
5. [ ] Build the coordination server skeleton.
6. [ ] Build one simulated robot service.
7. [ ] Build the basic web UI board view.
8. [ ] Document `Roomba i3` feasibility and control options.
9. [ ] Tape out a rough physical board and validate square sizing.

---

## Notes on prioritization

- Prioritize **simulation-first progress** so hardware uncertainty does not block the whole project.
- Prioritize **manual calibration and recovery** over sophisticated autonomy.
- Prioritize **one reliable robot moving at a time** before attempting coordinated multi-robot motion.
- Treat the `Roomba i3` as an early experiment, not as a guaranteed foundation for all robot support.
